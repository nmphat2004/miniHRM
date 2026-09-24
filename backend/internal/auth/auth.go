package auth

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"mini-hrm-backend/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("tài khoản hoặc mật khẩu không chính xác")
	ErrTokenExpired       = errors.New("phiên làm việc đã hết hạn")
	ErrUnauthorized       = errors.New("chưa xác thực người dùng")
)

// AuthService định nghĩa cổng xác thực tách biệt (Interface Segregation)
type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (string, *model.UserClaims, error)
	VerifyToken(ctx context.Context, tokenString string) (*model.UserClaims, error)
}

// CognitoAuthService implements the public Cognito User Pool auth flow and verifies
// signed ID tokens against the pool's rotating JWKS keys.
type CognitoAuthService struct {
	region, poolID, clientID string
	httpClient               *http.Client
	keyMu                    sync.RWMutex
	keys                     map[string]*rsa.PublicKey
}

func NewCognitoAuthService(region, poolID, clientID string) *CognitoAuthService {
	return &CognitoAuthService{region: region, poolID: poolID, clientID: clientID, httpClient: &http.Client{Timeout: 8 * time.Second}, keys: make(map[string]*rsa.PublicKey)}
}

func (s *CognitoAuthService) Authenticate(ctx context.Context, username, password string) (string, *model.UserClaims, error) {
	body, _ := json.Marshal(map[string]interface{}{"AuthFlow": "USER_PASSWORD_AUTH", "ClientId": s.clientID, "AuthParameters": map[string]string{"USERNAME": username, "PASSWORD": password}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://cognito-idp."+s.region+".amazonaws.com/", bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "AWSCognitoIdentityProviderService.InitiateAuth")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	var payload struct {
		AuthenticationResult struct {
			IDToken string `json:"IdToken"`
		} `json:"AuthenticationResult"`
	}
	if resp.StatusCode != http.StatusOK {
		return "", nil, ErrInvalidCredentials
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.AuthenticationResult.IDToken == "" {
		return "", nil, ErrInvalidCredentials
	}
	claims, err := s.VerifyToken(ctx, payload.AuthenticationResult.IDToken)
	return payload.AuthenticationResult.IDToken, claims, err
}

func (s *CognitoAuthService) VerifyToken(ctx context.Context, tokenString string) (*model.UserClaims, error) {
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", s.region, s.poolID)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "RS256" {
			return nil, ErrUnauthorized
		}
		kid, _ := token.Header["kid"].(string)
		key, keyErr := s.publicKey(ctx, kid)
		return key, keyErr
	}, jwt.WithIssuer(issuer), jwt.WithValidMethods([]string{"RS256"}))
	if err != nil || token == nil || !token.Valid {
		return nil, ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["token_use"] != "id" || claims["aud"] != s.clientID {
		return nil, ErrUnauthorized
	}
	userID, _ := claims["custom:employeeId"].(string)
	if userID == "" {
		userID, _ = claims["sub"].(string)
	}
	username, _ := claims["email"].(string)
	role, _ := claims["custom:role"].(string)
	departmentID, _ := claims["custom:departmentId"].(string)
	if userID == "" || (role != string(model.RoleAdmin) && role != string(model.RoleManager) && role != string(model.RoleEmployee)) {
		return nil, ErrUnauthorized
	}
	return &model.UserClaims{UserID: userID, Username: username, Role: model.UserRole(role), DepartmentID: departmentID}, nil
}

func (s *CognitoAuthService) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	s.keyMu.RLock()
	key := s.keys[kid]
	s.keyMu.RUnlock()
	if key != nil {
		return key, nil
	}
	url := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", s.region, s.poolID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var set struct {
		Keys []struct{ Kid, Kty, N, E string } `json:"keys"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&set) != nil {
		return nil, ErrUnauthorized
	}
	parsed := make(map[string]*rsa.PublicKey)
	for _, jwk := range set.Keys {
		if jwk.Kty != "RSA" {
			continue
		}
		nBytes, nErr := base64.RawURLEncoding.DecodeString(jwk.N)
		eBytes, eErr := base64.RawURLEncoding.DecodeString(jwk.E)
		if nErr != nil || eErr != nil || len(eBytes) == 0 {
			continue
		}
		exponent := 0
		for _, b := range eBytes {
			exponent = exponent<<8 | int(b)
		}
		parsed[jwk.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}
	}
	s.keyMu.Lock()
	s.keys = parsed
	key = s.keys[kid]
	s.keyMu.Unlock()
	if key == nil {
		return nil, ErrUnauthorized
	}
	return key, nil
}

// MockAuthService dùng cho môi trường Test và Local Development không cần mạng
type MockAuthService struct {
	jwtSecret []byte
	users     map[string]mockUser
}

type mockUser struct {
	password string
	claims   model.UserClaims
}

func NewMockAuthService(secret string) *MockAuthService {
	svc := &MockAuthService{
		jwtSecret: []byte(secret),
		users:     make(map[string]mockUser),
	}

	// Tài khoản mặc định cho dữ liệu mẫu
	svc.users["admin@minihrm.local"] = mockUser{
		password: "Admin@123",
		claims: model.UserClaims{
			UserID:   "emp-01",
			Username: "Trần Anh Khoa",
			Role:     model.RoleAdmin,
		},
	}
	svc.users["hung.nv@minihrm.local"] = mockUser{
		password: "Manager@123",
		claims: model.UserClaims{
			UserID:       "emp-03",
			Username:     "Nguyễn Văn Hùng",
			Role:         model.RoleManager,
			DepartmentID: "dept-be",
		},
	}
	svc.users["mai.lt@minihrm.local"] = mockUser{
		password: "Manager@123",
		claims: model.UserClaims{
			UserID:       "emp-04",
			Username:     "Lê Thị Mai",
			Role:         model.RoleManager,
			DepartmentID: "dept-biz",
		},
	}
	svc.users["tuan.hm@minihrm.local"] = mockUser{
		password: "User@123",
		claims: model.UserClaims{
			UserID:       "emp-07",
			Username:     "Hoàng Minh Tuấn",
			Role:         model.RoleEmployee,
			DepartmentID: "dept-be",
		},
	}

	return svc
}

func (s *MockAuthService) Authenticate(ctx context.Context, username, password string) (string, *model.UserClaims, error) {
	u, ok := s.users[username]
	if !ok || u.password != password {
		return "", nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":       u.claims.UserID,
		"username":     u.claims.Username,
		"role":         string(u.claims.Role),
		"departmentId": u.claims.DepartmentID,
		"exp":          time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &u.claims, nil
}

func (s *MockAuthService) VerifyToken(ctx context.Context, tokenString string) (*model.UserClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	return &model.UserClaims{
		UserID:       claims["userId"].(string),
		Username:     claims["username"].(string),
		Role:         model.UserRole(claims["role"].(string)),
		DepartmentID: claims["departmentId"].(string),
	}, nil
}
