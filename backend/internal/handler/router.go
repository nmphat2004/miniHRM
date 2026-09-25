package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"mini-hrm-backend/internal/auth"
	"mini-hrm-backend/internal/avatar"
	"mini-hrm-backend/internal/model"
	"mini-hrm-backend/internal/repository"
	"mini-hrm-backend/internal/service"
)

type contextKey string

const actorContextKey contextKey = "actor_claims"

type Server struct {
	authSvc     auth.AuthService
	deptSvc     *service.DepartmentService
	empSvc      *service.EmployeeService
	auditSvc    *service.AuditService
	repo        *repository.DynamoRepository
	avatarStore avatar.Store
	mux         *http.ServeMux
}

func NewServer(
	authSvc auth.AuthService,
	deptSvc *service.DepartmentService,
	empSvc *service.EmployeeService,
	auditSvc *service.AuditService,
	repo *repository.DynamoRepository,
	avatarStores ...avatar.Store,
) *Server {
	s := &Server{
		authSvc:  authSvc,
		deptSvc:  deptSvc,
		empSvc:   empSvc,
		auditSvc: auditSvc,
		repo:     repo,
		mux:      http.NewServeMux(),
	}
	if len(avatarStores) > 0 {
		s.avatarStore = avatarStores[0]
	}
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS Headers cho Next.js dev
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodPost && !strings.HasPrefix(r.URL.Path, "/api/v1/auth/") && r.Header.Get("Idempotency-Key") == "" {
		writeError(w, http.StatusBadRequest, "Mọi yêu cầu POST phải gửi Idempotency-Key")
		return
	}
	if r.Method == http.MethodPost && !strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
		claims, authErr := s.authenticate(r)
		if authErr != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}
		idempotencyKey := claims.UserID + ":" + r.Header.Get("Idempotency-Key")
		r.Header.Set("Idempotency-Key", idempotencyKey)
		if hit, status, body, err := s.repo.CheckIdempotency(r.Context(), idempotencyKey); err != nil {
			writeError(w, http.StatusInternalServerError, "Không thể kiểm tra Idempotency-Key")
			return
		} else if hit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) authenticate(r *http.Request) (*model.UserClaims, error) {
	// 1. Mock header cho kiểm thử nhanh
	if mockRole := r.Header.Get("X-Mock-Role"); os.Getenv("ENABLE_MOCK_HEADERS") == "true" && mockRole != "" {
		claims := &model.UserClaims{
			UserID:         r.Header.Get("X-Mock-User"),
			Username:       "Mock User",
			Role:           model.UserRole(mockRole),
			DepartmentID:   r.Header.Get("X-Mock-Dept"),
			DepartmentPath: r.Header.Get("X-Mock-Dept-Path"),
		}
		if claims.UserID == "" {
			claims.UserID = "emp-mock"
		}
		return claims, nil
	}

	var tokenStr string
	if cookie, err := r.Cookie("session_token"); err == nil {
		tokenStr = cookie.Value
	}

	if tokenStr == "" {
		return nil, auth.ErrUnauthorized
	}

	return s.authSvc.VerifyToken(r.Context(), tokenStr)
}

