package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"mini-hrm-backend/internal/auth"
	"mini-hrm-backend/internal/avatar"
	"mini-hrm-backend/internal/handler"
	"mini-hrm-backend/internal/repository"
	"mini-hrm-backend/internal/service"
)

func main() {
	ctx := context.Background()
	loadDotEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-southeast-1"
	}
	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if dynamoEndpoint == "" && !useAWSDatabase() {
		dynamoEndpoint = "http://localhost:8000"
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "mini_hrm_table"
	}

	log.Printf("Khởi chạy Mini HRM Go Server trên port :%s...", port)
	if dynamoEndpoint == "" {
		log.Printf("Kết nối DynamoDB AWS tại region %s (Bảng: %s)", region, tableName)
	} else {
		log.Printf("Kết nối DynamoDB endpoint %s (Bảng: %s)", dynamoEndpoint, tableName)
	}

	repo, err := repository.NewDynamoRepository(ctx, dynamoEndpoint, region, tableName)
	if err != nil {
		log.Printf("Cảnh báo: Chưa khởi tạo được kết nối DynamoDB (%v).", err)
	}

	var avatarStore avatar.Store
	bucket := firstEnv("AVATAR_S3_BUCKET", "S3_BUCKET_NAME", "S3_BUCKET", "AWS_S3_BUCKET")
	if bucket != "" {
		avatarStore, err = avatar.NewS3Store(ctx, region, bucket)
		if err != nil {
			log.Printf("S3 avatar chưa sẵn sàng (%v); API vẫn chạy, thao tác avatar sẽ bị vô hiệu hóa", err)
			avatarStore = nil
		} else {
			log.Printf("Avatar storage đã kết nối tới bucket %s ở region %s", bucket, region)
		}
	} else {
		log.Print("Avatar storage chưa cấu hình S3; các thao tác avatar sẽ bị từ chối")
	}

	var authSvc auth.AuthService = auth.NewMockAuthService("super_secret_local_minihrm_key_32bytes")
	if poolID := os.Getenv("COGNITO_USER_POOL_ID"); poolID != "" {
		clientID := os.Getenv("COGNITO_CLIENT_ID")
		if clientID == "" {
			log.Fatal("COGNITO_CLIENT_ID phải được cấu hình cùng COGNITO_USER_POOL_ID")
		}
		authSvc = auth.NewCognitoAuthService(region, poolID, clientID)
		log.Printf("Sử dụng AWS Cognito User Pool %s tại %s", poolID, region)
	}
	deptSvc := service.NewDepartmentService(repo)
	empSvc := service.NewEmployeeService(repo, avatarStore)
	auditSvc := service.NewAuditService(repo)

	srv := handler.NewServer(authSvc, deptSvc, empSvc, auditSvc, repo, avatarStore)

	fmt.Printf("✓ Máy chủ Mini HRM API sẵn sàng tại: http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("Lỗi máy chủ: %v", err)
	}
}

// AWS is selected explicitly with DYNAMODB_USE_AWS=true or implicitly when an
// AWS profile is configured. Without either, local development keeps using
// DynamoDB Local unless DYNAMODB_ENDPOINT specifies another endpoint.
func useAWSDatabase() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("DYNAMODB_USE_AWS")), "true") || strings.TrimSpace(os.Getenv("AWS_PROFILE")) != ""
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

// loadDotEnv loads non-exported development configuration without overwriting
// environment variables supplied by the process manager or deployment platform.
func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !found || key == "" || os.Getenv(key) != "" {
			continue
		}
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			if decoded, decodeErr := strconv.Unquote(value); decodeErr == nil {
				value = decoded
			}
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
		_ = os.Setenv(key, value)
	}
}
