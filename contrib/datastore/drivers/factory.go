package driver

import (
	"github.com/micro/go-micro/v2/util/log"
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
	log.Infof("Himanshu Register Backend %+v", backend)
	log.Infof("Himanshu Register Map %+v", driverFactoryMgr)
	if factory, ok := driverFactoryMgr[backend.Type]; ok {
		return factory.CreateDriver(backend)
	}
	return nil, exp.NoSuchType.Error()
}
