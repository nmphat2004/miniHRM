package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"mini-hrm-backend/internal/model"
	"mini-hrm-backend/internal/repository"
)

var (
	ErrCycleDetected   = errors.New("lỗi BR-PB-04: không được di chuyển phòng ban vào chính nó hoặc cây con của nó")
	ErrMaxDepthExceed  = errors.New("lỗi BR-PB-03: cây phòng ban sâu tối đa 5 cấp")
	ErrParentArchived  = errors.New("lỗi BR-PB-07: phòng ban cha đã lưu trữ không được tiếp nhận thêm phòng mới")
	ErrDeptHasActiveNV = errors.New("lỗi BR-PB-06: không lưu trữ phòng ban còn nhân viên hoặc phòng con ACTIVE")
	ErrForbidden       = errors.New("bạn không có quyền thực hiện hành động này")
)

var deptCodeRegex = regexp.MustCompile("^[A-Z0-9_]{2,20}$")

type DepartmentService struct {
	repo *repository.DynamoRepository
}

func NewDepartmentService(repo *repository.DynamoRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

// CreateDepartment thực thi UC-01
func (s *DepartmentService) CreateDepartment(ctx context.Context, actor *model.UserClaims, code, name, parentID string, idempotencyKey string) (*model.Department, error) {
	if actor.Role != model.RoleAdmin {
		return nil, ErrForbidden
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if !deptCodeRegex.MatchString(code) {
		return nil, errors.New("mã phòng ban phải từ 2-20 ký tự, chỉ gồm chữ hoa, số và gạch dưới (BR-PB-02)")
	}

	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return nil, errors.New("tên phòng ban phải từ 2 đến 100 ký tự (BR-PB-01)")
	}

	deptID := uuid.New().String()
	now := time.Now().UTC()

	var pathStr string
	if parentID == "" {
		pathStr = fmt.Sprintf("/%s/", deptID)
	} else {
		parentDept, err := s.repo.GetDepartmentByID(ctx, parentID)
		if err != nil {
			return nil, errors.New("phòng ban cha không tồn tại")
		}
		if parentDept.Status == model.DeptStatusArchived {
			return nil, ErrParentArchived
		}
		pathStr = fmt.Sprintf("%s%s/", parentDept.Path, deptID)
		depth := strings.Count(pathStr, "/") - 1
		if depth > 5 {
			return nil, ErrMaxDepthExceed
		}
	}

	dept := &model.Department{
		ID:            deptID,
		Code:          code,
		Name:          name,
		ParentID:      parentID,
		Path:          pathStr,
		Status:        model.DeptStatusActive,
		Version:       1,
		EmployeeCount: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.created",
		TargetType: "Department",
		TargetID:   deptID,
		After: map[string]interface{}{
			"code":     code,
			"name":     name,
			"parentId": parentID,
			"path":     pathStr,
		},
		OccurredAt: now,
	}

	err := s.repo.PutDepartmentAtomic(ctx, dept, audit, idempotencyKey)
	if err != nil {
		return nil, err
	}

	return dept, nil
}

// GetDepartmentByID lấy thông tin phòng ban kèm kiểm tra Scoping (Ràng buộc 1, 2)
func (s *DepartmentService) GetDepartmentByID(ctx context.Context, actor *model.UserClaims, id string) (*model.Department, error) {
	dept, err := s.repo.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Data Scoping & 404 security check (Ràng buộc 1 & 2)
	if actor.Role == model.RoleAdmin {
		return dept, nil
	} else if actor.Role == model.RoleManager {
		if !strings.HasPrefix(dept.Path, actor.DepartmentPath) {
			return nil, repository.ErrNotFound
		}
	} else {
		if dept.ID != actor.DepartmentID {
			return nil, repository.ErrNotFound
		}
	}

	return dept, nil
}

// GetDepartmentTree trả về cây phân cấp phòng ban theo phạm vi scoping
func (s *DepartmentService) GetDepartmentTree(ctx context.Context, actor *model.UserClaims) ([]*model.Department, error) {
	allDepts, err := s.repo.GetAllDepartments(ctx)
	if err != nil {
		return nil, err
	}

	// Lọc theo Data Scoping (Ràng buộc 1)
	var scopedDepts []*model.Department
	for _, d := range allDepts {
		if actor.Role == model.RoleAdmin {
			scopedDepts = append(scopedDepts, d)
		} else if actor.Role == model.RoleManager {
			if strings.HasPrefix(d.Path, actor.DepartmentPath) {
				scopedDepts = append(scopedDepts, d)
			}
		} else {
			if d.ID == actor.DepartmentID {
				scopedDepts = append(scopedDepts, d)
			}
		}
	}

	// Xây dựng cây phân cấp (Tree structure)
	deptMap := make(map[string]*model.Department)
	for _, d := range scopedDepts {
		d.Children = []*model.Department{}
		deptMap[d.ID] = d
	}

	var rootNodes []*model.Department
	for _, d := range scopedDepts {
		if parent, exists := deptMap[d.ParentID]; exists {
			parent.Children = append(parent.Children, d)
		} else {
			rootNodes = append(rootNodes, d)
		}
	}

	return rootNodes, nil
}

// UpdateDepartment cập nhật tên phòng ban kèm OCC Version check (Ràng buộc 5)
func (s *DepartmentService) UpdateDepartment(ctx context.Context, actor *model.UserClaims, id, name string, version int) (*model.Department, error) {
	if actor.Role != model.RoleAdmin {
		return nil, ErrForbidden
	}

	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return nil, errors.New("tên phòng ban phải từ 2 đến 100 ký tự (BR-PB-01)")
	}

	dept, err := s.repo.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dept.Version != version {
		return nil, repository.ErrVersionConflict
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.updated",
		TargetType: "Department",
		TargetID:   id,
		Before:     map[string]interface{}{"name": dept.Name, "version": dept.Version},
		After:      map[string]interface{}{"name": name, "version": dept.Version + 1},
		OccurredAt: time.Now().UTC(),
	}

	dept.Name = name
	err = s.repo.UpdateDepartmentAtomic(ctx, dept, audit)
	if err != nil {
		return nil, err
	}
	return dept, nil
}

