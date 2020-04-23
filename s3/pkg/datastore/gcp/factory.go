package gcp

import (
	"github.com/sodafoundation/multi-cloud/backend/pkg/utils/constants"
	backendpb "github.com/sodafoundation/multi-cloud/backend/proto"
	"github.com/sodafoundation/multi-cloud/s3/pkg/datastore/driver"
	"github.com/webrtcn/s3client"
)

type GcsDriverFactory struct {
}

func (cdf *GcsDriverFactory) CreateDriver(backend *backendpb.BackendDetail) (driver.StorageDriver, error) {
	endpoint := backend.Endpoint
	AccessKeyID := backend.Access
	AccessKeySecret := backend.Security
	sess := s3client.NewClient(endpoint, AccessKeyID, AccessKeySecret)
	adap := &GcsAdapter{backend: backend, session: sess}

	return adap, nil
}

func init() {
	driver.RegisterDriverFactory(constants.BackendTypeGcs, &GcsDriverFactory{})
}
