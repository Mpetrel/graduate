package biz

import (
	"base-service/app/account/service/internal/pkg/passwd"
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"strings"
	"time"
)

type Account struct {
	Id uint64
	Nickname string
	Avatar string
	Email string
	Password string
	Salt string
	Platform int
	OpenId string
	State int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AccountRepo interface {
	CreateAccountByEmail(ctx context.Context, account *Account) error
	OauthAccount(ctx context.Context, account *Account) error
	GetAccount(ctx context.Context, email string) (*Account, error)
	GetAccountById(ctx context.Context, id uint64) (*Account, error)
	GetAccountByOpenId(ctx context.Context, openId string) (*Account, error)
	UpdateAccount(ctx context.Context, account *Account) error
	GetAccountByIds(ctx context.Context, ids []uint64) ([]*Account, error)
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

func (uc *AccountUsecase) CreateAccount(ctx context.Context, account *Account) error {
	if account.Nickname == "" { // 昵称为空把邮箱名设置为昵称
		sepStr := strings.Split(account.Email, "@")
		if len(sepStr) > 0 {
			account.Nickname = sepStr[0]
		}
	}
	// 密码处理
	encryptPsd, salt, err := passwd.BCrypt(account.Password)
	if err != nil {
		return errors.New(5002, "encrypt password failed", "create account failed")
	}
	account.Password = encryptPsd
	account.Salt = salt
	return uc.repo.CreateAccountByEmail(ctx, account)
}

func (uc *AccountUsecase) GetAccount(ctx context.Context, id uint64) (*Account, error) {
	return uc.repo.GetAccountById(ctx, id)
}

func (uc *AccountUsecase) UpdateAccount(ctx context.Context, account *Account) error {
	return uc.repo.UpdateAccount(ctx, account)
}

func (uc *AccountUsecase) GetAccountByIds(ctx context.Context, ids []uint64) ([]*Account, error) {
	return uc.repo.GetAccountByIds(ctx, ids)
}

// EmailLogin 邮箱登录
func (uc *AccountUsecase) EmailLogin(ctx context.Context, account *Account) error {
	savedAccount, err := uc.repo.GetAccount(ctx, account.Email)
	if err != nil {
		return err
	}
	valid := passwd.Verify(account.Password, savedAccount.Salt, savedAccount.Password)
	if !valid {
		return errors.New(5001,"invalid email or password", "invalid email or password")
	}
	*account = *savedAccount
	return nil
}



