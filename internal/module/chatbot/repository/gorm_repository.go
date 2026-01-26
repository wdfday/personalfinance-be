package repository

import (
	"context"

	"personalfinancedss/internal/module/chatbot/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// CONVERSATION REPOSITORY
// ============================================

// gormConversationRepository implements domain.ConversationRepository
type gormConversationRepository struct {
	db *gorm.DB
}

// NewGormConversationRepository creates a new GORM conversation repository
func NewGormConversationRepository(db *gorm.DB) domain.ConversationRepository {
	return &gormConversationRepository{db: db}
}

func (r *gormConversationRepository) Create(ctx context.Context, conv *domain.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *gormConversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	var conv domain.Conversation
	err := r.db.WithContext(ctx).Where("id = ? AND status != ?", id, domain.ConversationStatusDeleted).First(&conv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *gormConversationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, status string) ([]domain.Conversation, int64, error) {
	var convs []domain.Conversation
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Conversation{}).Where("user_id = ? AND status != ?", userID, domain.ConversationStatusDeleted)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&convs).Error; err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

func (r *gormConversationRepository) Update(ctx context.Context, conv *domain.Conversation) error {
	return r.db.WithContext(ctx).Save(conv).Error
}

func (r *gormConversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Conversation{}).Where("id = ?", id).Update("status", domain.ConversationStatusDeleted).Error
}

// ============================================
// MESSAGE REPOSITORY
// ============================================

// gormMessageRepository implements domain.MessageRepository
type gormMessageRepository struct {
	db *gorm.DB
}

// NewGormMessageRepository creates a new GORM message repository
func NewGormMessageRepository(db *gorm.DB) domain.MessageRepository {
	return &gormMessageRepository{db: db}
}

func (r *gormMessageRepository) Create(ctx context.Context, msg *domain.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *gormMessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	var msg domain.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&msg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *gormMessageRepository) GetByConversationID(ctx context.Context, convID uuid.UUID, limit int) ([]domain.Message, error) {
	var msgs []domain.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func (r *gormMessageRepository) GetLastMessage(ctx context.Context, convID uuid.UUID) (*domain.Message, error) {
	var msg domain.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *gormMessageRepository) CountByConversationID(ctx context.Context, convID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Message{}).Where("conversation_id = ?", convID).Count(&count).Error
	return count, err
}
