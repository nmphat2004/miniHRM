package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
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
	ErrMoveTooLarge    = errors.New("cây vượt giới hạn kích thước transaction DynamoDB")
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
	siblings, err := s.repo.QueryDepartmentsByParent(ctx, parentID)
	if err != nil {
		return nil, err
	}
	for _, sibling := range siblings {
		if strings.EqualFold(strings.TrimSpace(sibling.Name), name) {
			return nil, fmt.Errorf("tên phòng ban %q đã tồn tại trong cùng phòng cha", name)
		}
	}
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
		TargetName: name,
		After: map[string]interface{}{
			"code":     code,
			"name":     name,
			"parentId": parentID,
			"path":     pathStr,
		},
		OccurredAt: now,
	}

	err = s.repo.PutDepartmentAtomic(ctx, dept, audit, idempotencyKey)
	if err != nil {
		return nil, err
	}

	return dept, nil
}

// GetDepartmentByID lấy thông tin phòng ban kèm kiểm tra Scoping (Ràng buộc 1, 2)
func (s *DepartmentService) GetDepartmentByID(ctx context.Context, actor *model.UserClaims, id string) (*model.Department, error) {
	if actor.Role == model.RoleAdmin {
		return s.repo.GetDepartmentByID(ctx, id)
	}
	departments, err := s.ListDepartments(ctx, actor)
	if err != nil {
		return nil, err
	}
	for _, department := range departments {
		if department.ID == id {
			return department, nil
		}
	}
	return nil, repository.ErrNotFound
}

