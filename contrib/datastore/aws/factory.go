package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/opensds/multi-cloud/backend/pkg/utils/constants"
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/contrib/datastore/drivers"
)

type AwsFSDriverFactory struct {
}

func (factory *AwsFSDriverFactory) CreateDriver(backend *backendpb.BackendDetail) (driver.StorageDriver, error) {
	log.Infof("Entered to create driver")

	// Create AWS session with the AWS cloud credentials
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(backend.Region),
		Credentials: credentials.NewStaticCredentials(backend.Access, backend.Security, ""),
	})
	if err != nil {
		log.Errorf("Error in getting the Session")
		return nil, err
	}

	adapter := &AwsAdapter{session: sess}
	return adapter, nil
}

func init() {
	log.Infof("Himanshu Register Factory")
	driver.RegisterDriverFactory(constants.BackendTypeAwsFile, &AwsFSDriverFactory{})
}
