package service

import (
	"testing"

	"mini-hrm-backend/internal/model"
)

func TestAccumulateSubtreeCounts(t *testing.T) {
	// Direct counts include ACTIVE and RESIGNED records. Sibling branches must
	// remain independent even when the input is not in tree order.
	depts := []*model.Department{
		{ID: "tech-child", ParentID: "tech", Path: "/root/tech/tech-child/", DirectEmployeeCount: 2},
		{ID: "root", Path: "/root/", DirectEmployeeCount: 2},
		{ID: "sales", ParentID: "root", Path: "/root/sales/", DirectEmployeeCount: 1},
		{ID: "tech", ParentID: "root", Path: "/root/tech/", DirectEmployeeCount: 0},
		{ID: "sales-child", ParentID: "sales", Path: "/root/sales/sales-child/", DirectEmployeeCount: 1},
	}
	accumulateSubtreeCounts(depts)
	want := map[string]int{"root": 6, "tech": 2, "tech-child": 2, "sales": 2, "sales-child": 1}
	for _, dept := range depts {
		if dept.EmployeeCount != want[dept.ID] {
			t.Errorf("%s: got %d, want %d", dept.ID, dept.EmployeeCount, want[dept.ID])
		}
	}
	// Repeating an enrichment should never double count descendants.
	accumulateSubtreeCounts(depts)
	if depts[1].EmployeeCount != 6 {
		t.Errorf("root doubled after refresh: %d", depts[1].EmployeeCount)
	}
}
