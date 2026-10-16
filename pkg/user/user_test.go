package user

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

// fakeUserRepo implements UserInterfaceRepository; only UpdateUserData is exercised.
type fakeUserRepo struct {
	UserInterfaceRepository
	updateErr error
}

func (f *fakeUserRepo) UpdateUserData(*User) error { return f.updateErr }

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
		{name: "failed update does not refresh", repoErr: errors.New("update failed"), wantN: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &countingCache{}
			s := NewUserService(&fakeUserRepo{updateErr: tt.repoErr}, nil, c, logger)
			err := s.UpdateUser(&User{ID: "u1", Mobile: "09120000000"})
			if !errors.Is(err, tt.repoErr) {
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