// GetDepartmentTree trả về cây phân cấp phòng ban theo phạm vi scoping
func (s *DepartmentService) GetDepartmentTree(ctx context.Context, actor *model.UserClaims) ([]*model.Department, error) {
	scopedDepts, err := s.ListDepartmentsWithCounts(ctx, actor)
	if err != nil {
		return nil, err
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

// ListDepartmentsWithCounts reports all employees (ACTIVE and RESIGNED) in each
// accessible subtree. Stored employeeCount remains the direct ACTIVE count used
// by write paths; the returned objects are separate DynamoDB reads.
func (s *DepartmentService) ListDepartmentsWithCounts(ctx context.Context, actor *model.UserClaims) ([]*model.Department, error) {
	depts, err := s.ListDepartments(ctx, actor)
	if err != nil {
		return nil, err
	}
	if actor.Role == model.RoleEmployee {
		if len(depts) == 0 {
			return depts, nil
		}
		depts[0].DirectEmployeeCount = 1
		depts[0].EmployeeCount = 1
		return depts, nil
	}
	for _, dept := range depts {
		employees, err := s.repo.QueryEmployeesByDepartment(ctx, dept.ID)
		if err != nil {
			return nil, err
		}
		dept.DirectEmployeeCount = len(employees)
	}
	accumulateSubtreeCounts(depts)
	return depts, nil
}

func accumulateSubtreeCounts(depts []*model.Department) {
	byID := make(map[string]*model.Department, len(depts))
	for _, dept := range depts {
		dept.EmployeeCount = dept.DirectEmployeeCount
		byID[dept.ID] = dept
	}
	ordered := append([]*model.Department(nil), depts...)
	sort.Slice(ordered, func(i, j int) bool { return len(ordered[i].Path) > len(ordered[j].Path) })
	for _, dept := range ordered {
		if parent := byID[dept.ParentID]; parent != nil {
			parent.EmployeeCount += dept.EmployeeCount
		}
	}
}

// ListDepartments retrieves only departments inside the caller's current scope.
func (s *DepartmentService) ListDepartments(ctx context.Context, actor *model.UserClaims) ([]*model.Department, error) {
	if actor.Role == model.RoleAdmin {
		return s.repo.ListDepartmentTree(ctx, "")
	}
	deptID := actor.DepartmentID
	if actor.Role == model.RoleManager || actor.Role == model.RoleEmployee {
		currentEmployee, lookupErr := s.repo.GetEmployeeByID(ctx, actor.UserID)
		if lookupErr != nil {
			return nil, repository.ErrNotFound
		}
		deptID = currentEmployee.DepartmentID
	}
	dept, err := s.repo.GetDepartmentByID(ctx, deptID)
	if err != nil {
		return nil, repository.ErrNotFound
	}
	if actor.Role != model.RoleManager {
		return []*model.Department{dept}, nil
	}
	children, err := s.repo.ListDepartmentTree(ctx, dept.ID)
	if err != nil {
		return nil, err
	}
	return append([]*model.Department{dept}, children...), nil
}

// UpdateDepartment cập nhật tên phòng ban kèm OCC Version check (Ràng buộc 5)
func (s *DepartmentService) UpdateDepartment(ctx context.Context, actor *model.UserClaims, id, name string, version int) (*model.Department, error) {
	if actor.Role != model.RoleAdmin && actor.Role != model.RoleManager {
		return nil, ErrForbidden
	}

	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return nil, errors.New("tên phòng ban phải từ 2 đến 100 ký tự (BR-PB-01)")
	}

	var dept *model.Department
	var err error
	if actor.Role == model.RoleAdmin {
		dept, err = s.repo.GetDepartmentByID(ctx, id)
	} else {
		departments, listErr := s.ListDepartments(ctx, actor)
		if listErr != nil {
			return nil, listErr
		}
		for _, candidate := range departments {
			if candidate.ID == id {
				dept = candidate
				break
			}
		}
		if dept == nil {
			return nil, repository.ErrNotFound
		}
	}
	if err != nil {
		return nil, err
	}

	if dept.Version != version {
		return nil, repository.ErrVersionConflict
	}
	if actor.Role == model.RoleManager {
		manager, managerErr := s.repo.GetEmployeeByID(ctx, actor.UserID)
		if managerErr != nil {
			return nil, ErrForbidden
		}
		managerDept, scopeErr := s.repo.GetDepartmentByID(ctx, manager.DepartmentID)
		if scopeErr != nil || !strings.HasPrefix(dept.Path, managerDept.Path) {
			return nil, ErrForbidden
		}
	}
	siblings, err := s.repo.QueryDepartmentsByParent(ctx, dept.ParentID)
	if err != nil {
		return nil, err
	}
	for _, sibling := range siblings {
		if sibling.ID != dept.ID && strings.EqualFold(strings.TrimSpace(sibling.Name), name) {
			return nil, fmt.Errorf("tên phòng ban %q đã tồn tại trong cùng phòng cha", name)
		}
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.updated",
		TargetType: "Department",
		TargetID:   id,
		TargetName: name,
		Before:     map[string]interface{}{"name": dept.Name, "version": dept.Version},
		After:      map[string]interface{}{"name": name, "version": dept.Version + 1},
		OccurredAt: time.Now().UTC(),
	}

	dept.Name = name
	err = s.repo.UpdateDepartmentAtomic(ctx, dept, audit, "")
	if err != nil {
		return nil, err
	}
	return dept, nil
}

// MoveDepartment thực thi UC-02 (Di chuyển phòng ban)
func (s *DepartmentService) MoveDepartment(ctx context.Context, actor *model.UserClaims, deptID, newParentID, idempotencyKey string) error {
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
	siblings, err := s.repo.QueryDepartmentsByParent(ctx, newParentID)
	if err != nil {
		return err
	}
	for _, sibling := range siblings {
		if sibling.ID != deptID && strings.EqualFold(strings.TrimSpace(sibling.Name), strings.TrimSpace(dept.Name)) {
			return fmt.Errorf("không thể di chuyển: tên %q đã có trong phòng cha đích", dept.Name)
		}
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
	subDepts, err := s.repo.ListDepartmentTree(ctx, deptID)
	if err != nil {
		return err
	}

	oldPath := dept.Path
	for _, d := range subDepts {
		if strings.HasPrefix(d.Path, oldPath) {
			subPath := strings.Replace(d.Path, oldPath, newDeptPath, 1)
			depth := strings.Count(subPath, "/") - 1
			if depth > 5 {
				return ErrMaxDepthExceed
			}
			d.Path = subPath
		}
	}

	// Kiểm tra độ sâu của chính phòng ban đang di chuyển
	if (strings.Count(newDeptPath, "/") - 1) > 5 {
		return ErrMaxDepthExceed
	}

	var subEmployees []*model.Employee
	affectedDepts := append([]*model.Department{dept}, subDepts...)
	for _, affectedDept := range affectedDepts {
		employees, queryErr := s.repo.QueryEmployeesByDepartment(ctx, affectedDept.ID)
		if queryErr != nil {
			return queryErr
		}
		for _, e := range employees {
			if strings.HasPrefix(e.DepartmentPath, oldPath) {
				e.DepartmentPath = strings.Replace(e.DepartmentPath, oldPath, newDeptPath, 1)
				subEmployees = append(subEmployees, e)
			}
		}
	}
	if 1+len(subDepts)+len(subEmployees)+1+1 > 100 {
		return fmt.Errorf("%w: cần %d thao tác kể cả nhật ký và Idempotency (giới hạn 100); chưa có dữ liệu nào được đổi", ErrMoveTooLarge, 1+len(subDepts)+len(subEmployees)+1+1)
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.moved",
		TargetType: "Department",
		TargetID:   deptID,
		TargetName: dept.Name,
		Before:     map[string]interface{}{"parentId": dept.ParentID, "path": oldPath},
		After:      map[string]interface{}{"parentId": newParentID, "path": newDeptPath},
		OccurredAt: time.Now().UTC(),
	}

	dept.ParentID = newParentID
	dept.Path = newDeptPath

	return s.repo.MoveDepartmentAtomic(ctx, dept, subDepts, subEmployees, audit, idempotencyKey)
}

// ArchiveDepartment thực thi lưu trữ phòng ban (BR-PB-06)
func (s *DepartmentService) ArchiveDepartment(ctx context.Context, actor *model.UserClaims, deptID, idempotencyKey string) error {
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

	// Check actual records rather than relying on the denormalized count.
	employees, queryErr := s.repo.QueryEmployeesByDepartment(ctx, deptID)
	if queryErr != nil {
		return queryErr
	}
	var blockers []string
	for _, employee := range employees {
		if employee.Status == model.EmpStatusActive {
			blockers = append(blockers, employee.FullName)
		}
	}
	if len(blockers) > 0 {
		return fmt.Errorf("không thể lưu trữ: còn nhân viên ACTIVE: %s", strings.Join(blockers, ", "))
	}

	// BR-PB-06: Chặn nếu còn phòng ban con ACTIVE
	children, err := s.repo.QueryDepartmentsByParent(ctx, deptID)
	if err != nil {
		return err
	}
	var activeChildren []string
	for _, d := range children {
		if d.Status == model.DeptStatusActive {
			activeChildren = append(activeChildren, d.Name)
		}
	}
	if len(activeChildren) > 0 {
		return fmt.Errorf("không thể lưu trữ: còn phòng ban con ACTIVE: %s", strings.Join(activeChildren, ", "))
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.archived",
		TargetType: "Department",
		TargetID:   deptID,
		TargetName: dept.Name,
		After:      map[string]interface{}{"status": model.DeptStatusArchived},
		OccurredAt: time.Now().UTC(),
	}

	return s.repo.ArchiveDepartmentAtomic(ctx, dept, audit, idempotencyKey)
}

// AssignManager bổ nhiệm Trưởng phòng (BR-PB-08)
func (s *DepartmentService) AssignManager(ctx context.Context, actor *model.UserClaims, deptID, empID, idempotencyKey string) error {
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
	if dept.Status != model.DeptStatusActive {
		return errors.New("không thể bổ nhiệm trưởng phòng cho phòng ban không ACTIVE")
	}
	if !strings.HasPrefix(emp.DepartmentPath, dept.Path) {
		return errors.New("trưởng phòng phải thuộc chính phòng ban hoặc cây con của phòng đó (BR-PB-08)")
	}

	audit := &model.AuditLog{
		ID:         uuid.New().String(),
		ActorID:    actor.UserID,
		ActorName:  actor.Username,
		Action:     "department.manager_assigned",
		TargetType: "Department",
		TargetID:   deptID,
		TargetName: dept.Name,
		Before:     map[string]interface{}{"managerId": dept.ManagerID, "managerName": dept.ManagerName},
		After:      map[string]interface{}{"managerId": emp.ID, "managerName": emp.FullName},
		OccurredAt: time.Now().UTC(),
	}

	dept.ManagerID = emp.ID
	dept.ManagerName = emp.FullName
	return s.repo.UpdateDepartmentAtomic(ctx, dept, audit, idempotencyKey)
}
