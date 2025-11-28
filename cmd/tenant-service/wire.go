//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gaoyong06/middleground/tenant-service/internal/biz"
	"github.com/gaoyong06/middleground/tenant-service/internal/conf"
	"github.com/gaoyong06/middleground/tenant-service/internal/data"
	"github.com/gaoyong06/middleground/tenant-service/internal/server"
	"github.com/gaoyong06/middleground/tenant-service/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
