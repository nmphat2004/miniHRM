package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"mini-hrm-backend/internal/model"
	"mini-hrm-backend/internal/repository"
)

func main() {
	ctx := context.Background()
	loadDotEnv(".env")
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
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

	if dynamoEndpoint == "" {
		log.Printf("Đang nạp dữ liệu mẫu tối thiểu vào DynamoDB AWS region %s (bảng %s)...", region, tableName)
	} else {
		log.Printf("Đang nạp dữ liệu mẫu tối thiểu vào %s (bảng %s)...", dynamoEndpoint, tableName)
	}

	repo, err := repository.NewDynamoRepository(ctx, dynamoEndpoint, region, tableName)
	if err != nil {
		log.Fatalf("Lỗi kết nối: %v", err)
	}

	now := time.Now().UTC()

	// 1. Tạo các phòng ban mẫu (3 cấp, 2 nhánh song song, 1 phòng ARCHIVED)
	departments := []*model.Department{
		{
			ID:        "root-01",
			Code:      "MINI_HRM",
			Name:      "Tập Đoàn Mini HRM (Gốc)",
			ParentID:  "",
			Path:      "/root-01/",
			Status:    model.DeptStatusActive,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "dept-tech",
			Code:      "TECH_DIV",
			Name:      "Khối Công Nghệ",
			ParentID:  "root-01",
			Path:      "/root-01/dept-tech/",
			Status:    model.DeptStatusActive,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "dept-be",
			Code:        "BACKEND_DEPT",
			Name:        "Phòng Kỹ Thuật Backend",
			ParentID:    "dept-tech",
			Path:        "/root-01/dept-tech/dept-be/",
			ManagerID:   "emp-03",
			ManagerName: "Nguyễn Văn Hùng",
			Status:      model.DeptStatusActive,
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "dept-biz",
			Code:        "BIZ_DIV",
			Name:        "Khối Kinh Doanh",
			ParentID:    "root-01",
			Path:        "/root-01/dept-biz/",
			ManagerID:   "emp-04",
			ManagerName: "Lê Thị Mai",
			Status:      model.DeptStatusActive,
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:        "dept-sales",
			Code:      "SALES_DEPT",
			Name:      "Phòng Kinh Doanh B2B",
			ParentID:  "dept-biz",
			Path:      "/root-01/dept-biz/dept-sales/",
			Status:    model.DeptStatusActive,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "dept-old",
			Code:      "OLD_PROJECT",
			Name:      "Dự Án Nghiên Cứu Cũ (Đã Lưu Trữ)",
			ParentID:  "root-01",
			Path:      "/root-01/dept-old/",
			Status:    model.DeptStatusArchived,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, d := range departments {
		audit := &model.AuditLog{
			ID:         "audit-dept-" + d.ID,
			ActorID:    "system",
			ActorName:  "Hệ Thống Khởi Tạo",
			Action:     "system.seeded",
			TargetType: "Department",
			TargetID:   d.ID,
			TargetName: d.Name,
			After:      map[string]interface{}{"name": d.Name, "code": d.Code},
			OccurredAt: now,
		}
		if err := repo.PutDepartmentAtomic(ctx, d, audit, "seed-"+d.ID); err != nil {
			log.Printf("Ghi chú phòng ban %s: %v", d.Code, err)
		} else {
			fmt.Printf("✓ Đã nạp phòng ban: %s (%s)\n", d.Name, d.Code)
		}
	}

	// 2. Tạo nhân viên mẫu (1 Admin, 2 Managers ở 2 nhánh, 1 nhân viên thường, 1 đã RESIGNED)
	employees := []*model.Employee{
		{
			ID:             "emp-01",
			Code:           "EMP-001",
			FullName:       "Trần Anh Khoa",
			Email:          "admin@minihrm.local",
			DepartmentID:   "root-01",
			DepartmentName: "Tập Đoàn Mini HRM (Gốc)",
			DepartmentPath: "/root-01/",
			Title:          "Tổng Giám Đốc (Admin)",
			JoinedAt:       "2023-01-15",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-03",
			Code:           "EMP-003",
			FullName:       "Nguyễn Văn Hùng",
			Email:          "hung.nv@minihrm.local",
			DepartmentID:   "dept-be",
			DepartmentName: "Phòng Kỹ Thuật Backend",
			DepartmentPath: "/root-01/dept-tech/dept-be/",
			Title:          "Trưởng Phòng Kỹ Thuật Backend",
			JoinedAt:       "2023-03-01",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-04",
			Code:           "EMP-004",
			FullName:       "Lê Thị Mai",
			Email:          "mai.lt@minihrm.local",
			DepartmentID:   "dept-biz",
			DepartmentName: "Khối Kinh Doanh",
			DepartmentPath: "/root-01/dept-biz/",
			Title:          "Giám Đốc Khối Kinh Doanh",
			JoinedAt:       "2023-04-10",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-07",
			Code:           "EMP-007",
			FullName:       "Hoàng Minh Tuấn",
			Email:          "tuan.hm@minihrm.local",
			DepartmentID:   "dept-be",
			DepartmentName: "Phòng Kỹ Thuật Backend",
			DepartmentPath: "/root-01/dept-tech/dept-be/",
			Title:          "Kỹ Sư Backend",
			JoinedAt:       "2023-08-01",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-08",
			Code:           "EMP-008",
			FullName:       "Vũ Minh Đức",
			Email:          "duc.vm@minihrm.local",
			DepartmentID:   "dept-be",
			DepartmentName: "Phòng Kỹ Thuật Backend",
			DepartmentPath: "/root-01/dept-tech/dept-be/",
			Title:          "Kỹ Sư Backend",
			JoinedAt:       "2024-01-08",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-10",
			Code:           "EMP-010",
			FullName:       "Đỗ Ngọc Lan",
			Email:          "lan.dn@minihrm.local",
			DepartmentID:   "dept-sales",
			DepartmentName: "Phòng Kinh Doanh B2B",
			DepartmentPath: "/root-01/dept-biz/dept-sales/",
			Title:          "Chuyên Viên Kinh Doanh",
			JoinedAt:       "2024-02-12",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-11",
			Code:           "EMP-011",
			FullName:       "Ngô Gia Hân",
			Email:          "han.ng@minihrm.local",
			DepartmentID:   "dept-sales",
			DepartmentName: "Phòng Kinh Doanh B2B",
			DepartmentPath: "/root-01/dept-biz/dept-sales/",
			Title:          "Chuyên Viên Khách Hàng",
			JoinedAt:       "2024-03-04",
			Status:         model.EmpStatusActive,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "emp-09",
			Code:           "EMP-009",
			FullName:       "Phạm Quốc Bảo",
			Email:          "bao.pq@minihrm.local",
			DepartmentID:   "dept-sales",
			DepartmentName: "Phòng Kinh Doanh B2B",
			DepartmentPath: "/root-01/dept-biz/dept-sales/",
			Title:          "Chuyên Viên Sales",
			JoinedAt:       "2022-05-20",
			Status:         model.EmpStatusResigned,
			Version:        1,
			ResignedAt:     "2024-02-15",
			ResignReason:   "Chuyển định cư nước ngoài",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	for _, e := range employees {
		audit := &model.AuditLog{
			ID:         "audit-emp-" + e.ID,
			ActorID:    "system",
			ActorName:  "Hệ Thống Khởi Tạo",
			Action:     "system.seeded",
			TargetType: "Employee",
			TargetID:   e.ID,
			TargetName: e.FullName,
			After:      map[string]interface{}{"fullName": e.FullName, "email": e.Email},
			OccurredAt: now,
		}
		if err := repo.PutEmployeeAtomic(ctx, e, audit, "seed-"+e.ID); err != nil {
			log.Printf("Ghi chú nhân viên %s: %v", e.FullName, err)
		} else {
			fmt.Printf("✓ Đã nạp nhân sự: %s - %s (%s)\n", e.FullName, e.Title, e.Status)
		}
	}

	fmt.Println("\n========================================================")
	fmt.Println("✓ Hoàn tất nạp dữ liệu mẫu đáp ứng toàn bộ các tiêu chí:")
	fmt.Println("  1. Cấu trúc 3 cấp, 2 nhánh song song (Tech & Biz)")
	fmt.Println("  2. 2 Managers ở 2 nhánh: Hùng (Tech) & Mai (Biz)")
	fmt.Println("  3. 1 Quản trị viên (Admin): Khoa")
	fmt.Println("  4. 1 Nhân viên thường: Tuấn")
	fmt.Println("  5. 1 Phòng ban ARCHIVED: OLD_PROJECT")
	fmt.Println("  6. 1 Nhân sự RESIGNED: Bảo (xóa mềm)")
	fmt.Println("========================================================")
}

func useAWSDatabase() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("DYNAMODB_USE_AWS")), "true") || strings.TrimSpace(os.Getenv("AWS_PROFILE")) != ""
}

// loadDotEnv makes the seed command use the same backend/.env configuration
// as the server without overriding variables already set by the shell.
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
