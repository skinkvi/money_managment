package postgres

import (
	"context"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/skinkvi/money_managment/internal/domain/user"
	"github.com/skinkvi/money_managment/pkg/logger"
	"github.com/stretchr/testify/require"
)

type nopLogger struct{}

func (nopLogger) Debug(ctx context.Context, msg string, fields ...logger.Field) {}
func (nopLogger) Info(ctx context.Context, msg string, fields ...logger.Field)  {}
func (nopLogger) Warn(ctx context.Context, msg string, fields ...logger.Field)  {}
func (nopLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {}

// With обязателен по интерфейсу logger.Logger
func (nopLogger) With(fields ...logger.Field) logger.Logger { return nopLogger{} }

// Sync тоже обязателен, но в тестах ничего не делает
func (nopLogger) Sync() error { return nil }

func TestUserRepository_Create_Success(t *testing.T) {
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`insert into users`).
		WithArgs("dima", "dima@example.com", "hash").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(42)))

	db := &DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	u := &user.User{
		Username: "dima",
		Email:    "dima@example.com",
		PassHash: "hash",
	}
	id, err := repo.Create(ctx, u)

	require.NoError(t, err)
	require.Equal(t, int64(42), id)
	require.NoError(t, mockPool.ExpectationsWereMet())

}
