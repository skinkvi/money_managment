package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/skinkvi/money_managment/internal/storage"
	"github.com/skinkvi/money_managment/pkg/logger"
)

type Repository interface {
	Create(ctx context.Context, u *User) (int64, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	Update(ctx context.Context, u *User) (*User, error)
	Delete(ctx context.Context, id int64) error

	// limit - максмальное количество записей
	// offset - смещение от начала
	List(ctx context.Context, limit, offset int) ([]User, error)
	// Эта функция нужна для пагинации для мобилки, она возвращает общее количество пользователей.
	Count(ctx context.Context) (int64, error)
}

type pgUserRepository struct {
	db  *storage.DB
	log logger.Logger
}

func NewUserRepository(db *storage.DB, log logger.Logger) Repository {
	return &pgUserRepository{db: db, log: log}
}

func (r *pgUserRepository) Create(ctx context.Context, u *User) (int64, error) {
	const query = `insert into users 
		(username, email, passhash)
		values
		($1, $2, $3)
		on conflict (email) do nothing
		returning id`

	var id int64

	err := r.db.Pool.QueryRow(ctx, query, u.Username, u.Email, u.PassHash).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, storage.ErrUserAlreadyExists
	}

	if err != nil {
		r.log.Error(ctx, "failed to create user", logger.Field{Key: "error", Value: err})
		return 0, fmt.Errorf("%w: %s", storage.ErrDB, err)
	}

	r.log.Info(ctx, "created user with id: ", logger.Field{Key: "user_id", Value: id})
	return id, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	const query = `select id, username, email, passhash, create_at, update_at
				   from users
				   where id = $1`

	rows, err := r.db.Pool.Query(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "failed to execute query GetByID",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: id})
		return nil, fmt.Errorf("failed GetByID query: %w", err)
	}
	defer rows.Close()

	var u User
	if rows.Next() {
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PassHash, &u.CreateAt, &u.UpdateAt); err != nil {
			r.log.Error(ctx, "failed to scan row GetByID",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "user_id", Value: id})
			return nil, fmt.Errorf("scan GetByID row: %w", err)
		}
	} else {
		if err := rows.Err(); err != nil {
			r.log.Error(ctx, "rows iteration err GetByID",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "user_id", Value: id})
			return nil, fmt.Errorf("rows integration GetByID: %w", err)
		}

		notFound := fmt.Errorf("user with id %d not found", id)
		r.log.Info(ctx, "user not found",
			logger.Field{Key: "user_id", Value: id})
		return nil, notFound
	}

	return &u, nil
}

func (r *pgUserRepository) Update(ctx context.Context, u *User) (*User, error) {
	const query = `update users 
	set username = $1, email = $2, passhash = $3, update_at = now()
	where id = $4
	returning id, username, email, passhash, create_at, update_at`

	var usr User

	if err := r.db.Pool.QueryRow(ctx, query, u.Username, u.Email, u.PassHash, u.ID).Scan(&usr.ID, &usr.Username, &usr.Email, &usr.PassHash, &usr.CreateAt, &usr.UpdateAt); err != nil {
		r.log.Error(ctx, "failed to execute query Update",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: u.ID})

		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Error(ctx, "user not found", logger.Field{Key: "user_id", Value: u.ID})
			return nil, fmt.Errorf("user with id %d not found: %w", u.ID, err)
		}

		return nil, fmt.Errorf("failed query Update: %w", err)
	}

	return &usr, nil
}

func (r *pgUserRepository) Delete(ctx context.Context, id int64) error {
	const query = `delete
	from users
	where id = $1`

	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "failed to execute query Delete",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "user_id", Value: id})

		return fmt.Errorf("failed delete user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		r.log.Error(ctx, "user not found", logger.Field{Key: "user_id", Value: id})
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}

func (r *pgUserRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
	const query = `select id, username, email, passhash, create_at, update_at
	from users
	order by id
	limit $1 offset $2`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.log.Error(ctx, "failed to execute query List",
			logger.Field{Key: "error", Value: err})

		return nil, fmt.Errorf("failed query List: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PassHash, &u.CreateAt, &u.UpdateAt); err != nil {
			r.log.Error(ctx, "failed scan List",
				logger.Field{Key: "error", Value: err})
			return nil, fmt.Errorf("failed scan user List: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		r.log.Error(ctx, "rows iteration error in users List",
			logger.Field{Key: "error", Value: err})
		return nil, fmt.Errorf("rows interation List: %w", err)
	}

	return users, nil
}

func (r *pgUserRepository) Count(ctx context.Context) (int64, error) {
	const query = `select count(id) from users`

	var count int64

	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.log.Error(ctx, "failed execute query Count", logger.Field{Key: "error", Value: err})
		return 0, fmt.Errorf("failed query Count: %w", err)
	}

	if count == 0 {
		r.log.Error(ctx, "not users found")
		return 0, storage.ErrNoUsers
	}

	return count, nil

}
