package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"mini-hrm-backend/internal/auth"
	"mini-hrm-backend/internal/handler"
	"mini-hrm-backend/internal/repository"
	"mini-hrm-backend/internal/service"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if dynamoEndpoint == "" {
		dynamoEndpoint = "http://localhost:8000"
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "mini_hrm_table"
	}

	log.Printf("Khởi chạy Mini HRM Go Server trên port :%s...", port)
	log.Printf("Kết nối DynamoDB tại: %s (Bảng: %s)", dynamoEndpoint, tableName)

	repo, err := repository.NewDynamoRepository(ctx, dynamoEndpoint, "ap-southeast-1", tableName)
	if err != nil {
		log.Printf("Cảnh báo: Chưa kết nối được DynamoDB Local (%v).", err)
	}

	authSvc := auth.NewMockAuthService("super_secret_local_minihrm_key_32bytes")
	deptSvc := service.NewDepartmentService(repo)
	empSvc := service.NewEmployeeService(repo)
	auditSvc := service.NewAuditService(repo)

	srv := handler.NewServer(authSvc, deptSvc, empSvc, auditSvc, repo)

	fmt.Printf("✓ Máy chủ Mini HRM API sẵn sàng tại: http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("Lỗi máy chủ: %v", err)
	}
}