func (s *Server) registerRoutes() {
	// Health check
	s.mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"system":  "Mini HRM Go API",
			"version": "1.0.0",
		})
	})

	// POST /api/v1/auth/login
	s.mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
			return
		}

		token, claims, err := s.authSvc.Authenticate(r.Context(), req.Username, req.Password)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil || os.Getenv("COOKIE_SECURE") == "true",
			SameSite: sessionCookieSameSite(),
			MaxAge:   24 * 60 * 60,
		})

		writeJSON(w, http.StatusOK, map[string]interface{}{"user": claims})
	})

	// GET /api/v1/auth/me
	s.mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}
		writeJSON(w, http.StatusOK, claims)
	})

	// POST /api/v1/auth/logout
	s.mux.HandleFunc("POST /api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil || os.Getenv("COOKIE_SECURE") == "true",
			SameSite: sessionCookieSameSite(),
			MaxAge:   -1,
		})
		writeJSON(w, http.StatusOK, map[string]string{"message": "Đăng xuất thành công"})
	})

	// GET /api/v1/departments
	s.mux.HandleFunc("GET /api/v1/departments", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		format := r.URL.Query().Get("format")
		if format == "tree" {
			tree, err := s.deptSvc.GetDepartmentTree(r.Context(), claims)
			if err != nil {
				s.handleError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, tree)
			return
		}

		allDepts, err := s.deptSvc.ListDepartmentsWithCounts(r.Context(), claims)
		if err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, allDepts)
	})

	// GET /api/v1/departments/{id}
	s.mux.HandleFunc("GET /api/v1/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		dept, err := s.deptSvc.GetDepartmentByID(r.Context(), claims, id)
		if err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dept)
	})

	// POST /api/v1/departments
	s.mux.HandleFunc("POST /api/v1/departments", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey != "" {
			hit, sc, body, err := s.repo.CheckIdempotency(r.Context(), idempotencyKey)
			if err == nil && hit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(sc)
				w.Write(body)
				return
			}
		}

		var req struct {
			Code     string `json:"code"`
			Name     string `json:"name"`
			ParentID string `json:"parentId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}

		dept, err := s.deptSvc.CreateDepartment(r.Context(), claims, req.Code, req.Name, req.ParentID, idempotencyKey)
		if err != nil {
			if idempotencyKey != "" {
				if hit, status, body, checkErr := s.repo.CheckIdempotency(r.Context(), idempotencyKey); checkErr == nil && hit {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write(body)
					return
				}
			}
			s.handleError(w, err)
			return
		}

		respBytes, _ := json.Marshal(map[string]interface{}{"success": true, "data": dept})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(respBytes)
	})

	// PUT /api/v1/departments/{id}
	s.mux.HandleFunc("PUT /api/v1/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		var req struct {
			Name    string `json:"name"`
			Version int    `json:"version"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}

		dept, err := s.deptSvc.UpdateDepartment(r.Context(), claims, id, req.Name, req.Version)
		if err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dept)
	})

	// POST /api/v1/departments/{id}/move (UC-02)
	s.mux.HandleFunc("POST /api/v1/departments/{id}/move", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		var req struct {
			NewParentID string `json:"newParentId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}

		err = s.deptSvc.MoveDepartment(r.Context(), claims, id, req.NewParentID, r.Header.Get("Idempotency-Key"))
		if err != nil {
			if s.replayIdempotency(w, r) {
				return
			}
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Di chuyển phòng ban thành công"})
	})

	// POST /api/v1/departments/{id}/archive (BR-PB-06)
	s.mux.HandleFunc("POST /api/v1/departments/{id}/archive", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		err = s.deptSvc.ArchiveDepartment(r.Context(), claims, id, r.Header.Get("Idempotency-Key"))
		if err != nil {
			if s.replayIdempotency(w, r) {
				return
			}
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Lưu trữ phòng ban thành công"})
	})

	// POST /api/v1/departments/{id}/manager (BR-PB-08)
	s.mux.HandleFunc("POST /api/v1/departments/{id}/manager", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		var req struct {
			EmployeeID string `json:"employeeId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}

		err = s.deptSvc.AssignManager(r.Context(), claims, id, req.EmployeeID, r.Header.Get("Idempotency-Key"))
		if err != nil {
			if s.replayIdempotency(w, r) {
				return
			}
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Bổ nhiệm Trưởng phòng thành công"})
	})

	// GET /api/v1/employees
	s.mux.HandleFunc("GET /api/v1/employees", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		deptID := r.URL.Query().Get("departmentId")
		status := r.URL.Query().Get("status")
		if status == "ALL" {
			status = ""
		} else if status == "" {
			status = string(model.EmpStatusActive)
		}
		search := r.URL.Query().Get("search")

		emps, err := s.empSvc.ListEmployees(r.Context(), claims, deptID, status, search, r.URL.Query().Get("scope") == "subtree")
		if err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, emps)
	})

	// GET /api/v1/employees/{id}
	s.mux.HandleFunc("GET /api/v1/employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		emp, err := s.empSvc.GetEmployeeByID(r.Context(), claims, id)
		if err != nil {
			s.handleError(w, err)
			return
		}
		if err := s.attachAvatarURL(r.Context(), emp); err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, emp)
	})

	// PUT /api/v1/employees/{id} — editable profile fields only; code/email are immutable.
	s.mux.HandleFunc("PUT /api/v1/employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}
		var req struct {
			FullName string  `json:"fullName"`
			Title    string  `json:"title"`
			JoinedAt string  `json:"joinedAt"`
			Version  int     `json:"version"`
			Avatar   *string `json:"avatar"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}
		employee, err := s.empSvc.UpdateEmployee(r.Context(), claims, r.PathValue("id"), req.FullName, req.Title, req.JoinedAt, req.Version, req.Avatar)
		if err != nil {
			s.handleError(w, err)
			return
		}
		if err := s.attachAvatarURL(r.Context(), employee); err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, employee)
	})

	// POST /api/v1/employees
	s.mux.HandleFunc("POST /api/v1/employees", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey != "" {
			hit, sc, body, err := s.repo.CheckIdempotency(r.Context(), idempotencyKey)
			if err == nil && hit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(sc)
				w.Write(body)
				return
			}
		}

		var req struct {
			Code         string `json:"code"`
			FullName     string `json:"fullName"`
			Email        string `json:"email"`
			DepartmentID string `json:"departmentId"`
			Title        string `json:"title"`
			JoinedAt     string `json:"joinedAt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}

		emp, err := s.empSvc.CreateEmployee(r.Context(), claims, req.Code, req.FullName, req.Email, req.DepartmentID, req.Title, req.JoinedAt, idempotencyKey)
		if err != nil {
			if idempotencyKey != "" {
				if hit, status, body, checkErr := s.repo.CheckIdempotency(r.Context(), idempotencyKey); checkErr == nil && hit {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write(body)
					return
				}
			}
			s.handleError(w, err)
			return
		}

		respBytes, _ := json.Marshal(map[string]interface{}{"success": true, "data": emp})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(respBytes)
	})

	// POST /api/v1/employees/{id}/transfer (UC-03)
	s.mux.HandleFunc("POST /api/v1/employees/{id}/transfer", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		var req struct {
			NewDepartmentID       string `json:"newDepartmentId"`
			ConfirmManagerRemoval bool   `json:"confirmManagerRemoval"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Dữ liệu JSON không hợp lệ")
			return
		}

		err = s.empSvc.TransferEmployee(r.Context(), claims, id, req.NewDepartmentID, req.ConfirmManagerRemoval, r.Header.Get("Idempotency-Key"))
		if err != nil {
			if s.replayIdempotency(w, r) {
				return
			}
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Chuyển phòng ban cho nhân sự thành công"})
	})

	// POST /api/v1/employees/{id}/resign (UC-04)
	s.mux.HandleFunc("POST /api/v1/employees/{id}/resign", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		id := r.PathValue("id")
		var req struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// reason optional
		}

		err = s.empSvc.ResignEmployee(r.Context(), claims, id, req.Reason, r.Header.Get("Idempotency-Key"))
		if err != nil {
			if s.replayIdempotency(w, r) {
				return
			}
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Cập nhật thôi việc nhân sự thành công"})
	})

	// GET /api/v1/audit-logs
	s.mux.HandleFunc("GET /api/v1/audit-logs", func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.authenticate(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Chưa đăng nhập")
			return
		}

		targetID := r.URL.Query().Get("targetId")
		logs, err := s.auditSvc.ListAuditLogs(r.Context(), claims, targetID)
		if err != nil {
			s.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, logs)
	})
}

func sessionCookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SAME_SITE"))) {
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteStrictMode
	}
}

func (s *Server) attachAvatarURL(ctx context.Context, employee *model.Employee) error {
	if employee.AvatarKey != "" {
		if s.avatarStore == nil {
			return service.ErrAvatarStorageUnavailable
		}
		url, err := s.avatarStore.URL(ctx, employee.AvatarKey)
		if err != nil {
			return err
		}
		employee.Avatar = url
		return nil
	}
	// Keep displaying legacy data URLs while users migrate existing avatars to S3.
	if strings.HasPrefix(employee.LegacyAvatar, "data:image/") {
		employee.Avatar = employee.LegacyAvatar
	}
	return nil
}

func (s *Server) replayIdempotency(w http.ResponseWriter, r *http.Request) bool {
	hit, status, body, err := s.repo.CheckIdempotency(r.Context(), r.Header.Get("Idempotency-Key"))
	if err != nil || !hit {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
	return true
}

func (s *Server) handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Không tìm thấy bản ghi hoặc không có quyền truy cập")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		writeError(w, http.StatusForbidden, "Hành động bị từ chối - không đủ thẩm quyền")
		return
	}
	if errors.Is(err, repository.ErrDuplicateCode) || errors.Is(err, repository.ErrDuplicateEmail) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, repository.ErrVersionConflict) {
		writeError(w, http.StatusConflict, "Dữ liệu đã bị thay đổi bởi người khác. Vui lòng tải lại trang.")
		return
	}
	if errors.Is(err, service.ErrCycleDetected) ||
		errors.Is(err, service.ErrInvalidAvatar) ||
		errors.Is(err, service.ErrMoveTooLarge) ||
		errors.Is(err, service.ErrMaxDepthExceed) ||
		errors.Is(err, service.ErrParentArchived) ||
		errors.Is(err, service.ErrDeptHasActiveNV) ||
		errors.Is(err, service.ErrManagerCannotResign) ||
		errors.Is(err, service.ErrJoinDateTooFar) ||
		errors.Is(err, service.ErrManagerRoleTransfer) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeError(w, http.StatusInternalServerError, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": status < 400,
		"data":    data,
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"message": msg,
		},
	})
}
