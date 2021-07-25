package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type Account struct {
	Id uint64
	Nickname string
	Avatar string
	Email string
	Platform int
	OpenId string
	State int8
	CreatedAt int64
	UpdatedAt int64
}

type AccountRepo interface {
	GetAccount(ctx context.Context, id uint64) (*Account, error)
	ListByIds(ctx context.Context, ids []uint64) ([]*Account, error)
	Login(ctx context.Context, email string, password string) (*Account, string, error)
}

type AccountUsecase struct {
	repo AccountRepo
	log *log.Helper
}

func NewAccountUsecase(repo AccountRepo, logger log.Logger) *AccountUsecase {
	return &AccountUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *AccountUsecase) GetAccount(ctx context.Context, id uint64) (*Account, error) {
	return uc.repo.GetAccount(ctx, id)
}

func (uc *AccountUsecase) Login(ctx context.Context, email string, password string) (account *Account, token string, err error) {
	return uc.repo.Login(ctx, email, password)
}