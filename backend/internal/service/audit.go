package service

import (
	"context"
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
	if actor.Role != model.RoleAdmin {
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

	return logs, nil
}
