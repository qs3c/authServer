package service

import (
	"authServer/server/internal/domain"
	"authServer/server/internal/repository"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserService) InitAuthTimes(ctx context.Context, email string, auth_hide_times int, auth_extract_times int) error {

	// 根据 email 找到用户并修改次数记录
	err := svc.repo.ModifyAuthTimesByEmail(ctx, email, auth_hide_times, auth_extract_times)
	if err != nil {
		return err
	}
	return nil

}

func (svc *UserService) CheckAuthTimes(ctx *gin.Context, userId int64) (int, int, error) {

	hideRemainTimes, extractRemainTimes, err := svc.repo.CheckByUserId(ctx, userId)
	if err != nil {
		return 0, 0, err
	}
	return hideRemainTimes, extractRemainTimes, nil

}

func (svc *UserService) MinusOneAuthTimes(ctx *gin.Context, userId int64, authType bool) (int, error) {

	if authType == false {
		remainTimes, err := svc.repo.HideMinusOneByUserId(ctx, userId)
		if err != nil {
			return 0, err
		}
		return remainTimes, nil
	} else {
		remainTimes, err := svc.repo.ExtractMinusOneByUserId(ctx, userId)
		if err != nil {
			return 0, err
		}
		return remainTimes, nil
	}
}
