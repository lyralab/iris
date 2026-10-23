package user

import (
	"testing"

	"github.com/root-ali/iris/internal/apperr"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// fakeUserRepo implements UserInterfaceRepository; only the methods below are exercised.
type fakeUserRepo struct {
	UserInterfaceRepository
	updateErr error
	getErr    error // returned by GetUserByUsername when set
	stored    *User // returned by GetUserByUsername; nil means not found
	added     *User
	newHash   string
	newSalt   string
}

func (f *fakeUserRepo) UpdateUserData(*User) error { return f.updateErr }

func (f *fakeUserRepo) GetUserByUsername(u *User) error {
	if f.getErr != nil {
		return f.getErr
	}
	if f.stored == nil {
		return apperr.ErrUserNotFound
	}
	*u = *f.stored
	return nil
}

func (f *fakeUserRepo) AddUser(u *User) error { f.added = u; return nil }

func (f *fakeUserRepo) UpdateUserPassword(_, hash, salt string) error {
	f.newHash, f.newSalt = hash, salt
	return nil
}

type fakeRoleRepo struct{}

func (fakeRoleRepo) GetRoleIDByName(name string) (string, error) { return name + "-id", nil }

func TestCreatedUsersMustChangePassword(t *testing.T) {
	logger := zap.NewNop().Sugar()

	repo := &fakeUserRepo{}
	s := NewUserService(repo, fakeRoleRepo{}, nil, logger)
	if err := s.AddUser(&User{UserName: "bob", Password: "Secret#123"}); err != nil {
		t.Fatalf("AddUser() = %v", err)
	}
	if !repo.added.MustChangePassword {
		t.Error("AddUser: MustChangePassword = false, want true")
	}

	repo = &fakeUserRepo{}
	s = NewUserService(repo, fakeRoleRepo{}, nil, logger)
	if err := s.CreateDefaultAdminUser(); err != nil {
		t.Fatalf("CreateDefaultAdminUser() = %v", err)
	}
	if !repo.added.MustChangePassword {
		t.Error("CreateDefaultAdminUser: MustChangePassword = false, want true")
	}
}

func TestAddUserExistenceCheck(t *testing.T) {
	t.Parallel()
	dbDown := apperr.New("connection refused")
	tests := []struct {
		name      string
		repo      *fakeUserRepo
		wantErr   error
		wantAdded bool
	}{
		{name: "new user is created", repo: &fakeUserRepo{}, wantAdded: true},
		{name: "existing user is rejected", repo: &fakeUserRepo{stored: &User{UserName: "bob"}}, wantErr: apperr.ErrUserAlreadyExists},
		{name: "lookup failure is returned", repo: &fakeUserRepo{getErr: dbDown}, wantErr: dbDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewUserService(tt.repo, fakeRoleRepo{}, nil, zap.NewNop().Sugar())
			err := s.AddUser(&User{UserName: "bob", Password: "Secret#123"})
			if !apperr.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if (tt.repo.added != nil) != tt.wantAdded {
				t.Fatalf("added = %v, want %v", tt.repo.added != nil, tt.wantAdded)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	t.Parallel()
	const current = "Current#123"
	salt, _ := generateSalt()
	hash, _ := hashPassword(current, salt)
	stored := User{ID: "u1", UserName: "bob", Password: hash, Salt: salt, Status: "Verified", MustChangePassword: true}

	tests := []struct {
		name    string
		entered string
		newPass string
		wantErr error
	}{
		{name: "wrong current password", entered: "Wrong#1234", newPass: "Fresh#1234", wantErr: apperr.ErrPasswordNotMatch},
		{name: "same as current", entered: current, newPass: current, wantErr: apperr.ErrPasswordReused},
		{name: "success", entered: current, newPass: "Fresh#1234"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s0 := stored
			repo := &fakeUserRepo{stored: &s0}
			s := NewUserService(repo, nil, nil, zap.NewNop().Sugar())
			u := &User{UserName: "bob", Password: tt.entered}
			err := s.ChangePassword(u, tt.newPass)
			if !apperr.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if repo.newHash != "" {
					t.Fatal("password updated despite error")
				}
				return
			}
			if bcrypt.CompareHashAndPassword([]byte(repo.newHash), []byte(tt.newPass+repo.newSalt)) != nil {
				t.Fatal("stored hash does not match new password")
			}
			if u.MustChangePassword {
				t.Fatal("MustChangePassword still set after change")
			}
		})
	}
}

type countingCache struct{ n int }

func (c *countingCache) Refresh() { c.n++ }

func TestUpdateUserRefreshesReceptorCache(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name    string
		repoErr error
		wantN   int
	}{
		{name: "successful update refreshes once", wantN: 1},
		{name: "failed update does not refresh", repoErr: apperr.New("update failed"), wantN: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &countingCache{}
			s := NewUserService(&fakeUserRepo{updateErr: tt.repoErr}, nil, c, logger)
			err := s.UpdateUser(&User{ID: "u1", Mobile: "09120000000"})
			if !apperr.Is(err, tt.repoErr) {
				t.Fatalf("err = %v, want %v", err, tt.repoErr)
			}
			if c.n != tt.wantN {
				t.Fatalf("Refresh called %d times, want %d", c.n, tt.wantN)
			}
		})
	}
}

func TestUpdateUserNilCacheDoesNotPanic(t *testing.T) {
	t.Parallel()
	s := NewUserService(&fakeUserRepo{}, nil, nil, zap.NewNop().Sugar())
	if err := s.UpdateUser(&User{ID: "u1"}); err != nil {
		t.Fatalf("UpdateUser() = %v", err)
	}
}
