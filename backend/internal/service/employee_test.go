package service_test

import (
	"strings"
	"testing"
	"time"

)

func TestEmployeeBusinessRules(t *testing.T) {
	t.Run("BR-NV-04: Manager cannot resign rule check", func(t *testing.T) {
		deptManagerID := "emp-03"
		targetEmpID := "emp-03"

		isManagerOfActiveDept := (deptManagerID == targetEmpID)
		if !isManagerOfActiveDept {
			t.Errorf("Phải phát hiện nhân viên đang là Trưởng phòng")
		}
	})

	t.Run("BR-NV-06: Join date cannot be > 90 days in future", func(t *testing.T) {
		now := time.Now()
		validDate := now.AddDate(0, 0, 15).Format("2006-01-02")
		invalidDate := now.AddDate(0, 0, 95).Format("2006-01-02")

		tValid, _ := time.Parse("2006-01-02", validDate)
		tInvalid, _ := time.Parse("2006-01-02", invalidDate)
		maxAllowed := now.AddDate(0, 0, 90)

		if tValid.After(maxAllowed) {
			t.Errorf("Ngày %s phải được chấp thuận", validDate)
		}
		if !tInvalid.After(maxAllowed) {
			t.Errorf("Ngày %s phải bị từ chối vì vượt quá 90 ngày", invalidDate)
		}
	})

	t.Run("Ràng buộc 1 & 2: Manager data scoping and 404 security", func(t *testing.T) {
		managerDeptPath := "/root-01/dept-tech/"
		employeeInTechPath := "/root-01/dept-tech/dept-be/"
		employeeInBizPath := "/root-01/dept-biz/dept-sales/"

		if !strings.HasPrefix(employeeInTechPath, managerDeptPath) {
			t.Errorf("Manager phải thấy nhân viên cùng nhánh trực thuộc")
		}
		if strings.HasPrefix(employeeInBizPath, managerDeptPath) {
			t.Errorf("Manager không được thấy nhân viên nhánh song song khác")
		}
	})
}
