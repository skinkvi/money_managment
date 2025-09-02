package user

import (
	"context"
	"errors"
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

// Вынес в константы все запросы что бы не писать их постоянно + они не изменяемы
const (
	insertQuery  = `insert into users`
	updateQuery  = `update users set username = $1, email = $2, passhash = $3, update_at = now() where id = $4 returning id, username, email, passhash, create_at, update_at`
	getByIDQuery = `select id, username, email, passhash, create_at, update_at from users`
	deleteQuery  = `delete from users where id = $1`
	listQuery    = `select id, username, email, passhash, create_at, update_at from users order by id limit $1 offset $2`
	countQuery   = `select count(id) from users`
)

func newTestRepo(t *testing.T) (Repository, pgxmock.PgxPoolIface) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() {
		mockPool.Close()
	})
	db := &storage.DB{Pool: mockPool}
	return NewUserRepository(db, nopLogger{}), mockPool
}

var (
	fixedTime = time.Now()
)

func TestUserRepository_GetByID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		mockSetup func(pgxmock.PgxPoolIface)
		wantUser  *User
		wantErr   string
		wantErrIs error
	}{
		{
			name: "success",
			mockSetup: func(p pgxmock.PgxPoolIface) {
				p.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
					WithArgs(int64(42)).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "username", "email", "passhash", "create_at", "update_at",
					}).AddRow(42, "dima", "dima@example.com", "hash", fixedTime, fixedTime))
			},
			wantUser: &User{ID: 42, Username: "dima", Email: "dima@example.com",
				PassHash: "hash", CreateAt: fixedTime, UpdateAt: fixedTime},
		},
		{
			name: "not found",
			mockSetup: func(p pgxmock.PgxPoolIface) {
				p.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
					WithArgs(int64(99)).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: "failed GetByID query",
		},
		{
			name: "query error",
			mockSetup: func(p pgxmock.PgxPoolIface) {
				p.ExpectQuery(regexp.QuoteMeta(getByIDQuery)).
					WithArgs(int64(42)).
					WillReturnError(errors.New("database connection lost"))
			},
			wantErr: "failed GetByID query: database connection lost",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			got, err := repo.GetByID(context.Background(), 42)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			if tc.wantErrIs != nil {
				require.ErrorIs(t, err, tc.wantErrIs)
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantUser, got)
		})
	}
}

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		mockSetup func(pgxmock.PgxPoolIface)
		wantID    int64
		wantErr   string
		wantErrIs error
	}{
		{
			name: "success",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(
					insertQuery)).
					WithArgs("dima", "dima@example.com", "hash").
					WillReturnRows(pgxmock.NewRows(
						[]string{"id"}).AddRow(int64(42)))
			},
			wantID: 42,
		},
		{
			name: "already exists",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs("dima", "dima@example.com", "hash").
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr:   storage.ErrUserAlreadyExists.Error(),
			wantErrIs: storage.ErrUserAlreadyExists,
			wantID:    0,
		},
		{
			name: "database error",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs("dima", "dima@example.com", "hash").
					WillReturnError(errors.New("failed to create user"))
			},
			wantErr:   "failed to create user",
			wantErrIs: storage.ErrDB,
			wantID:    0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			gotID, err := repo.Create(context.Background(), &User{
				Username: "dima",
				Email:    "dima@example.com",
				PassHash: "hash",
			})

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			if tc.wantErrIs != nil {
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantID, gotID)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		mockSetup func(pgxmock.PgxPoolIface)
		wantUser  *User
		wantErr   string
		wantErrIs error
	}{
		{
			name: "success",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				u := &User{
					ID:       1,
					Username: "dima",
					Email:    "dima@example.com",
					PassHash: "hash",
					CreateAt: fixedTime,
					UpdateAt: fixedTime,
				}

				ppi.ExpectQuery(regexp.QuoteMeta(updateQuery)).
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
					))
			},
			wantUser: &User{
				ID:       1,
				Username: "dima",
				Email:    "dima@example.com",
				PassHash: "hash",
				CreateAt: fixedTime,
				UpdateAt: fixedTime,
			},
		},
		{
			name: "error not found",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				u := &User{
					ID:       1,
					Username: "dima",
					Email:    "dima@example.com",
					PassHash: "hash",
				}

				ppi.ExpectQuery(regexp.QuoteMeta(updateQuery)).
					WithArgs(u.Username, u.Email, u.PassHash, u.ID).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr:   "user with id 1 not found",
			wantErrIs: pgx.ErrNoRows,
		},
		{
			name: "error db",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				u := &User{
					ID:       1,
					Username: "dima",
					Email:    "dima@example.com",
					PassHash: "hash",
				}

				origErr := errors.New("some db error")

				ppi.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs(u.Username, u.Email, u.PassHash, u.ID).
					WillReturnError(origErr)
			},
			wantErr: "failed query Update:",
		},
		{
			name: "scan error",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(insertQuery)).
					WithArgs("dima", "dima@example.com", "hash", int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "username", "email", "passhash", "create_at", "update_at",
					}).AddRow(
						int64(1), "dima", "dima@example.com", "hash", "invalid-time", fixedTime,
					))
			},
			wantErr: "failed query Update:",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			got, err := repo.Update(context.Background(), &User{
				ID:       1,
				Username: "dima",
				Email:    "dima@example.com",
				PassHash: "hash",
			})

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			if tc.wantErrIs != nil {
				require.ErrorIs(t, err, tc.wantErrIs)
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantUser, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	cases := []struct {
		name      string
		mockSetup func(pgxmock.PgxPoolIface)
		inputID   int64
		wantErr   string
		wantErrIs error
	}{
		{
			name: "success",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectExec(regexp.QuoteMeta(deleteQuery)).
					WithArgs(int64(1)).
					WillReturnResult(pgconn.NewCommandTag("DELETE 1"))
			},
			inputID: 1,
		},
		{
			name: "driver error",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectExec(regexp.QuoteMeta(deleteQuery)).
					WithArgs(int64(1)).
					WillReturnError(errors.New("connection closed"))
			},

			inputID: 1,
			wantErr: "failed delete user:",
		},
		{
			name: "user not found",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectExec(regexp.QuoteMeta(deleteQuery)).
					WithArgs(int64(1)).
					WillReturnResult(pgconn.NewCommandTag("DELETE 0"))
			},

			inputID: 1,
			wantErr: "user with id 1 not found",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			err := repo.Delete(context.Background(), tc.inputID)

			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	cases := []struct {
		name      string
		limit     int
		offset    int
		mockSetup func(pgxmock.PgxPoolIface)
		wantUsers []User
		wantErr   string
		wantErrIs error
	}{
		{
			name:   "success multiple users",
			limit:  2,
			offset: 0,
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WithArgs(2, 0).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "username", "email", "passhash", "create_at", "update_at",
					}).
						AddRow(int64(1), "user1", "user1@example.com", "hash1", fixedTime, fixedTime).
						AddRow(int64(2), "user2", "user2@example.com", "hash2", fixedTime, fixedTime))
			},
			wantUsers: []User{
				{ID: 1, Username: "user1", Email: "user1@example.com", PassHash: "hash1", CreateAt: fixedTime, UpdateAt: fixedTime},
				{ID: 2, Username: "user2", Email: "user2@example.com", PassHash: "hash2", CreateAt: fixedTime, UpdateAt: fixedTime},
			},
		},
		{
			name:   "success",
			limit:  2,
			offset: 0,
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WithArgs(2, 0).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "username", "email", "passhash", "create_at", "update_at",
					}))
			},
			wantUsers: nil,
		},
		{
			name:   "query error",
			limit:  1,
			offset: 0,
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WithArgs(1, 0).
					WillReturnError(errors.New("database connection lost"))
			},
			wantErr: "failed query List: database connection lost",
		},
		{
			name:   "scan error",
			limit:  1,
			offset: 0,
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WithArgs(1, 0).
					WillReturnRows(pgxmock.NewRows([]string{
						"id", "username", "email", "passhash", "create_at", "update_at",
					}).
						AddRow(int64(2), "user2", "user2@example.com", "hash2", "fake time for force error", fixedTime))
			},
			wantErr: "failed scan user List:",
		},
		{
			name:   "rows iteration error",
			limit:  1,
			offset: 0,
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "email", "passhash", "create_at", "update_at",
				}).AddRow(int64(1), "user1", "user1@example.com", "hash1", fixedTime, fixedTime)
				rows.RowError(0, errors.New("iteration error"))
				ppi.ExpectQuery(regexp.QuoteMeta(listQuery)).
					WithArgs(1, 0).
					WillReturnRows(rows)
			},
			wantErr: "failed scan user List: iteration error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			got, err := repo.List(context.Background(), tc.limit, tc.offset)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			if tc.wantErrIs != nil {
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantUsers, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Count(t *testing.T) {
	cases := []struct {
		name      string
		mockSetup func(pgxmock.PgxPoolIface)
		wantCount int64
		wantErr   string
		wantErrIs error
	}{
		{
			name: "success",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(countQuery)).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(5)))
			},
			wantCount: 5,
		},
		{
			name: "success no users",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(countQuery)).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))
			},
			wantCount: 0,
			wantErrIs: storage.ErrNoUsers,
		},
		{
			name: "query error",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(countQuery)).
					WillReturnError(errors.New("database connection lost"))
			},
			wantCount: 0,
			wantErr:   "failed query Count: database connection lost",
		},
		{
			name: "scan error",
			mockSetup: func(ppi pgxmock.PgxPoolIface) {
				ppi.ExpectQuery(regexp.QuoteMeta(countQuery)).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow("fjkdfjdk"))
			},
			wantCount: 0,
			wantErr:   "failed query Count:",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, mock := newTestRepo(t)
			tc.mockSetup(mock)

			got, err := repo.Count(context.Background())

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			if tc.wantErrIs != nil {
				require.ErrorIs(t, err, tc.wantErrIs)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantCount, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
