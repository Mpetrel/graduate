package data

import (
	accountV1 "base-service/api/account/service/v1"
	commentV1 "base-service/api/comment/service/v1"
	"base-service/app/baseapp/interface/internal/conf"
	"context"
	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	consulAPI "github.com/hashicorp/consul/api"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDiscovery,
	NewCommentServiceClient,
	NewAccountServiceClient,
	NewCommentRepo,
	NewAccountRepo,
)

// Data .
type Data struct {
	log *log.Helper
	cc commentV1.CommentClient
	ac accountV1.AccountClient
}

// NewData .
func NewData(
	_ *conf.Data,
	logger log.Logger,
	cc commentV1.CommentClient,
	ac accountV1.AccountClient,) (*Data, func(), error) {
	l := log.NewHelper(logger)
	cleanup := func() {
		l.Infof("closing the data resources\n")
	}
	return &Data{
		log: l,
		cc: cc,
		ac: ac,
	}, cleanup, nil
}


func NewDiscovery(conf *conf.Registry) registry.Discovery {
	c := consulAPI.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli)
	return r
}

func NewCommentServiceClient(conf *conf.Server, r registry.Discovery) commentV1.CommentClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithTimeout(conf.Grpc.Timeout.AsDuration()),
		grpc.WithEndpoint("discovery:///base.comment.service"),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	c := commentV1.NewCommentClient(conn)
	return c
}

func NewAccountServiceClient(conf *conf.Server, r registry.Discovery) accountV1.AccountClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithTimeout(conf.Grpc.Timeout.AsDuration()),
		grpc.WithEndpoint("discovery:///base.account.service"),
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	c := accountV1.NewAccountClient(conn)
	return c
}