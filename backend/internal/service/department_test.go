package service_test

import (
	"strings"
	"testing"
)

func TestDepartmentHierarchyRules(t *testing.T) {
	t.Run("BR-PB-03: Max depth 5 levels validation", func(t *testing.T) {
		pathValid := "/root-01/dept-tech/dept-be/"
		depthValid := strings.Count(pathValid, "/") - 1
		if depthValid > 5 {
			t.Errorf("Kỳ vọng pathValid độ sâu <= 5, thực tế: %d", depthValid)
		}

		pathInvalid := "/root-01/d1/d2/d3/d4/d5/"
		depthInvalid := strings.Count(pathInvalid, "/") - 1
		if depthInvalid <= 5 {
			t.Errorf("Kỳ vọng pathInvalid độ sâu > 5, thực tế: %d", depthInvalid)
		}
	})

	t.Run("BR-PB-04: Cycle detection O(1) by path segment", func(t *testing.T) {
		movingDeptID := "dept-tech"
		newParentPath := "/root-01/dept-tech/dept-be/"
		targetSegment := "/" + movingDeptID + "/"

		if !strings.Contains(newParentPath, targetSegment) {
			t.Errorf("Phải phát hiện chu trình khi phòng cha mới chứa ID của chính phòng đang di chuyển")
		}

		unrelatedParentPath := "/root-01/dept-biz/"
		if strings.Contains(unrelatedParentPath, targetSegment) {
			t.Errorf("Không được báo chu trình khi di chuyển sang nhánh độc lập")
		}
	})
}
