package postgresql

import (
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Storage) AddUser(user *user.User) error {
	result := s.db.Save(user)
	if result.Error != nil {
		s.logger.Error("Error saving user", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

func (s *Storage) GetUserByID(id string) (*user.User, error) {
	var u *user.User
	result := s.db.First(&u, "id = ?", id)
	if result.Error != nil {
		s.logger.Error("Error getting user", zap.Error(result.Error))
		return nil, result.Error
	}
	return u, nil
}

func (s *Storage) GetUserByUsername(u *user.User) error {
	result := s.db.First(u, "user_name = ?", u.UserName)
	if result.Error != nil {
		s.logger.Error("Error getting user", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

func (s *Storage) GetUserByEmail(email string) (*user.User, error) {
	var u *user.User
	result := s.db.First(&u, "email = ?", email)
	if result.Error != nil {
		s.logger.Error("Error getting user", zap.Error(result.Error))
		return nil, result.Error
	}
	return u, nil
}

func (s *Storage) GetAllUsers() ([]*user.User, error) {
	var users []*user.User
	result := s.db.Find(&users)
	if result.Error != nil {
		s.logger.Error("Error getting users", zap.Error(result.Error))
		return nil, result.Error
	}
	return users, nil
}

func (s *Storage) UpdateUserPassword(_ string, _ string) error {
	return nil
}

func (s *Storage) DeleteUser(_ string) error {
	return nil
}

func (s *Storage) GetRole(u *user.User) error {
	result := s.db.Select("role_id").First(u, "user_name = ?", u.UserName)
	s.logger.Info("user role is ", u.Role)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Storage) VerifyUser(u *user.User) error {
	result := s.db.Model(&user.User{}).Where("user_name = ?", u.UserName).Update("status", u.Status)
	s.logger.Infow("user status is set to ", "status", u.Status, "user_id", u.UserName)

	if result.Error != nil {
		s.logger.Error("Error updating user", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

func nullable(v string) interface{} {
	if v == "" {
		return gorm.Expr("NULL")
	}
	return v
}

func (s *Storage) UpdateUserData(u *user.User) error {
	s.logger.Debugw("updating user data", "user_id", u.ID, "username", u.UserName)
	// Map update writes every contact field, including a cleared one.
	result := s.db.Model(&user.User{}).Where("user_name = ?", u.UserName).Updates(map[string]interface{}{
		"first_name":    u.FirstName,
		"last_name":     u.LastName,
		"email":         nullable(u.Email),
		"mobile":        nullable(u.Mobile),
		"telegram_id":   nullable(u.TelegramID),
		"mattermost_id": nullable(u.MattermostId),
		"bale_id":       nullable(u.BaleID),
		"modified_at":   u.ModifiedAt,
	})
	if result.Error != nil {
		s.logger.Error("Error updating user data", zap.Error(result.Error))
		return result.Error
	}
	return nil
}
