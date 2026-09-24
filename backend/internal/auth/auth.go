package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"mini-hrm-backend/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("tài khoản hoặc mật khẩu không chính xác")
	ErrTokenExpired        = errors.New("phiên làm việc đã hết hạn")
	ErrUnauthorized        = errors.New("chưa xác thực người dùng")
)

// AuthService định nghĩa cổng xác thực tách biệt (Interface Segregation)
type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (string, *model.UserClaims, error)
	VerifyToken(ctx context.Context, tokenString string) (*model.UserClaims, error)
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
