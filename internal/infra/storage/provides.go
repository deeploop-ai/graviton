package storage

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewMinioObjectStore,
)
