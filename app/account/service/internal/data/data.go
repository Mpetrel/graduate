package data

import (
	"base-service/app/account/service/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAccountRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logg := log.NewHelper(logger)
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		logg.Errorf("failed opening connection to mysql: %v", err)
		return nil, nil, err
	}
	d := &Data{db: db}
	cleanup := func() {
		logg.Infof("comment service data clean up")
	}
	return d, cleanup, nil
}


func getOffset(page, size int) int {
	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	return offset
}