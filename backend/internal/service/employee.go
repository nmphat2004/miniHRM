package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"mini-hrm-backend/internal/model"
	"mini-hrm-backend/internal/repository"
)

var (
	ErrManagerCannotResign = errors.New("lỗi BR-NV-04: nhân viên đang là Trưởng phòng không thể nghỉ việc trước khi chuyển giao chức vụ")
	ErrJoinDateTooFar      = errors.New("lỗi BR-NV-06: ngày vào làm không được muộn hơn hôm nay quá 90 ngày")
	ErrManagerRoleTransfer = errors.New("lỗi BR-NV-05: nhân viên đang là Trưởng phòng. Cần xác nhận gỡ chức vụ trước khi chuyển phòng")
	ErrInvalidAvatar       = errors.New("ảnh đại diện không hợp lệ: chỉ nhận JPG/PNG dưới 64 KB, tối đa 512×512 px")
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
		TargetName: fullName,
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
	if actor.Role == model.RoleEmployee {
		if id != actor.UserID {
			return nil, repository.ErrNotFound
		}
		return s.repo.GetEmployeeByID(ctx, actor.UserID)
	}
	if actor.Role == model.RoleManager {
		employees, err := s.ListEmployees(ctx, actor, "", "", "", false)
		if err != nil {
			return nil, err
		}
		for _, employee := range employees {
			if employee.ID == id {
				return employee, nil
			}
		}
		return nil, repository.ErrNotFound
	}
	emp, err := s.repo.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Data Scoping & 404 security check (Ràng buộc 1 & 2)
	if actor.Role == model.RoleAdmin {
		return emp, nil
	}

	return emp, nil
}

// ListEmployees lấy danh sách nhân sự có phân quyền Scoping (Ràng buộc 1)
func (s *EmployeeService) ListEmployees(ctx context.Context, actor *model.UserClaims, filterDeptID, filterStatus, search string, includeDescendants bool) ([]*model.Employee, error) {
	if actor.Role == model.RoleEmployee {
		// Employee identity is the employee record ID, not their department ID.
		own, err := s.repo.GetEmployeeByID(ctx, actor.UserID)
		if err != nil {
			return nil, nil
		}
		if filterDeptID != "" && own.DepartmentID != filterDeptID {
			return nil, nil
		}
		if filterStatus != "" && string(own.Status) != filterStatus {
			return nil, nil
		}
		if search != "" && !employeeMatches(own, search) {
			return nil, nil
		}
		return []*model.Employee{own}, nil
	}
	var departments []*model.Department
	var err error
	if actor.Role == model.RoleAdmin {
		departments, err = s.repo.ListDepartmentTree(ctx, "")
	} else if actor.Role == model.RoleManager {
		manager, managerErr := s.repo.GetEmployeeByID(ctx, actor.UserID)
		if managerErr != nil {
			return nil, nil
		}
		root, rootErr := s.repo.GetDepartmentByID(ctx, manager.DepartmentID)
		if rootErr != nil {
			return nil, nil
		}
		departments, err = s.repo.ListDepartmentTree(ctx, root.ID)
		if err == nil {
			departments = append([]*model.Department{root}, departments...)
		}
	}
	if err != nil {
		return nil, err
	}
	var subtreePath string
	if includeDescendants && filterDeptID != "" {
		for _, dept := range departments {
			if dept.ID == filterDeptID {
				subtreePath = dept.Path
				break
			}
		}
		if subtreePath == "" {
			return nil, repository.ErrNotFound
		}
	}
	var result []*model.Employee
	seen := make(map[string]bool)
	for _, dept := range departments {
		if subtreePath != "" && !departmentInSubtree(dept.Path, subtreePath) {
			continue
		}
		if subtreePath == "" && filterDeptID != "" && dept.ID != filterDeptID {
			continue
		}
		employees, queryErr := s.repo.QueryEmployeesByDepartment(ctx, dept.ID)
		if queryErr != nil {
			return nil, queryErr
		}
		for _, employee := range employees {
			if seen[employee.ID] || (filterStatus != "" && string(employee.Status) != filterStatus) || (search != "" && !employeeMatches(employee, search)) {
				continue
			}
			seen[employee.ID] = true
			result = append(result, employee)
		}
	}
	return result, nil
}

func departmentInSubtree(departmentPath, rootPath string) bool {
	return strings.HasPrefix(departmentPath, rootPath)
}

