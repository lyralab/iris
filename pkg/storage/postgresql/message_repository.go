package postgresql

import (
	"context"
	"time"

	"github.com/root-ali/iris/pkg/message"
	"gorm.io/gorm/clause"
)

func (s *Storage) SaveMessage(msg *message.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := s.db.Table("message").
		Create(msg).
		WithContext(ctx)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *Storage) GetMessageByID(id int) (*message.Message, error) {
	var m *message.Message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := s.db.Table("message").
		Where("id = ?", id).
		First(m).
		WithContext(ctx)

	if result.Error != nil {
		return nil, result.Error
	}
	return m, nil
}

func (s *Storage) UpdateMessage(msg *message.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := s.db.Table("message").
		Where("id = ?", msg.Id).
		Updates(message.Message{
			Attempt:     msg.Attempt,
			LastAttempt: time.Now(),
			Status:      msg.Status,
			Response:    msg.Response,
			UpdatedAt:   time.Now(),
		}).
		WithContext(ctx)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) ListMessages() ([]*message.Message, error) {
	var msgs []*message.Message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := s.db.Table("message").
		Find(&msgs).
		WithContext(ctx)

	if result.Error != nil {
		return nil, result.Error
	}
	return msgs, nil
}

// ListMessagesBySender returns the latest messages sent through a provider, newest first.
// users.deleted_at is deliberately ignored so the user name always resolves.
func (s *Storage) ListMessagesBySender(sender string, limit int) ([]message.Sent, error) {
	msgs := make([]message.Sent, 0, limit)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := s.db.WithContext(ctx).
		Table("message m").
		Select(`m.receptor, COALESCE(u.user_name, '') AS user_name, m.status,
			COALESCE(m.response, '') AS response, m.message, m.created_at`).
		Joins("LEFT JOIN users u ON u.id = m.user_id").
		Where("m.sender = ?", sender).
		Order("m.created_at DESC").
		Limit(limit).
		Scan(&msgs)

	if result.Error != nil {
		return nil, result.Error
	}
	return msgs, nil
}

func (s *Storage) ListNotFinishedMessages() ([]message.Message, error) {
	var msgs []message.Message
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	result := s.db.Table("message").
		Where("status = ?", "Sent").
		Find(&msgs).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "NOWAIT",
		}).
		WithContext(ctx)

	if result.Error != nil {
		return nil, result.Error
	}

	return msgs, nil
}
