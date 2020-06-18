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

/*
type AwsAdapter struct {
	session *session.Session
	worker *Worker
	fileshare *pb.FileShare
}

func (ad *AwsAdapter) CreateEFSSession(region, ak, sk string) {
	ad.session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(ak, sk, ""),
	}))
}
*/

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

/*
// Worker will do its Action once every interval, making up for lost time that
// happened during the Action by only waiting the time left in the interval.
type Worker struct {
	Stopped         bool          // A flag determining the state of the worker
	ShutdownChannel chan string   // A channel to communicate to the routine
	Interval        time.Duration // The interval with which to run the Action
	period          time.Duration // The actual period of the wait
}

// NewWorker creates a new worker and instantiates all the data structures required.
func (ad *AwsAdapter) NewAdapterWorker(interval time.Duration) *AwsAdapter {
	return &AwsAdapter{
		session: ad.session,
		worker: &Worker{
			Stopped:         false,
			ShutdownChannel: make(chan string),
			Interval:        interval,
			period:          interval,
		},
		fileshare: &pb.FileShare{},
	}
}

// Run starts the worker and listens for a shutdown call.
func (ad *AwsAdapter) Run(ctx context.Context, fs *pb.GetFileShareRequest) {

	log.Println("Worker Started")

	for {
		select {
		case <-ad.worker.ShutdownChannel:
			ad.worker.ShutdownChannel <- "Down"
			return
		case <-time.After(ad.worker.period):
			break
		}

		started := time.Now()

		getFs, err := ad.GetFileShare(ctx, fs)

		if err != nil {
			log.Errorf("Received error in getting file shares at AWS backend ", err)
		}
		ad.fileshare = getFs.Fileshare

		finished := time.Now()

		duration := finished.Sub(started)
		ad.worker.period = ad.worker.Interval - duration
	}

}

// Shutdown is a graceful shutdown mechanism
func (ad *AwsAdapter) Shutdown() {
	ad.worker.Stopped = true

	ad.worker.ShutdownChannel <- "Down"
	<-ad.worker.ShutdownChannel

	close(ad.worker.ShutdownChannel)
}
*/

func (ad *AwsAdapter) ParseFileShare(fsDesc *efs.FileSystemDescription) (*pb.FileShare, error) {
	var tags []*pb.Tag
	for _, tag := range fsDesc.Tags {
		tags = append(tags, &pb.Tag{
			Key:   *tag.Key,
			Value: *tag.Value,
		})
	}

	meta := map[string]interface{}{
		"Name": *fsDesc.Name,
		"FileSystemId": *fsDesc.FileSystemId,
		"OwnerId": *fsDesc.OwnerId,
		"FileSystemSize":  *fsDesc.SizeInBytes,
		"ThroughputMode": *fsDesc.ThroughputMode,
		"PerformanceMode": *fsDesc.PerformanceMode,
		"CreationToken": *fsDesc.CreationToken,
		"CreationTimeAtBackend": *fsDesc.CreationTime,
		"LifeCycleState": *fsDesc.LifeCycleState,
		"NumberOfMountTargets": *fsDesc.NumberOfMountTargets,
	}

	if *fsDesc.ThroughputMode == efs.ThroughputModeProvisioned {
		meta["ProvisionedThroughputInMibps"] = *fsDesc.ProvisionedThroughputInMibps
	}

	metadata, err := ToStruct(meta)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	fileshare := &pb.FileShare{
		Size:                 *fsDesc.SizeInBytes.Value,
		Encrypted:            *fsDesc.Encrypted,
		Status:               *fsDesc.LifeCycleState,
		Tags:                 tags,
		Metadata:             metadata,
	}

	if *fsDesc.Encrypted {
		fileshare.EncryptionSettings = map[string]string {
			"KmsKeyId": *fsDesc.KmsKeyId,
		}
	}

	return fileshare, nil
}

func (ad *AwsAdapter) DescribeFileShare(input *efs.DescribeFileSystemsInput) (*efs.DescribeFileSystemsOutput, error) {
	// Create a EFS client from just a session.
	svc := efs.New(ad.session)

	result, err := svc.DescribeFileSystems(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case efs.ErrCodeBadRequest:
				log.Errorf(efs.ErrCodeBadRequest, aerr.Error())
			case efs.ErrCodeInternalServerError:
				log.Errorf(efs.ErrCodeInternalServerError, aerr.Error())
			case efs.ErrCodeFileSystemNotFound:
				log.Errorf(efs.ErrCodeFileSystemNotFound, aerr.Error())
			default:
				log.Errorf(aerr.Error())
			}
		} else {
			log.Error(err)
		}
		return nil, err
	}
	return result, nil
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

	fileShare, err := ad.ParseFileShare(result)

	if err != nil {
		log.Error(err)
		return nil, err
	}
/*
	worker := ad.NewAdapterWorker(2 * time.Second)
	fsInput := &pb.GetFileShareRequest{Fileshare:fileShare}
	go worker.Run(ctx, fsInput)
	time.Sleep(2 * time.Second)
	worker.Shutdown()
	time.Sleep(1 * time.Second)

	return &pb.CreateFileShareResponse{
		Fileshare: worker.fileshare,
	}, nil
	*/

	return &pb.CreateFileShareResponse{
		Fileshare: fileShare,
	}, nil
}

func (ad *AwsAdapter) GetFileShare(ctx context.Context, fs *pb.GetFileShareRequest) (*pb.GetFileShareResponse, error) {

	input := &efs.DescribeFileSystemsInput{
		FileSystemId: aws.String(fs.Fileshare.Metadata.Fields["FileSystemId"].GetStringValue()),
	}

	result, err := ad.DescribeFileShare(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Infof("Get File share response = %+v", result)

	fileShare, err := ad.ParseFileShare(result.FileSystems[0])

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &pb.GetFileShareResponse{
		Fileshare: fileShare,
	}, nil
}

func (ad *AwsAdapter) ListFileShare(ctx context.Context, in *pb.ListFileShareRequest) (*pb.ListFileShareResponse, error) {

	input := &efs.DescribeFileSystemsInput{}

	result, err := ad.DescribeFileShare(input)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Infof("Get File share response = %+v", result)

	var fileshares []*pb.FileShare
	for _, fileshare := range result.FileSystems {
		fs, err := ad.ParseFileShare(fileshare)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		fileshares = append(fileshares, fs)
	}

	return &pb.ListFileShareResponse{
		Fileshares: fileshares,
	}, nil
}


func (ad *AwsAdapter) Close() error {
	// TODO:
	return nil
}
