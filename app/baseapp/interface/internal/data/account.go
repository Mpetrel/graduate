package data

import (
	v1 "base-service/api/account/service/v1"
	"base-service/app/baseapp/interface/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type accountRepo struct {
	data *Data
	log *log.Helper
}

func NewAccountRepo(data *Data, logger log.Logger) biz.AccountRepo {
	return &accountRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}



func (a accountRepo) GetAccount(ctx context.Context, id uint64) (*biz.Account, error) {
	result, err := a.data.ac.GetAccount(ctx, &v1.GetAccountRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return toBizAccount(result.Account), nil
}


func (a accountRepo) ListByIds(ctx context.Context, ids []uint64) ([]*biz.Account, error) {
	result, err := a.data.ac.ListWithIds(ctx, &v1.ListWithIdsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	accounts := make([]*biz.Account, len(result.Accounts))
	for i := range result.Accounts {
		accounts[i] = toBizAccount(result.Accounts[i])
	}
	return accounts, nil
}

func (a accountRepo) Login(ctx context.Context, email string, password string) (*biz.Account, string, error) {
	result, err := a.data.ac.EmailLogin(ctx, &v1.EmailLoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, "", err
	}
	return toBizAccount(result.Account), result.Token, nil
}


func toBizAccount(account *v1.AccountInfo) *biz.Account {
	return &biz.Account{
		Id:        account.Id,
		Nickname:  account.Nickname,
		Avatar:    account.Avatar,
		Email:     account.Email,
		Platform:  int(account.Platform),
		OpenId:    account.OpenId,
		State: 	   int8(account.State),
		CreatedAt: account.CreatedAt,
	}
}