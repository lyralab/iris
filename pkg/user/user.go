package user

import (
	"errors"
	"math/rand"
	"regexp"
	"time"

	"github.com/oklog/ulid/v2"
	iris_error "github.com/root-ali/iris/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewUserService(repo UserInterfaceRepository, role UserRoleRepository, l *zap.SugaredLogger) UserInterfaceService {
	return &userServiceImpl{
		repo:   repo,
		role:   role,
		logger: l,
	}
}

func (us *userServiceImpl) AddUser(u *User) error {
	err := us.repo.GetUserByUsername(u)
	us.logger.Infow("error in getting user", "error", err, "username", u.UserName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		us.logger.Infow("User does not exist, creating new user", "username", u.UserName)
	}
	if err == nil {
		return errors.New("cannot complete signup process")
	}
	us.logger.Infow("About to create a new user with",
		"name", u.UserName, "firstName", u.FirstName, "lastName", u.LastName, "password", u.Password)
	t := time.Now()
	u.CreatedAt = t
	u.ModifiedAt = t
	u.Status = "New"
	roleId, err := us.role.GetRoleIDByName("viewer")
	if err != nil {
		us.logger.Infow("Cannot get role id from database", "error", err)
		return err
	}
	u.Role = roleId
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	u.ID = id.String()
	salt, err := generateSalt()
	if err != nil {
		return err
	}
	u.Salt = salt
	hashedPassword, err := hashPassword(u.Password, salt)
	u.Password = hashedPassword
	if err != nil {
		return err
	}
	err = us.repo.AddUser(u)

	if err != nil {
		us.logger.Errorw("Failed to add user", "user", u.UserName, "error", err)
		return err
	}
	us.logger.Infow("User added successfully", "user", u.UserName)
	return nil
}

func (us *userServiceImpl) GetUserByEmail(email string) (*User, error) {
	u, err := us.repo.GetUserByEmail(email)
	if err != nil {
		us.logger.Errorw("Failed to get user by email", "email", email, "error", err)
		return nil, errors.New(err.Error())
	}
	return u, nil
}

func (us *userServiceImpl) CreateDefaultAdminUser() error {
	u := User{UserName: "admin"}
	err := us.repo.GetUserByUsername(&u)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		t := time.Now()
		u.FirstName = "admin"
		u.LastName = "admin"
		u.CreatedAt = t
		u.ModifiedAt = t
		u.Status = "Verified"
		var generatedPassword string
		for {
			generatedPassword, err = generateRandomPassword(20)
			if err != nil {
				us.logger.Info("cannot generate random password ", "error", err)
				return err
			}
			matched, _ := regexp.MatchString(`[^a-zA-Z0-9]`, generatedPassword)
			if matched {
				us.logger.Info("Generated password contains special characters")
				break
			} else {
				us.logger.Infow("Generated password is not valid, retrying...", "password", generatedPassword)
				continue
			}

		}

		roleId, err := us.role.GetRoleIDByName("admin")
		if err != nil {
			us.logger.Infow("Cannot get role id from database", "error", err)
			return err
		}
		u.Role = roleId
		entropy := rand.New(rand.NewSource(t.UnixNano()))
		id := ulid.MustNew(ulid.Timestamp(t), entropy)

		u.ID = id.String()
		salt, err := generateSalt()
		if err != nil {
			return err
		}
		u.Salt = salt
		hashedPassword, err := hashPassword(generatedPassword, salt)
		u.Password = hashedPassword
		if err != nil {
			return err
		}
		err = us.repo.AddUser(&u)

		if err != nil {
			us.logger.Errorw("Failed to add user", "user", u.UserName, "error", err)
			return err
		}
		us.logger.Infow("User added successfully", "user", u.UserName, "password", generatedPassword)
		return nil
	} else if err == nil {
		us.logger.Infow("Default admin user already exists nothing to do...")
		return nil
	} else {
		return err
	}

}

func (us *userServiceImpl) ValidateUser(u *User) error {
	enteredPassword := u.Password
	err := us.repo.GetUserByUsername(u)
	if err != nil {
		us.logger.Infow("cannot get user from database", "username", u.UserName)
		return err
	}
	if u.Status == "New" {
		us.logger.Infow("user cannot login because not verified yet", u.UserName)
		return iris_error.ErrUserNotVerified
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(enteredPassword+u.Salt))
	if err != nil {
		us.logger.Errorw("cannot validate password", "error", err)
		return iris_error.ErrPasswordNotMatch
	}
	return nil
}

func (us *userServiceImpl) GetUserRole(u *User) error {
	err := us.repo.GetRole(u)
	if err != nil {
		return err
	}
	return nil
}

func (us *userServiceImpl) VerifyUser(u *User) error {
	if u.Status == "Verified" {
		return iris_error.ErrUserAlreadyVerified
	}
	u.Status = "Verified"
	err := us.repo.VerifyUser(u)
	if err != nil {
		return err
	}
	return nil
}

func (us *userServiceImpl) UpdateUser(u *User) error {
	err := us.repo.UpdateUserData(u)
	if err != nil {
		return err
	}
	return nil
}

func (us *userServiceImpl) GetAllUsers() ([]*User, error) {
	users, err := us.repo.GetAllUsers()
	if err != nil {
		us.logger.Errorw("Failed to get all users", "error", err)
		return nil, err
	}
	us.logger.Infow("Successfully retrieved all users", "count", len(users))
	return users, nil
}

func (us *userServiceImpl) GetByUserName(name string) (*User, error) {
	if name == "" {
		return nil, errors.New("username is required")
	}
	user := &User{UserName: name}
	err := us.repo.GetUserByUsername(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
