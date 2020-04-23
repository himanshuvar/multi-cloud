package datastore

import (
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/aws"
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/azure"
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/ceph"
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/gcp"
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/huawei"
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/ibm"
	_ "github.com/sodafoundation/multi-cloud/s3/pkg/datastore/yig"
)
