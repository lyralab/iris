package asiatech

import (
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"github.com/stretchr/testify/mock"
)

// MockCache implements cache.Interface[string, string] for testing
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) (string, bool) {
	args := m.Called(key)
	return args.String(0), args.Bool(1)
}

func (m *MockCache) Set(key string, value string, duration time.Duration) error {
	args := m.Called(key, value, duration)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) {
	m.Called(key)
}

// Ensure MockCache implements the interface
var _ cache.Interface[string, string] = (*MockCache)(nil)
