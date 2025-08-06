package postgresql

import (
	"errors"

	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
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

func (s *Storage) UpdateUserData(u *user.User) error {
	tmpUser := &user.User{}
	s.logger.Infow("updating user data", "user_id", u.ID, "username", u.UserName)
	if u.UserName == "" {
		s.logger.Error("UserName is required for updating user data")
		return errors.New("UserName is required")
	}
	if u.Mobile != "" {
		tmpUser.Mobile = u.Mobile
		s.logger.Infow("updating user mobile", "user_id", u.ID, "mobile", u.Mobile)
	}
	if u.Email != "" {
		tmpUser.Email = u.Email
		s.logger.Infow("updating user email", "user_id", u.ID, "email", u.Email)
	}
	if u.FirstName != "" {
		tmpUser.FirstName = u.FirstName
		s.logger.Infow("updating user first name", "user_id", u.ID, "first_name", u.FirstName)
	}
	if u.LastName != "" {
		tmpUser.LastName = u.LastName
		s.logger.Infow("updating user last name", "user_id", u.ID, "last_name", u.LastName)
	}
	if u.Password != "" {
		s.logger.Infow("updating user password", "user_id", u.ID)
		tmpUser.Password = u.Password
	} else {
		s.logger.Infow("not updating user password", "user_id", u.ID)
	}

	result := s.db.Model(&user.User{}).Where("user_name = ?", u.UserName).Updates(tmpUser)
	if result.Error != nil {
		s.logger.Error("Error updating user data", zap.Error(result.Error))
		return result.Error
	}
	return nil
}
