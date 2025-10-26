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
