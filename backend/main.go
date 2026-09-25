package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"mini-hrm-backend/internal/auth"
	"mini-hrm-backend/internal/avatar"
	"mini-hrm-backend/internal/handler"
	"mini-hrm-backend/internal/repository"
	"mini-hrm-backend/internal/service"
)

func main() {
	ctx := context.Background()
	region := envOr("AWS_REGION", "ap-southeast-1")
	tableName := envOr("DYNAMODB_TABLE_NAME", "mini_hrm_table")

	// Lambda uses its execution role through the AWS SDK credential chain.
	// Do not set DYNAMODB_ENDPOINT or AWS_PROFILE in the Lambda environment.
	repo, err := repository.NewDynamoRepository(ctx, os.Getenv("DYNAMODB_ENDPOINT"), region, tableName)
	if err != nil {
		log.Fatalf("không khởi tạo được DynamoDB: %v", err)
	}

	var authSvc auth.AuthService = auth.NewMockAuthService("super_secret_local_minihrm_key_32bytes")
	if poolID := strings.TrimSpace(os.Getenv("COGNITO_USER_POOL_ID")); poolID != "" {
		clientID := strings.TrimSpace(os.Getenv("COGNITO_CLIENT_ID"))
		if clientID == "" {
			log.Fatal("COGNITO_CLIENT_ID phải được cấu hình cùng COGNITO_USER_POOL_ID")
		}
		authSvc = auth.NewCognitoAuthService(region, poolID, clientID)
	} else {
		log.Print("CẢNH BÁO: Lambda đang dùng mock authentication")
	}

	var avatarStore avatar.Store
	if bucket := strings.TrimSpace(os.Getenv("AVATAR_S3_BUCKET")); bucket != "" {
		avatarStore, err = avatar.NewS3Store(ctx, region, bucket)
		if err != nil {
			log.Fatalf("không khởi tạo được S3 avatar store: %v", err)
		}
	}

	app := handler.NewServer(
		authSvc,
		service.NewDepartmentService(repo),
		service.NewEmployeeService(repo, avatarStore),
		service.NewAuditService(repo),
		repo,
		avatarStore,
	)
	lambda.Start(httpadapter.NewV2(app).ProxyWithContext)
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
