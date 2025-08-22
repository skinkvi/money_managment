package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/skinkvi/money_managment/internal/storage"
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

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	u := &User{
		Username: "dima",
		Email:    "dima@example.com",
		PassHash: "hash",
	}
	id, err := repo.Create(ctx, u)

	require.NoError(t, err)
	require.Equal(t, int64(42), id)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Create_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`insert into users`).
		WithArgs("dima", "dima@example.com", "hash").
		WillReturnError(pgx.ErrNoRows)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	u := &User{
		Username: "dima",
		Email:    "dima@example.com",
		PassHash: "hash",
	}

	id, err := repo.Create(ctx, u)

	require.Error(t, err)
	require.ErrorIs(t, err, storage.ErrUserAlreadyExists)
	require.Equal(t, int64(0), id)

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserReposotory_GetByID_Success(t *testing.T) {
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	want := &User{
		ID:       42,
		Username: "dima",
		Email:    "dima@example",
		PassHash: "hash",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}

	mockPool.ExpectQuery(`select username, email, passhash, create_at, update_at from users`).
		WithArgs(want.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"username", "email", "passhash", "create_at", "update_at",
			}).AddRow(
				want.Username,
				want.Email,
				want.PassHash,
				want.CreateAt,
				want.UpdateAt,
			),
		)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	got, err := repo.GetByID(ctx, want.ID)

	require.NoError(t, err)
	require.Equal(t, want.Username, got.Username)
	require.Equal(t, want.Email, got.Email)
	require.Equal(t, want.PassHash, got.PassHash)
	require.WithinDuration(t, want.CreateAt, got.CreateAt, time.Second)
	require.WithinDuration(t, want.UpdateAt, got.UpdateAt, time.Second)

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserReposotory_GetByID_ErrorExecuteQuery(t *testing.T) {
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`select username, email, passhash, create_at, update_at from users`).
		WithArgs(int64(42)).
		WillReturnError(fmt.Errorf("connection lost"))

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	_, err = repo.GetByID(ctx, 42)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed GetByID query")
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_GetByID_ScanError(t *testing.T) {
	ctx := context.Background()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`select username, email, passhash, create_at, update_at from users`).
		WithArgs(int64(42)).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"username", "email", "passhash", "create_at", "update_at",
			}).AddRow(
				"dima",
				"dima@example.com",
				"hash",
				"bad_time_value",
				"BTV2",
			),
		)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	usr, err := repo.GetByID(ctx, 42)

	require.Error(t, err, "ожидаем ошибку сканирования")
	require.Contains(t, err.Error(), "scan GetByID row")
	require.Nil(t, usr)

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`select username, email, passhash, create_at, update_at`).
		WithArgs(int64(99)).
		WillReturnError(pgx.ErrNoRows)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	usr, err := repo.GetByID(ctx, 99)

	require.Error(t, err)
	require.Nil(t, usr)

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	const query = `update users 
	set username = $1, email = $2, passhash = $3, update_at = now()
	where id = $4
	returning id, username, email, passhash, create_at, update_at`

	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	u := &User{
		ID:       1,
		Username: "dima",
		Email:    "dima@example.com",
		PassHash: "hash",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}

	mockPool.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(u.Username, u.Email, u.PassHash, u.ID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "username", "email", "passhash", "create_at", "update_at",
		}).AddRow(
			u.ID,
			u.Username,
			u.Email,
			u.PassHash,
			u.CreateAt,
			u.UpdateAt,
		),
		)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	got, err := repo.Update(ctx, u)

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, u.ID, got.ID)
	require.Equal(t, u.Username, got.Username)
	require.Equal(t, u.Email, got.Email)
	require.Equal(t, u.PassHash, got.PassHash)
	require.Equal(t, u.CreateAt, got.CreateAt)
	require.Equal(t, u.UpdateAt, got.UpdateAt)

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	const query = `update users 
	set username = $1, email = $2, passhash = $3, update_at = now()
	where id = $4
	returning id, username, email, passhash, create_at, update_at`

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	u := &User{
		ID:       123,
		Username: "kdfjdfj",
		Email:    "jfkdjfk",
		PassHash: "fjkdjfkd",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}

	mockPool.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(u.Username, u.Email, u.PassHash, u.ID).
		WillReturnError(pgx.ErrNoRows)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	got, err := repo.Update(context.Background(), u)

	require.Error(t, err)
	require.Nil(t, got)
	require.Contains(t, err.Error(), "user with id 123 not found")

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Update_DBError(t *testing.T) {
	const query = `update users 
	set username = $1, email = $2, passhash = $3, update_at = now()
	where id = $4
	returning id, username, email, passhash, create_at, update_at`

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	u := &User{
		ID:       123,
		Username: "dima",
		Email:    "dima@example.com",
		PassHash: "hash",
	}

	origErr := errors.New("some db error")

	mockPool.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(u.Username, u.Email, u.PassHash, u.ID).
		WillReturnError(origErr)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	got, err := repo.Update(context.Background(), u)

	require.Error(t, err)
	require.Nil(t, got)
	require.ErrorContains(t, err, "failed query Update:")
	require.True(t, errors.Is(err, origErr),
		"orig err must be wrapped (use %w in fmt.Errorf)")
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Delete_Success(t *testing.T) {
	const query = `delete
	from users
	where id = $1`

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(int64(1)).
		WillReturnResult(pgconn.NewCommandTag("DELETE 1"))

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	u := &User{
		ID: 1,
	}

	err = repo.Delete(context.Background(), u.ID)

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Delete_DriverErr(t *testing.T) {
	const deleteQuery = `delete
	from users
	where id = $1`

	const insertQuery = `insert into users 
		(username, email, passhash)
		values
		($1, $2, $3)
		on conflict (email) do nothing
		returning id`

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	driverErr := errors.New("connection closed(maybe)")

	u := &User{
		Username: "fjskdajf",
		Email:    "fkdjfkd@example.com",
		PassHash: "jfdkfjd",
	}

	mockPool.ExpectQuery(regexp.QuoteMeta(insertQuery)).
		WithArgs(u.Username, u.Email, u.PassHash).
		WillReturnRows(
			pgxmock.NewRows([]string{"id"}).
				AddRow(int64(1)),
		)

	mockPool.ExpectExec(regexp.QuoteMeta(deleteQuery)).
		WithArgs(int64(1)).
		WillReturnError(driverErr)

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	created, err := repo.Create(context.Background(), u)
	require.NoError(t, err)
	require.NotNil(t, created)

	err = repo.Delete(context.Background(), created)
	require.Error(t, err)
	require.True(t, errors.Is(err, driverErr), "orig error must be wrappet")
	require.ErrorContains(t, err, "failed delete user:")
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestUserRepository_Delete_UserNotFound(t *testing.T) {
	const deleteQuery = `delete
	from users
	where id = $1`

	const insertQuery = `insert into users 
		(username, email, passhash)
		values
		($1, $2, $3)
		on conflict (email) do nothing
		returning id`

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	u := &User{
		Username: "fjskdajf",
		Email:    "fkdjfkd@example.com",
		PassHash: "jfdkfjd",
	}

	mockPool.ExpectQuery(regexp.QuoteMeta(insertQuery)).
		WithArgs(u.Username, u.Email, u.PassHash).
		WillReturnRows(
			pgxmock.NewRows([]string{"id"}).
				AddRow(int64(1)),
		)

	mockPool.ExpectExec(regexp.QuoteMeta(deleteQuery)).
		WithArgs(int64(1)).
		WillReturnResult(pgconn.NewCommandTag("DELETE 0"))

	db := &storage.DB{Pool: mockPool}
	repo := NewUserRepository(db, nopLogger{})

	created, err := repo.Create(context.Background(), u)
	require.NoError(t, err)
	require.NotNil(t, created)

	err = repo.Delete(context.Background(), created)
	require.Error(t, err)
	require.ErrorContains(t, err, "user with id 1 not found")
	require.NoError(t, mockPool.ExpectationsWereMet())
}
