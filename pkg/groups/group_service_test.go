package groups

import (
	"testing"

	"github.com/root-ali/iris/internal/apperr"
	"go.uber.org/zap"
)

type fakeGroupRepo struct {
	exists   bool
	getErr   error
	writeErr error
}

func (f *fakeGroupRepo) AddGroup(*Group) error { return f.writeErr }
func (f *fakeGroupRepo) GetGroupById(string) (*Group, error) {
	return nil, apperr.ErrGroupNotFound
}
func (f *fakeGroupRepo) GetGroupByName(string) (*Group, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.exists {
		return &Group{ID: "g1", Name: "ops"}, nil
	}
	return nil, apperr.ErrGroupNotFound
}
func (f *fakeGroupRepo) GetAllGroups() ([]*Group, error)          { return nil, nil }
func (f *fakeGroupRepo) DeleteGroup(*Group) error                 { return f.writeErr }
func (f *fakeGroupRepo) AddUserToGroup(string, string) error      { return f.writeErr }
func (f *fakeGroupRepo) RemoveUserFromGroup(string, string) error { return f.writeErr }
func (f *fakeGroupRepo) FindUsersById(string) ([]string, error)   { return nil, nil }
func (f *fakeGroupRepo) FindGroupById(string) ([]string, error)   { return nil, nil }

type countingCache struct{ n int }

func (c *countingCache) Refresh() { c.n++ }

func TestGroupChangesRefreshCache(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name    string
		repo    *fakeGroupRepo
		call    func(GroupServiceInterface) error
		wantN   int
		wantErr error
	}{
		{
			name:  "create group",
			repo:  &fakeGroupRepo{},
			call:  func(s GroupServiceInterface) error { return s.CreateGroup(&Group{Name: "ops"}) },
			wantN: 1,
		},
		{
			name:    "create existing group",
			repo:    &fakeGroupRepo{exists: true},
			call:    func(s GroupServiceInterface) error { return s.CreateGroup(&Group{Name: "ops"}) },
			wantErr: apperr.ErrGroupAlreadyExists,
		},
		{
			name:    "create group write failure",
			repo:    &fakeGroupRepo{writeErr: apperr.New("insert failed")},
			call:    func(s GroupServiceInterface) error { return s.CreateGroup(&Group{Name: "ops"}) },
			wantErr: apperr.New("insert failed"),
		},
		{
			name:    "create group lookup failure",
			repo:    &fakeGroupRepo{getErr: apperr.New("db down")},
			call:    func(s GroupServiceInterface) error { return s.CreateGroup(&Group{Name: "ops"}) },
			wantErr: apperr.New("db down"),
		},
		{
			name:  "delete group",
			repo:  &fakeGroupRepo{},
			call:  func(s GroupServiceInterface) error { return s.DeleteGroup(&Group{ID: "g1"}) },
			wantN: 1,
		},
		{
			name:    "delete group failure",
			repo:    &fakeGroupRepo{writeErr: apperr.New("delete failed")},
			call:    func(s GroupServiceInterface) error { return s.DeleteGroup(&Group{ID: "g1"}) },
			wantErr: apperr.New("delete failed"),
		},
		{
			name:  "add user",
			repo:  &fakeGroupRepo{},
			call:  func(s GroupServiceInterface) error { return s.AddUser(&Group{ID: "g1"}, "u1") },
			wantN: 1,
		},
		{
			name:    "add user failure",
			repo:    &fakeGroupRepo{writeErr: apperr.New("add failed")},
			call:    func(s GroupServiceInterface) error { return s.AddUser(&Group{ID: "g1"}, "u1") },
			wantErr: apperr.New("add failed"),
		},
		{
			name:  "remove user",
			repo:  &fakeGroupRepo{},
			call:  func(s GroupServiceInterface) error { return s.RemoveUser(&Group{ID: "g1"}, "u1") },
			wantN: 1,
		},
		{
			name:    "remove user failure",
			repo:    &fakeGroupRepo{writeErr: apperr.New("remove failed")},
			call:    func(s GroupServiceInterface) error { return s.RemoveUser(&Group{ID: "g1"}, "u1") },
			wantErr: apperr.New("remove failed"),
		},
		{
			name: "read does not refresh",
			repo: &fakeGroupRepo{},
			call: func(s GroupServiceInterface) error {
				_, err := s.GetAllGroups()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := &countingCache{}
			svc := NewGroupService(logger, tt.repo, cache)
			err := tt.call(svc)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if cache.n != tt.wantN {
				t.Fatalf("refresh calls = %d, want %d", cache.n, tt.wantN)
			}
		})
	}
}

func TestGroupChangeWithoutCacheDoesNotPanic(t *testing.T) {
	t.Parallel()
	svc := NewGroupService(zap.NewNop().Sugar(), &fakeGroupRepo{}, nil)
	if err := svc.AddUser(&Group{ID: "g1"}, "u1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveUser(&Group{ID: "g1"}, "u1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteGroup(&Group{ID: "g1"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateGroup(&Group{Name: "ops"}); err != nil {
		t.Fatal(err)
	}
}
