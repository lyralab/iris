package postgresql

import (
	"github.com/root-ali/iris/pkg/user"
)

func (s *Storage) GetMobileNumbersByGroupID(groupID string) ([]string, error) {
	var mobiles []string
	err := s.db.
		Model(&user.User{}).
		Joins("left join user_groups on user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupID).
		Pluck("users.mobile", &mobiles).Error
	if err != nil {
		s.logger.Errorw("error getting mobile numbers", "error", err)
		return nil, err
	}
	return mobiles, nil
}

func (s *Storage) GetEmailsByGroupID(groupID string) ([]string, error) {
	var emails []string
	err := s.db.
		Model(&user.User{}).
		Joins("left join user_groups on user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupID).
		Pluck("users.emails", &emails).Error
	if err != nil {
		s.logger.Errorw("error getting mobile numbers", "error", err)
		return nil, err
	}
	return emails, nil
}
