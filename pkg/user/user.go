package user

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/groups"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func NewUserService(repo UserInterfaceRepository, role UserRoleRepository, cache groups.ReceptorCache, l *zap.SugaredLogger) UserInterfaceService {
	return &userServiceImpl{
		repo:   repo,
		role:   role,
		cache:  cache,
		logger: l,
	}
}

func (us *userServiceImpl) AddUser(u *User) error {
	// Lookup into a copy so a found user does not overwrite the signup data.
	err := us.repo.GetUserByUsername(&User{UserName: u.UserName})
	if err == nil {
		return apperr.ErrUserAlreadyExists
	}
	if !apperr.Is(err, apperr.ErrUserNotFound) {
		us.logger.Errorw("cannot check whether user exists", "username", u.UserName, "error", err)
		return err
	}
	us.logger.Infow("About to create a new user with",
		"name", u.UserName, "firstName", u.FirstName, "lastName", u.LastName)
	t := time.Now()
	u.CreatedAt = t
	u.ModifiedAt = t
	u.Status = "New"
	u.MustChangePassword = true
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
		return nil, err
	}
	return u, nil
}

func (us *userServiceImpl) CreateDefaultAdminUser() error {
	u := User{UserName: "admin"}
	err := us.repo.GetUserByUsername(&u)
	if apperr.Is(err, apperr.ErrUserNotFound) {
		t := time.Now()
		u.FirstName = "admin"
		u.LastName = "admin"
		u.CreatedAt = t
		u.ModifiedAt = t
		u.Status = "Verified"
		u.MustChangePassword = true
		var generatedPassword string
		for {
			generatedPassword, err = generateRandomPassword(20)
			if err != nil {
				us.logger.Infow("cannot generate random password", "error", err)
				return err
			}
			if meetsSigninPolicy(generatedPassword) {
				break
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
		us.logger.Warnw("Default admin created with a generated password; it must be changed at first login",
			"user", u.UserName, "password", generatedPassword)
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
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(enteredPassword+u.Salt))
	if err != nil {
		us.logger.Errorw("cannot validate password", "error", err)
		return apperr.ErrPasswordNotMatch
	}
	if u.Status != "Verified" {
		us.logger.Infow("user cannot login because not verified yet", "username", u.UserName)
		return apperr.ErrUserNotVerified
	}
	return nil
}

// ChangePassword checks u.Password as the current password, then stores newPassword
// and clears MustChangePassword. On return u holds the stored user.
func (us *userServiceImpl) ChangePassword(u *User, newPassword string) error {
	currentPassword := u.Password
	if err := us.ValidateUser(u); err != nil {
		return err
	}
	if newPassword == currentPassword {
		return apperr.ErrPasswordReused
	}
	salt, err := generateSalt()
	if err != nil {
		return err
	}
	hashedPassword, err := hashPassword(newPassword, salt)
	if err != nil {
		return err
	}
	if err := us.repo.UpdateUserPassword(u.ID, hashedPassword, salt); err != nil {
		us.logger.Errorw("Failed to update password", "user", u.UserName, "error", err)
		return err
	}
	u.Password, u.Salt, u.MustChangePassword = hashedPassword, salt, false
	us.logger.Infow("Password changed", "user", u.UserName)
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
		return apperr.ErrUserAlreadyVerified
	}
	u.Status = "Verified"
	err := us.repo.VerifyUser(u)
	if err != nil {
		return err
	}
	return nil
}

func (us *userServiceImpl) UpdateUser(u *User) error {
	u.ModifiedAt = time.Now()
	err := us.repo.UpdateUserData(u)
	if err != nil {
		return err
	}
	// Contact details live in the receptor cache; refill it so alerts use the new values.
	if us.cache != nil {
		us.cache.Refresh()
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
		return nil, apperr.ErrUsernameRequired
	}
	user := &User{UserName: name}
	err := us.repo.GetUserByUsername(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (us *userServiceImpl) GetByUserId(id string) (*User, error) {
	if id == "" {
		return nil, apperr.ErrUserIDRequired
	}
	user, err := us.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
