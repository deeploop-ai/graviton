//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx"
	"github.com/lynx-go/lynx/boot"
)

func wireBootstrap(app lynx.Lynx) (*boot.Bootstrap, func(), error) {
	panic(wire.Build(ProviderSet))
}
