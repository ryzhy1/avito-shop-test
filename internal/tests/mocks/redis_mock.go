package mocks

import "github.com/stretchr/testify/mock"

type RedisClientMock struct {
	mock.Mock
}

func (m *RedisClientMock) StoreRefreshToken(userID, refreshToken string) error {
	args := m.Called(userID, refreshToken)
	return args.Error(0)
}
