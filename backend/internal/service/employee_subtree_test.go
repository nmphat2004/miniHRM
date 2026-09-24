package service

import "testing"

func TestDepartmentInSubtreeExcludesSiblingWithSimilarName(t *testing.T) {
	root := "/company/tech/"
	for _, tc := range []struct {
		path string
		want bool
	}{
		{root, true},
		{"/company/tech/backend/", true},
		{"/company/tech/backend/api/", true},
		{"/company/tech-other/", false},
		{"/company/sales/", false},
	} {
		if got := departmentInSubtree(tc.path, root); got != tc.want {
			t.Errorf("departmentInSubtree(%q, %q) = %v, want %v", tc.path, root, got, tc.want)
		}
	}
}
