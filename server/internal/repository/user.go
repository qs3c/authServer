package repository

import (
	"authServer/server/internal/domain"
	"authServer/server/internal/repository/dao"
	"context"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *UserRepository) ModifyAuthTimesByEmail(ctx context.Context, email string, auth_hide_times int, auth_extract_times int) error {
	err := repo.dao.ModifyAuthTimesByEmail(ctx, email, auth_hide_times, auth_extract_times)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) HideMinusOneByUserId(ctx context.Context, userId int64) (int, error) {

	updatedTimes, err := repo.dao.HideMinusOneByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}

	return updatedTimes, nil
}

func (repo *UserRepository) ExtractMinusOneByUserId(ctx context.Context, userId int64) (int, error) {
	updatedTimes, err := repo.dao.ExtractMinusOneByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}

	return updatedTimes, nil
}

func (repo *UserRepository) CheckByUserId(ctx context.Context, userId int64) (int, int, error) {

	hideRemainTimes, extractRemainTimes, err := repo.dao.ExtractCheckByUserId(ctx, userId)
	if err != nil {
		return 0, 0, err
	}

	return hideRemainTimes, extractRemainTimes, nil
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}