// MoveDepartment thực thi UC-02 (Di chuyển phòng ban)
func (s *DepartmentService) MoveDepartment(ctx context.Context, actor *model.UserClaims, deptID, newParentID string) error {
	if actor.Role != model.RoleAdmin {
		return ErrForbidden
	}

	if deptID == newParentID {
		return ErrCycleDetected
	}

	dept, err := s.repo.GetDepartmentByID(ctx, deptID)
	if err != nil {
		return err
	}

	if dept.Status != model.DeptStatusActive {
		return errors.New("không thể di chuyển phòng ban không ở trạng thái ACTIVE")
	}

	var newParentPath string
	if newParentID != "" {
		newParent, err := s.repo.GetDepartmentByID(ctx, newParentID)
		if err != nil {
			return errors.New("phòng ban cha mới không tồn tại")
		}
		if newParent.Status == model.DeptStatusArchived {
			return ErrParentArchived
		}
		// BR-PB-04: Cycle detection O(1) qua path check
		targetSegment := fmt.Sprintf("/%s/", deptID)
		if strings.Contains(newParent.Path, targetSegment) {
			return ErrCycleDetected
		}
		newParentPath = newParent.Path
	}

	var newDeptPath string
	if newParentID == "" {
		newDeptPath = fmt.Sprintf("/%s/", deptID)
	} else {
		newDeptPath = fmt.Sprintf("%s%s/", newParentPath, deptID)
	}

	// Lấy tất cả phòng ban và nhân sự để cập nhật cây con (BR-PB-05)
	allDepts, err := s.repo.GetAllDepartments(ctx)
	if err != nil {
		return err
	}

	oldPath := dept.Path
	var subDepts []*model.Department
	for _, d := range allDepts {
		if d.ID != deptID && strings.HasPrefix(d.Path, oldPath) {
			subPath := strings.Replace(d.Path, oldPath, newDeptPath, 1)
			depth := strings.Count(subPath, "/") - 1
			if depth > 5 {
				return ErrMaxDepthExceed
			}
			d.Path = subPath
			subDepts = append(subDepts, d)
		}
	}

	// Kiểm tra độ sâu của chính phòng ban đang di chuyển
	if (strings.Count(newDeptPath, "/") - 1) > 5 {
		return ErrMaxDepthExceed
	}

	allEmps, err := s.repo.GetAllEmployees(ctx)
	if err != nil {
		return err
	}

	var subEmployees []*model.Employee
	for _, e := range allEmps {
		if strings.HasPrefix(e.DepartmentPath, oldPath) {
			e.DepartmentPath = strings.Replace(e.DepartmentPath, oldPath, newDeptPath, 1)
			subEmployees = append(subEmployees, e)
		}
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.moved",
		TargetType: "Department",
		TargetID:   deptID,
		Before:     map[string]interface{}{"parentId": dept.ParentID, "path": oldPath},
		After:      map[string]interface{}{"parentId": newParentID, "path": newDeptPath},
		OccurredAt: time.Now().UTC(),
	}

	dept.ParentID = newParentID
	dept.Path = newDeptPath

	return s.repo.MoveDepartmentAtomic(ctx, dept, subDepts, subEmployees, audit)
}

