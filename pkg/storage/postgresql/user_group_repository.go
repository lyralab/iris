package postgresql

import (
	"github.com/root-ali/iris/pkg/groups"
	"time"
)

func (s *Storage) AddUserToGroup(userID, groupID string) error {
	result := s.db.Model(&groups.UserGroup{}).
		Create(&groups.UserGroup{UId: userID, GId: groupID, CreatedAt: time.Now(), ModifiedAt: time.Now()})
	if result.Error != nil {
		s.logger.Errorf("Failed to add user group to user_group table: %v", result.Error)
		return result.Error
	}
	return nil
}

func (s *Storage) RemoveUserFromGroup(userID, groupID string) error {
	// ponytail: hard delete. PK is (group_id, user_id); a soft-deleted row blocks re-add.
	result := s.db.Unscoped().Where("user_id = ? AND group_id = ?", userID, groupID).Delete(&groups.UserGroup{})
	if result.Error != nil {
		s.logger.Errorf("Failed to remove user group from user_group table: %v", result.Error)
		return result.Error
	}
	return nil
}

func (s *Storage) FindGroupById(userID string) ([]string, error) {
	var groupIDs []string
	s.logger.Infow("FindGroupById", "user_id", userID)
	result := s.db.Model(&groups.UserGroup{}).Select("group_id").Find(&groupIDs, "user_id = ?", userID)
	if result.Error != nil {
		s.logger.Errorf("Failed to get user groups from user_group table: %v", result.Error)
		return nil, result.Error
	}
	return groupIDs, nil
}

func (s *Storage) FindUsersById(groupID string) ([]string, error) {
	var userIDs []string
	result := s.db.Model(&groups.UserGroup{}).Select("user_id").Find(&userIDs, "group_id = ?", groupID)
	if result.Error != nil {
		s.logger.Errorf("Failed to get user groups from user_group table: %v", result.Error)
		return nil, result.Error
	}
	return userIDs, nil
}
