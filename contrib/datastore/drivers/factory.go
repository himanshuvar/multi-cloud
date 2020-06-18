package driver

import (
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	exp "github.com/opensds/multi-cloud/s3/pkg/exception"
)

type DriverFactory interface {
	CreateDriver(backend *backendpb.BackendDetail) (StorageDriver, error)
}

var driverFactoryMgr = make(map[string]DriverFactory)

func RegisterDriverFactory(driverType string, factory DriverFactory) {
	driverFactoryMgr[driverType] = factory
}

func CreateStorageDriver(backend *backendpb.BackendDetail) (StorageDriver, error) {
	if factory, ok := driverFactoryMgr[backend.Type]; ok {
		return factory.CreateDriver(backend)
	}
	return nil, exp.NoSuchType.Error()
}