// ArchiveDepartment thực thi lưu trữ phòng ban (BR-PB-06)
func (s *DepartmentService) ArchiveDepartment(ctx context.Context, actor *model.UserClaims, deptID string) error {
	if actor.Role != model.RoleAdmin {
		return ErrForbidden
	}

	dept, err := s.repo.GetDepartmentByID(ctx, deptID)
	if err != nil {
		return err
	}

	if dept.Status != model.DeptStatusActive {
		return errors.New("phòng ban đã được lưu trữ trước đó")
	}

	// BR-PB-06: Chặn nếu còn nhân viên
	if dept.EmployeeCount > 0 {
		return ErrDeptHasActiveNV
	}

	// BR-PB-06: Chặn nếu còn phòng ban con ACTIVE
	allDepts, err := s.repo.GetAllDepartments(ctx)
	if err != nil {
		return err
	}
	for _, d := range allDepts {
		if d.ParentID == deptID && d.Status == model.DeptStatusActive {
			return ErrDeptHasActiveNV
		}
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.archived",
		TargetType: "Department",
		TargetID:   deptID,
		After:      map[string]interface{}{"status": model.DeptStatusArchived},
		OccurredAt: time.Now().UTC(),
	}

	return s.repo.ArchiveDepartmentAtomic(ctx, dept, audit)
}

// AssignManager bổ nhiệm Trưởng phòng (BR-PB-08)
func (s *DepartmentService) AssignManager(ctx context.Context, actor *model.UserClaims, deptID, empID string) error {
	if actor.Role != model.RoleAdmin {
		return ErrForbidden
	}

	dept, err := s.repo.GetDepartmentByID(ctx, deptID)
	if err != nil {
		return err
	}

	emp, err := s.repo.GetEmployeeByID(ctx, empID)
	if err != nil {
		return errors.New("nhân viên không tồn tại")
	}

	if emp.Status != model.EmpStatusActive {
		return errors.New("chỉ có thể bổ nhiệm nhân viên đang làm việc (ACTIVE)")
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.manager_assigned",
		TargetType: "Department",
		TargetID:   deptID,
		Before:     map[string]interface{}{"managerId": dept.ManagerID, "managerName": dept.ManagerName},
		After:      map[string]interface{}{"managerId": emp.ID, "managerName": emp.FullName},
		OccurredAt: time.Now().UTC(),
	}

	dept.ManagerID = emp.ID
	dept.ManagerName = emp.FullName
	return s.repo.UpdateDepartmentAtomic(ctx, dept, audit)
}
