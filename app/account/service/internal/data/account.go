package data

import (
	"base-service/app/account/service/internal/biz"
	"base-service/app/account/service/internal/pkg/errs"
	"base-service/pkg/orm"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"time"
)

type Account struct {
	orm.Model
	Nickname string
	Avatar string
	Email string
	Password string
	Salt string
	Platform int
	OpenId string
	State int8
}

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

func (a accountRepo) CreateAccountByEmail(ctx context.Context, account *biz.Account) error {
	sql := `INSERT INTO accounts ( id, nickname, avatar, email, password, salt, platform, open_id, state, created_at, updated_at ) SELECT
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? FROM DUAL 
			WHERE NOT EXISTS ( SELECT * FROM accounts WHERE email = ? );`
	t := time.Now()
	account.Id = orm.NextId()
	result := a.data.db.WithContext(ctx).Exec(sql, account.Id, account.Nickname, account.Avatar, account.Email, account.Password,
		account.Salt, account.Platform, account.OpenId, account.State, t, t, account.Email)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.EmailAlreadyUsed
	}
	account.CreatedAt = t
	return nil
}

// OauthAccount 第三方登录账户处理
func (a accountRepo) OauthAccount(ctx context.Context, account *biz.Account) error {
	// 检查是否已经存在，如果存在则查询并返回，否则插入一条记录
	var accountModel Account
	result := a.data.db.WithContext(ctx).
		Where(Account{OpenId: account.OpenId}).
		Attrs(Account{
			Model: orm.Model{Id: orm.NextId()},
			Avatar: account.Avatar,
			Nickname: account.Nickname,
			Platform: account.Platform,
		}).
		FirstOrCreate(&accountModel)
	if result.Error != nil {
		return result.Error
	}
	account.Id = accountModel.Id
	account.Avatar = accountModel.Avatar
	account.Nickname = accountModel.Nickname
	account.CreatedAt = accountModel.CreatedAt
	account.UpdatedAt = accountModel.UpdatedAt
	return nil
}

func (a accountRepo) GetAccount(ctx context.Context, email string) (*biz.Account, error) {
	var account Account
	result := a.data.db.WithContext(ctx).Where("email = ?", email).First(&account)
	return toBizAccount(&account), result.Error
}

func (a accountRepo) GetAccountById(ctx context.Context, id uint64) (*biz.Account, error) {
	var account Account
	result := a.data.db.WithContext(ctx).First(&account, id)
	return toBizAccount(&account), result.Error
}


func (a accountRepo) GetAccountByOpenId(ctx context.Context, openId string) (*biz.Account, error)  {
	var account Account
	result := a.data.db.WithContext(ctx).Where("open_id = ?",  openId).First(&account)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, result.Error
	}
	return toBizAccount(&account), nil
}

func (a accountRepo) UpdateAccount(ctx context.Context, account *biz.Account) error {
	accountModel := &Account{}
	accountModel.Id = account.Id
	result := a.data.db.Debug().WithContext(ctx).
		Model(accountModel).
		Updates(Account{Nickname: account.Nickname, Avatar: account.Avatar, Password: account.Password})
	return result.Error
}

func (a accountRepo) GetAccountByIds(ctx context.Context, ids []uint64) ([]*biz.Account, error) {
	var accounts []*Account
	result := a.data.db.WithContext(ctx).Where("id IN ?", ids).
		Find(&accounts)
	if result.Error != nil {
		return nil, result.Error
	}
	ret := make([]*biz.Account, len(accounts))
	for i := range accounts {
		ret[i] = toBizAccount(accounts[i])
	}
	return ret, nil
}

func toBizAccount(account *Account) *biz.Account {
	return &biz.Account{
		Id:        account.Id,
		Nickname:  account.Nickname,
		Avatar:    account.Avatar,
		Email:     account.Email,
		Password:  account.Password,
		Salt:      account.Salt,
		Platform:  account.Platform,
		OpenId:    account.OpenId,
		State:     account.State,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}
