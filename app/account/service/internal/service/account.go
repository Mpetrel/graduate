package service

import (
	"base-service/app/account/service/internal/biz"
	"base-service/app/account/service/internal/pkg/errs"
	"context"
	"github.com/go-kratos/kratos/v2/log"

	pb "base-service/api/account/service/v1"
)

type AccountService struct {
	pb.UnimplementedAccountServer
	uc *biz.AccountUsecase
	log *log.Helper
}

func NewAccountService(uc *biz.AccountUsecase, logger log.Logger) *AccountService {
	return &AccountService{
		uc: uc,
		log: log.NewHelper(logger),
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountReply, error) {
	account :=  &biz.Account{
		Email: req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	}
	err := s.uc.CreateAccount(ctx, account)
	if err != nil && errs.IsEmailAlreadyUsed(err) {
		return nil, pb.ErrorEmailAlreadyUsed("email %s has been used", req.Email)
	}
	return &pb.CreateAccountReply{
		Id: account.Id,
	}, err
}

func (s *AccountService) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountReply, error) {
	err := s.uc.UpdateAccount(ctx, &biz.Account{
		Id: req.Id,
		Avatar: req.Avatar,
		Nickname: req.Nickname,
		Password: req.Password,
	})
	return &pb.UpdateAccountReply{}, err
}

func (s *AccountService) DeleteAccount( context.Context,  *pb.DeleteAccountRequest) (*pb.DeleteAccountReply, error) {
	return &pb.DeleteAccountReply{}, nil
}

func (s *AccountService) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountReply, error) {
	account, err := s.uc.GetAccount(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountReply{
		Account: &pb.AccountInfo{
			Id:        account.Id,
			Nickname:  account.Nickname,
			Avatar:    account.Avatar,
			Email:     account.Email,
			Platform:  int32(account.Platform),
			OpenId:    account.OpenId,
			State:     int32(account.State),
			CreatedAt: account.CreatedAt.Unix(),
		},
	}, nil
}
func (s *AccountService) ListAccount( context.Context,  *pb.ListAccountRequest) (*pb.ListAccountReply, error) {
	return &pb.ListAccountReply{}, nil
}
func (s *AccountService) ListWithIds(ctx context.Context, req *pb.ListWithIdsRequest) (*pb.ListWithIdsReply, error) {
	accounts, err := s.uc.GetAccountByIds(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.AccountInfo, len(accounts))
	for i := range accounts {
		result[i] = toAccountInfo(accounts[i])
	}
	return &pb.ListWithIdsReply{Accounts: result}, nil
}
func (s *AccountService) EmailLogin(ctx context.Context, req *pb.EmailLoginRequest) (*pb.AccountLoginReply, error) {
	account := &biz.Account{
		Email: req.Email,
		Password: req.Password,
	}
	err := s.uc.EmailLogin(ctx, account)
	return &pb.AccountLoginReply{
		Account: toAccountInfo(account),
	}, err
}


func toAccountInfo(account *biz.Account) *pb.AccountInfo {
	return &pb.AccountInfo{
		Id:        account.Id,
		Nickname:  account.Nickname,
		Avatar:    account.Avatar,
		Email:     account.Email,
		Platform:  int32(account.Platform),
		OpenId:    account.OpenId,
		State:     int32(account.State),
		CreatedAt: account.CreatedAt.Unix(),
	}
}