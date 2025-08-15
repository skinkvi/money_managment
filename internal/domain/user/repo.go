package user

import "context"

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
