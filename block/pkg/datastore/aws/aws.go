// Copyright 2020 The OpenSDS Authors.
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
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	dscommon "github.com/opensds/multi-cloud/block/pkg/datastore/common"
	pb "github.com/opensds/multi-cloud/block/proto"
	. "github.com/opensds/multi-cloud/s3/error"
	log "github.com/sirupsen/logrus"
)

type AwsAdapter struct {
	session *session.Session
}

func (ad *AwsAdapter) List(ctx context.Context) (*pb.ListVolumeResponse, error) {
	getVolumesInput := &awsec2.DescribeVolumesInput{}
	var result pb.ListVolumeResponse
	log.Infof("Getting volumes list from AWS EC2 service")

	// Create a session of EC2
	svc := awsec2.New(ad.session)

	// Get the Volumes
	volResponse, err := svc.DescribeVolumes(getVolumesInput)
	if err != nil {
		log.Errorf("Errror in getting volumes list, err:%v", err)
		return nil, ErrGetFromBackendFailed
	}
	log.Infof("Describe volumes from AWS succeeded")
	for _, vol := range volResponse.Volumes {
		result.Volumes = append(result.Volumes, &pb.Volume{
			Id: *vol.VolumeId,
			// Always report size in Bytes
			Size:      (*vol.Size) * (dscommon.GB_FACTOR),
			Type:      *vol.VolumeType,
			Status:    *vol.State,
			Encrypted: *vol.Encrypted,
		})
	}
	log.Infof("Successfully got the volumes list")
	return &result, nil
}

func (ad *AwsAdapter) CreateVolume(ctx context.Context, volume *pb.Volume) (*pb.CreateVolumeResponse, error) {
	createVolumeInput := &awsec2.CreateVolumeInput{
		AvailabilityZone: aws.String(volume.AvailabilityZone),
		Size:             aws.Int64(volume.Size),
		VolumeType:       aws.String(volume.Type),
	}

	var result pb.CreateVolumeResponse
	log.Infof("Creating volumes list from AWS EC2 service")

	// Create a session of EC2
	svc := awsec2.New(ad.session)

	// Create the Volume
	volResponse, err := svc.CreateVolume(createVolumeInput)
	if err != nil {
		log.Errorf("Error in creating volumes list, err:%v", err)
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Errorf(aerr.Error())
			}
		} else {
			log.Errorf(err.Error())
		}
		return nil, err
	}
	log.Infof("Create Volume is successful")

	result.Volume = &pb.Volume{
		SnapshotId:           *volResponse.SnapshotId,
		Size:                 *volResponse.Size,
		Type:                 *volResponse.VolumeType,
		AvailabilityZone:     *volResponse.AvailabilityZone,
		Status:               *volResponse.State,
		Encrypted:            *volResponse.Encrypted,
		MultiAttachEnabled:   *volResponse.MultiAttachEnabled,
		Metadata:             map[string]string{
			                      "volumeId": *volResponse.VolumeId},
	}

	log.Infof("Successfully got the volumes list")
	return &result, nil
}
