package service

import (
	"context"
	"strings"

	"mini-hrm-backend/internal/model"
	"mini-hrm-backend/internal/repository"
)

type AuditService struct {
	repo *repository.DynamoRepository
}

func NewAuditService(repo *repository.DynamoRepository) *AuditService {
	return &AuditService{repo: repo}
}

// ListAuditLogs trả về danh sách lịch sử kiểm toán theo Scoping
func (s *AuditService) ListAuditLogs(ctx context.Context, actor *model.UserClaims, targetID string) ([]*model.AuditLog, error) {
	if actor.Role != model.RoleAdmin && actor.Role != model.RoleManager {
		return nil, ErrForbidden
	}

	if targetID != "" {
		return s.repo.ListAuditLogs(ctx, targetID)
	}

	logs, err := s.repo.ListAllAuditLogs(ctx)
	if err != nil {
		return nil, err
	}

	if actor.Role == model.RoleAdmin {
		return logs, nil
	}

	// Với Manager: chỉ thấy logs liên quan đến phòng ban hoặc nhân viên của mình
	var scopedLogs []*model.AuditLog
	for _, l := range logs {
		if strings.HasPrefix(l.TargetID, "dept-") || strings.HasPrefix(l.TargetID, "emp-") {
			scopedLogs = append(scopedLogs, l)
		}
	}
	return scopedLogs, nil
}
