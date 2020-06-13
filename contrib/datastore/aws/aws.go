// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/golang/protobuf/jsonpb"
	"github.com/opensds/multi-cloud/contrib/utils"

	pstruct "github.com/golang/protobuf/ptypes/struct"
	pb "github.com/opensds/multi-cloud/file/proto"
	log "github.com/sirupsen/logrus"
)

type AwsAdapter struct {
	session *session.Session
}

func ToStruct(msg map[string]interface{}) (*pstruct.Struct, error) {

	byteArray, err := json.Marshal(msg)

	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(byteArray)

	pbs := &pstruct.Struct{}
	if err = jsonpb.Unmarshal(reader, pbs); err != nil {
		return nil, err
	}

	return pbs, nil
}

func (ad *AwsAdapter) CreateFileShare(ctx context.Context, fs *pb.CreateFileShareRequest) (*pb.CreateFileShareResponse, error) {
	// Create a EFS client from just a session.
	svc := efs.New(ad.session)

	var tags []*efs.Tag
	for _, tag := range fs.Fileshare.Tags {
		tags = append(tags, &efs.Tag{
			Key:   aws.String(tag.Key),
			Value: aws.String(tag.Value),
		})
	}

	creationToken := utils.RandString(36)

	input := &efs.CreateFileSystemInput{
	    CreationToken:   aws.String(creationToken),
		Encrypted: aws.Bool(fs.Fileshare.Encrypted),
		PerformanceMode: aws.String(fs.Fileshare.Metadata.Fields["PerformanceMode"].GetStringValue()),
		Tags: tags,
		ThroughputMode: aws.String(fs.Fileshare.Metadata.Fields["ThroughputMode"].GetStringValue()),
	}

	if *input.ThroughputMode == efs.ThroughputModeProvisioned {
		input.ProvisionedThroughputInMibps = aws.Float64(fs.Fileshare.Metadata.Fields["ProvisionedThroughputInMibps"].GetNumberValue())
	}

	if *input.Encrypted {
		input.KmsKeyId = aws.String(fs.Fileshare.EncryptionSettings["KmsKeyId"])
	}

	result, err := svc.CreateFileSystem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case efs.ErrCodeBadRequest:
				log.Errorf(efs.ErrCodeBadRequest, aerr.Error())
			case efs.ErrCodeInternalServerError:
				log.Errorf(efs.ErrCodeInternalServerError, aerr.Error())
			case efs.ErrCodeFileSystemAlreadyExists:
				log.Errorf(efs.ErrCodeFileSystemAlreadyExists, aerr.Error())
			case efs.ErrCodeFileSystemLimitExceeded:
				log.Errorf(efs.ErrCodeFileSystemLimitExceeded, aerr.Error())
			case efs.ErrCodeInsufficientThroughputCapacity:
				log.Errorf(efs.ErrCodeInsufficientThroughputCapacity, aerr.Error())
			case efs.ErrCodeThroughputLimitExceeded:
				log.Errorf(efs.ErrCodeThroughputLimitExceeded, aerr.Error())
			default:
				log.Errorf(aerr.Error())
			}
		} else {
			log.Error(err)
		}
		return nil, err
	}

	log.Infof("Create File share response = %+v", result)

	myVar := map[string]interface{}{
		"Name": *result.Name,
		"FileSystemId": *result.FileSystemId,
		"OwnerId": *result.OwnerId,
		"FileSystemSize":  *result.SizeInBytes,
		"ThroughputMode": *result.ThroughputMode,
		"PerformanceMode": *result.PerformanceMode,
		"CreationToken": *result.CreationToken,
		"CreationTimeAtBackend": *result.CreationTime,
		"LifeCycleState": *result.LifeCycleState,
		"NumberOfMountTargets": *result.NumberOfMountTargets,
	}

	if *result.ThroughputMode == efs.ThroughputModeProvisioned {
		myVar["ProvisionedThroughputInMibps"] = *result.ProvisionedThroughputInMibps
	}

	metadata, err := ToStruct(myVar)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	fileshare := &pb.FileShare{
		Size:                 *result.SizeInBytes.Value,
		Encrypted:            *result.Encrypted,
		Status:               *result.LifeCycleState,
		Metadata: metadata,
	}

	if *result.Encrypted {
		fileshare.EncryptionSettings = map[string]string {
				"KmsKeyId": *result.KmsKeyId,
		}
	}

	return &pb.CreateFileShareResponse{
		Fileshare: fileshare,
	}, nil
}


func (ad *AwsAdapter) CreateEFSSession(region, ak, sk string) {
	ad.session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(ak, sk, ""),
	}))
}

func (ad *AwsAdapter) Close() error {
	// TODO:
	return nil
}
