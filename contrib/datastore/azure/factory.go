package azure

import (
	"github.com/opensds/multi-cloud/backend/pkg/utils/constants"
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/contrib/datastore/drivers"
)

type AzureDriverFactory struct {
}

func (factory *AzureDriverFactory) CreateDriver(backend *backendpb.BackendDetail) (driver.StorageDriver, error) {
	endpoint := backend.Endpoint
	AccessKeyID := backend.Access
	AccessKeySecret := backend.Security
	ad := AzureAdapter{}
	pipeline, err := ad.createPipeline(endpoint, AccessKeyID, AccessKeySecret)
	if err != nil {
		return nil, err
	}

	adapter := &AzureAdapter{backend: backend, pipeline: pipeline}

	return adapter, nil
}

func init() {
	driver.RegisterDriverFactory(constants.BackendTypeAzureFile, &AzureDriverFactory{})
}