func employeeMatches(employee *model.Employee, search string) bool {
	query := strings.ToLower(strings.TrimSpace(search))
	return strings.Contains(strings.ToLower(employee.FullName), query) || strings.Contains(strings.ToLower(employee.Code), query) || strings.Contains(strings.ToLower(employee.Email), query) || strings.Contains(strings.ToLower(employee.Title), query)
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, actor *model.UserClaims, id, fullName, title, joinedAt string, version int, avatar *string) (*model.Employee, error) {
	if actor.Role != model.RoleAdmin && actor.Role != model.RoleManager {
		return nil, ErrForbidden
	}
	employee, err := s.GetEmployeeByID(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if employee.Version != version {
		return nil, repository.ErrVersionConflict
	}
	fullName = strings.TrimSpace(fullName)
	title = strings.TrimSpace(title)
	if len(fullName) < 2 || len(fullName) > 100 {
		return nil, errors.New("họ và tên nhân viên phải từ 2 đến 100 ký tự")
	}
	if title == "" {
		return nil, errors.New("chức danh không được để trống")
	}
	parsed, err := time.Parse("2006-01-02", joinedAt)
	if err != nil {
		return nil, errors.New("ngày vào làm phải theo định dạng YYYY-MM-DD")
	}
	if parsed.After(time.Now().AddDate(0, 0, 90)) {
		return nil, ErrJoinDateTooFar
	}
	before, after := map[string]interface{}{}, map[string]interface{}{}
	if employee.FullName != fullName {
		before["fullName"], after["fullName"] = employee.FullName, fullName
	}
	if employee.Title != title {
		before["title"], after["title"] = employee.Title, title
	}
	if employee.JoinedAt != joinedAt {
		before["joinedAt"], after["joinedAt"] = employee.JoinedAt, joinedAt
	}
	avatarChanged := avatar != nil && *avatar != employee.Avatar
	if avatarChanged {
		if err := validateAvatar(*avatar); err != nil {
			return nil, err
		}
		// Audit only the presence of the photo, never its binary data.
		before["avatar"], after["avatar"] = employee.Avatar != "", *avatar != ""
	}
	if len(after) == 0 {
		return employee, nil
	}
	action := "employee.updated"
	if avatarChanged && len(after) == 1 {
		if *avatar == "" {
			action = "employee.avatar_removed"
		} else {
			action = "employee.avatar_updated"
		}
	}
	audit := &model.AuditLog{ID: uuid.New().String(), ActorID: actor.UserID, ActorName: actor.Username, Action: action, TargetType: "Employee", TargetID: employee.ID, TargetName: fullName, Before: before, After: after, OccurredAt: time.Now().UTC()}
	employee.FullName, employee.Title, employee.JoinedAt = fullName, title, joinedAt
	if avatarChanged {
		employee.Avatar = *avatar
	}
	if err := s.repo.UpdateEmployeeAtomic(ctx, employee, audit, avatarChanged); err != nil {
		return nil, err
	}
	return employee, nil
}

func validateAvatar(avatar string) error {
	if avatar == "" {
		return nil
	}
	parts := strings.SplitN(avatar, ",", 2)
	if len(parts) != 2 || (parts[0] != "data:image/jpeg;base64" && parts[0] != "data:image/png;base64") || len(parts[1]) > base64.StdEncoding.EncodedLen(64*1024) {
		return ErrInvalidAvatar
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(data) == 0 || len(data) > 64*1024 {
		return ErrInvalidAvatar
	}
	if http.DetectContentType(data) != strings.TrimSuffix(strings.TrimPrefix(parts[0], "data:"), ";base64") {
		return ErrInvalidAvatar
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 512 || config.Height > 512 {
		return ErrInvalidAvatar
	}
	return nil
}

// TransferEmployee thực thi UC-03 (Chuyển phòng ban)
func (s *EmployeeService) TransferEmployee(ctx context.Context, actor *model.UserClaims, empID, newDeptID string, confirmManagerRemoval bool, idempotencyKey string) error {
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
		TargetName: emp.FullName,
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

	return s.repo.TransferEmployeeAtomic(ctx, emp, oldDeptID, clearOldDeptManager, audit, idempotencyKey)
}

// ResignEmployee thực thi UC-04 (Cho thôi việc - xóa mềm)
func (s *EmployeeService) ResignEmployee(ctx context.Context, actor *model.UserClaims, empID, reason, idempotencyKey string) error {
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
			return fmt.Errorf("%w: %s", ErrManagerCannotResign, d.Name)
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
		TargetName: emp.FullName,
		After: map[string]interface{}{
			"status":       model.EmpStatusResigned,
			"resignedAt":   emp.ResignedAt,
			"resignReason": emp.ResignReason,
		},
		OccurredAt: now,
	}

	return s.repo.ResignEmployeeAtomic(ctx, emp, audit, idempotencyKey)
}
