// +build wireinject
// The build tag makes sure the stub is not built in the final build.

package main

import (
	"base-service/app/comment/job/internal/biz"
	"base-service/app/comment/job/internal/conf"
	"base-service/app/comment/job/internal/data"
	"base-service/app/comment/job/internal/server"
	"base-service/app/comment/job/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
