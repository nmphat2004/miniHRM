package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"mini-hrm-backend/internal/model"
	"mini-hrm-backend/internal/repository"
)

var (
	ErrManagerCannotResign = errors.New("lỗi BR-NV-04: nhân viên đang là Trưởng phòng không thể nghỉ việc trước khi chuyển giao chức vụ")
	ErrJoinDateTooFar       = errors.New("lỗi BR-NV-06: ngày vào làm không được muộn hơn hôm nay quá 90 ngày")
	ErrManagerRoleTransfer = errors.New("lỗi BR-NV-05: nhân viên đang là Trưởng phòng. Cần xác nhận gỡ chức vụ trước khi chuyển phòng")
)

type EmployeeService struct {
	repo *repository.DynamoRepository
}

func NewEmployeeService(repo *repository.DynamoRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

// CreateEmployee tạo nhân viên mới
func (s *EmployeeService) CreateEmployee(ctx context.Context, actor *model.UserClaims, code, fullName, email, departmentID, title, joinedAt string, idempotencyKey string) (*model.Employee, error) {
	if actor.Role != model.RoleAdmin {
		return nil, ErrForbidden
	}

	fullName = strings.TrimSpace(fullName)
	if len(fullName) < 2 || len(fullName) > 100 {
		return nil, errors.New("họ và tên nhân viên phải từ 2 đến 100 ký tự (BR-NV-02)")
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("định dạng email không hợp lệ (BR-NV-02)")
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) < 3 || len(code) > 20 {
		return nil, errors.New("mã nhân viên phải từ 3 đến 20 ký tự (BR-NV-01)")
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("chức danh không được để trống (BR-NV-03)")
	}

	// BR-NV-06: Kiểm tra ngày vào làm không muộn hơn hôm nay quá 90 ngày
	parsedJoinDate, err := time.Parse("2006-01-02", joinedAt)
	if err != nil {
		return nil, errors.New("định dạng ngày vào làm phải là YYYY-MM-DD")
	}
	maxAllowedDate := time.Now().AddDate(0, 0, 90)
	if parsedJoinDate.After(maxAllowedDate) {
		return nil, ErrJoinDateTooFar
	}

	// BR-NV-03: Kiểm tra phòng ban ACTIVE
	dept, err := s.repo.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return nil, errors.New("phòng ban được chỉ định không tồn tại")
	}
	if dept.Status != model.DeptStatusActive {
		return nil, errors.New("không thể thêm nhân viên vào phòng ban ARCHIVED (BR-PB-07)")
	}

	empID := uuid.New().String()
	now := time.Now().UTC()

	emp := &model.Employee{
		ID:             empID,
		Code:           code,
		FullName:       fullName,
		Email:          email,
		DepartmentID:   dept.ID,
		DepartmentName: dept.Name,
		DepartmentPath: dept.Path,
		Title:          title,
		JoinedAt:       joinedAt,
		Status:         model.EmpStatusActive,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "employee.created",
		TargetType: "Employee",
		TargetID:   empID,
		After: map[string]interface{}{
			"code":         code,
			"fullName":     fullName,
			"email":        email,
			"departmentId": dept.ID,
			"title":        title,
			"joinedAt":     joinedAt,
		},
		OccurredAt: now,
	}

	err = s.repo.PutEmployeeAtomic(ctx, emp, audit, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return emp, nil
}

// GetEmployeeByID lấy thông tin nhân viên kèm Data Scoping (Ràng buộc 1, 2)
func (s *EmployeeService) GetEmployeeByID(ctx context.Context, actor *model.UserClaims, id string) (*model.Employee, error) {
	emp, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Data Scoping & 404 security check (Ràng buộc 1 & 2)
	if actor.Role == model.RoleAdmin {
		return emp, nil
	} else if actor.Role == model.RoleManager {
		if !strings.HasPrefix(emp.DepartmentPath, actor.DepartmentPath) {
			return nil, repository.ErrNotFound
		}
	} else {
		if emp.DepartmentID != actor.DepartmentID {
			return nil, repository.ErrNotFound
		}
	}

	return emp, nil
}

// ListEmployees lấy danh sách nhân sự có phân quyền Scoping (Ràng buộc 1)
func (s *EmployeeService) ListEmployees(ctx context.Context, actor *model.UserClaims, filterDeptID, filterStatus string) ([]*model.Employee, error) {
	allEmps, err := s.repo.GetAllEmployees(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.Employee
	for _, e := range allEmps {
		// Scoping
		if actor.Role == model.RoleAdmin {
			// ok
		} else if actor.Role == model.RoleManager {
			if !strings.HasPrefix(e.DepartmentPath, actor.DepartmentPath) {
				continue
			}
		} else {
			if e.DepartmentID != actor.DepartmentID {
				continue
			}
		}

		// Filters
		if filterDeptID != "" && e.DepartmentID != filterDeptID {
			continue
		}
		if filterStatus != "" && string(e.Status) != filterStatus {
			continue
		}

		result = append(result, e)
	}
	return result, nil
}

// TransferEmployee thực thi UC-03 (Chuyển phòng ban)
func (s *EmployeeService) TransferEmployee(ctx context.Context, actor *model.UserClaims, empID, newDeptID string, confirmManagerRemoval bool) error {
	if actor.Role != model.RoleAdmin {
		return ErrForbidden
	}

	emp, err := s.repo.GetEmployeeByID(ctx, empID)
	if err != nil {
		return err
	}

	if emp.Status != model.EmpStatusActive {
		return errors.New("không thể chuyển nhân viên đã nghỉ việc")
	}

	if emp.DepartmentID == newDeptID {
		return errors.New("nhân viên đã thuộc phòng ban này")
	}

	newDept, err := s.repo.GetDepartmentByID(ctx, newDeptID)
	if err != nil {
		return errors.New("phòng ban mới không tồn tại")
	}
	if newDept.Status != model.DeptStatusActive {
		return errors.New("không thể chuyển nhân viên vào phòng ban ARCHIVED (BR-PB-07)")
	}

	oldDeptID := emp.DepartmentID
	oldDept, _ := s.repo.GetDepartmentByID(ctx, oldDeptID)

	clearOldDeptManager := false
	if oldDept != nil && oldDept.ManagerID == emp.ID {
		// BR-NV-05: Đang là Trưởng phòng
		if !confirmManagerRemoval {
			return ErrManagerRoleTransfer
		}
		clearOldDeptManager = true
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "employee.transferred",
		TargetType: "Employee",
		TargetID:   empID,
		Before: map[string]interface{}{
			"departmentId":   oldDeptID,
			"departmentName": emp.DepartmentName,
			"wasManager":     clearOldDeptManager,
		},
		After: map[string]interface{}{
			"departmentId":   newDept.ID,
			"departmentName": newDept.Name,
		},
		OccurredAt: time.Now().UTC(),
	}

	emp.DepartmentID = newDept.ID
	emp.DepartmentName = newDept.Name
	emp.DepartmentPath = newDept.Path

	return s.repo.TransferEmployeeAtomic(ctx, emp, oldDeptID, clearOldDeptManager, audit)
}

// ResignEmployee thực thi UC-04 (Cho thôi việc - xóa mềm)
func (s *EmployeeService) ResignEmployee(ctx context.Context, actor *model.UserClaims, empID, reason string) error {
	if actor.Role != model.RoleAdmin {
		return ErrForbidden
	}

	emp, err := s.repo.GetEmployeeByID(ctx, empID)
	if err != nil {
		return err
	}

	if emp.Status == model.EmpStatusResigned {
		return errors.New("nhân viên này đã nghỉ việc trước đó")
	}

	// BR-NV-04: Chặn nếu đang là Trưởng phòng của bất kỳ phòng ban nào
	allDepts, err := s.repo.GetAllDepartments(ctx)
	if err != nil {
		return err
	}
	for _, d := range allDepts {
		if d.ManagerID == emp.ID && d.Status == model.DeptStatusActive {
			return ErrManagerCannotResign
		}
	}

	now := time.Now().UTC()
	emp.ResignedAt = now.Format("2006-01-02")
	emp.ResignReason = strings.TrimSpace(reason)
	if emp.ResignReason == "" {
		emp.ResignReason = "Nghỉ việc theo nguyện vọng cá nhân"
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "employee.resigned",
		TargetType: "Employee",
		TargetID:   empID,
		After: map[string]interface{}{
			"status":       model.EmpStatusResigned,
			"resignedAt":   emp.ResignedAt,
			"resignReason": emp.ResignReason,
		},
		OccurredAt: now,
	}

	return s.repo.ResignEmployeeAtomic(ctx, emp, audit)
}
