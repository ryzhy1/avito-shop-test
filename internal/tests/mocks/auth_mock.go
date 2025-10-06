package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type AuthRepositoryMock struct {
	mock.Mock
}

func (m *AuthRepositoryMock) SaveUser(ctx context.Context, login string, password []byte) error {
	args := m.Called(ctx, login, password)
	return args.Error(0)
}

func (m *AuthRepositoryMock) LoginUser(ctx context.Context, inputType, input string) (string, []byte, error) {
	args := m.Called(ctx, inputType, input)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}

func (m *AuthRepositoryMock) CheckUsernameIsAvailable(ctx context.Context, login string) (bool, error) {
	args := m.Called(ctx, login)
	return args.Bool(0), args.Error(1)
}
