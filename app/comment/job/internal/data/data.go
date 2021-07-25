package data

import (
	"base-service/app/comment/job/internal/conf"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCommentRepo)

// Data .
type Data struct {
	db *gorm.DB
	redisDB *redis.Client
	Kafka *Kafka
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logg := log.NewHelper(logger)
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		logg.Errorf("failed opening connection to mysql: %v", err)
		return nil, nil, err
	}
	// redis
	r := redis.NewClient(&redis.Options{
		Addr: c.Redis.Addr,
		Password: "",
		ReadTimeout: c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})
	_, err = r.Ping(context.Background()).Result()
	if err != nil {
		return nil, nil, err
	}
	// kafka
	kafka, err := NewKafka(c.Kafka.Addr, logger)
	if err != nil {
		return nil, nil, err
	}
	d := &Data{db: db,redisDB: r, Kafka: kafka}
	cleanup := func() {
		kafka.close()
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