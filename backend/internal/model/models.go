package model

import "time"

type DepartmentStatus string
const (
	DeptStatusActive   DepartmentStatus = "ACTIVE"
	DeptStatusArchived DepartmentStatus = "ARCHIVED"
	DeptStatusMoving   DepartmentStatus = "MOVING"
)

type EmployeeStatus string
const (
	EmpStatusActive   EmployeeStatus = "ACTIVE"
	EmpStatusResigned EmployeeStatus = "RESIGNED"
)

type UserRole string
const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleEmployee UserRole = "employee"
)

// Department đại diện cho thực thể Phòng ban
type Department struct {
	ID            string           `json:"id" dynamodbav:"id"`
	Code          string           `json:"code" dynamodbav:"code"`
	Name          string           `json:"name" dynamodbav:"name"`
	ParentID      string           `json:"parentId" dynamodbav:"parentId"`
	Path          string           `json:"path" dynamodbav:"path"`
	ManagerID     string           `json:"managerId,omitempty" dynamodbav:"managerId,omitempty"`
	ManagerName   string           `json:"managerName,omitempty" dynamodbav:"managerName,omitempty"`
	Status        DepartmentStatus `json:"status" dynamodbav:"status"`
	Version       int              `json:"version" dynamodbav:"version"`
	EmployeeCount int              `json:"employeeCount" dynamodbav:"employeeCount"`
	CreatedAt     time.Time        `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt" dynamodbav:"updatedAt"`
	Children      []*Department    `json:"children,omitempty" dynamodbav:"-"`
}

// Employee đại diện cho thực thể Nhân viên
type Employee struct {
	ID             string         `json:"id" dynamodbav:"id"`
	Code           string         `json:"code" dynamodbav:"code"`
	FullName       string         `json:"fullName" dynamodbav:"fullName"`
	Email          string         `json:"email" dynamodbav:"email"`
	DepartmentID   string         `json:"departmentId" dynamodbav:"departmentId"`
	DepartmentName string         `json:"departmentName,omitempty" dynamodbav:"departmentName,omitempty"`
	DepartmentPath string         `json:"departmentPath,omitempty" dynamodbav:"departmentPath,omitempty"`
	Title          string         `json:"title" dynamodbav:"title"`
	JoinedAt       string         `json:"joinedAt" dynamodbav:"joinedAt"`
	Status         EmployeeStatus `json:"status" dynamodbav:"status"`
	Version        int            `json:"version" dynamodbav:"version"`
	ResignedAt     string         `json:"resignedAt,omitempty" dynamodbav:"resignedAt,omitempty"`
	ResignReason   string         `json:"resignReason,omitempty" dynamodbav:"resignReason,omitempty"`
	CreatedAt      time.Time      `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt" dynamodbav:"updatedAt"`
}

// AuditLog đại diện cho Sổ cái kiểm toán chỉ ghi thêm (Append-only)
type AuditLog struct {
	ID         string                 `json:"id" dynamodbav:"id"`
	ActorID    string                 `json:"actorId" dynamodbav:"actorId"`
	ActorName  string                 `json:"actorName" dynamodbav:"actorName"`
	Action     string                 `json:"action" dynamodbav:"action"`
	TargetType string                 `json:"targetType" dynamodbav:"targetType"`
	TargetID   string                 `json:"targetId" dynamodbav:"targetId"`
	Before     map[string]interface{} `json:"before,omitempty" dynamodbav:"before,omitempty"`
	After      map[string]interface{} `json:"after,omitempty" dynamodbav:"after,omitempty"`
	OccurredAt time.Time              `json:"occurredAt" dynamodbav:"occurredAt"`
}

// UserClaims giải mã từ JWT token
type UserClaims struct {
	UserID         string   `json:"userId"`
	Username       string   `json:"username"`
	Role           UserRole `json:"role"`
	DepartmentID   string   `json:"departmentId,omitempty"`
	DepartmentPath string   `json:"departmentPath,omitempty"`
}
