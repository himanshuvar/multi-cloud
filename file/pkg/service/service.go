// Copyright 2020 The SODA Authors.
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

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/contrib/datastore/drivers"
	"github.com/opensds/multi-cloud/file/pkg/db"
	"github.com/opensds/multi-cloud/file/pkg/model"
	"github.com/opensds/multi-cloud/file/pkg/utils"

	pstruct "github.com/golang/protobuf/ptypes/struct"
	backend "github.com/opensds/multi-cloud/backend/proto"
	pb "github.com/opensds/multi-cloud/file/proto"
	log "github.com/sirupsen/logrus"
)

var listFields map[string]*pstruct.Value

type fileService struct {
	fileClient    pb.FileService
	backendClient backend.BackendService
}

func NewFileService() pb.FileHandler {

	log.Infof("Init file service finished.\n")
	return &fileService{
		fileClient:    pb.NewFileService("file", client.DefaultClient),
		backendClient: backend.NewBackendService("backend", client.DefaultClient),
	}
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

func ParseStructFields(fields map[string]*pstruct.Value) (map[string]interface{}, error) {
	log.Infof("Parsing struct fields = [%+v]", fields)

	valuesMap := make(map[string]interface{})

	for key, value := range fields {
		if v, ok := value.GetKind().(*pstruct.Value_NullValue); ok {
			valuesMap[key] = v.NullValue
		} else if v, ok := value.GetKind().(*pstruct.Value_NumberValue); ok {
			valuesMap[key] = v.NumberValue
		} else if v, ok := value.GetKind().(*pstruct.Value_StringValue); ok {
			val := strings.Trim(v.StringValue, "\"")
			if key == utils.AZURE_FILESHARE_USAGE_BYTES || key == utils.AZURE_X_MS_SHARE_QUOTA || key == utils.CONTENT_LENGTH {
				valInt, err :=  strconv.ParseInt(val, 10, 64)
				if err != nil {
					log.Errorf("Failed to parse string Fields = ", key)
					return nil, err
				}
				valuesMap[key] = valInt
			} else {
				valuesMap[key] = val
			}
		} else if v, ok := value.GetKind().(*pstruct.Value_BoolValue); ok {
			valuesMap[key] = v.BoolValue
		} else if v, ok := value.GetKind().(*pstruct.Value_StructValue); ok {
			var err error
			valuesMap[key], err = ParseStructFields(v.StructValue.Fields)
			if err != nil {
				log.Errorf("Failed to parse struct Fields = [%+v]", v.StructValue.Fields)
				return nil, err
			}
		} else if v, ok := value.GetKind().(*pstruct.Value_ListValue); ok {
			listFields[key] = v.ListValue.Values[0]
		} else {
			msg := fmt.Sprintf("Failed to parse field for key = [%+v], value = [%+v]", key, value)
			err := errors.New(msg)
			log.Errorf(msg)
			return nil, err
		}
	}
	return valuesMap, nil
}

func (f *fileService) ListFileShare(ctx context.Context, in *pb.ListFileShareRequest, out *pb.ListFileShareResponse) error {
	log.Info("Received ListFileShare request.")

	if in.Limit < 0 || in.Offset < 0 {
		msg := fmt.Sprintf("Invalid pagination parameter, limit = %d and offset = %d.", in.Limit, in.Offset)
		log.Info(msg)
		return errors.New(msg)
	}

	res, err := db.DbAdapter.ListFileShare(ctx, int(in.Limit), int(in.Offset), in.Filter)
	if err != nil {
		log.Errorf("Failed to List FileShares: %v\n", err)
		return err
	}

	var fileshares []*pb.FileShare
	for _, fs := range res {
		var tags []*pb.Tag
		for _, tag := range fs.Tags {
			tags = append(tags, &pb.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
		metadata, err := ToStruct(fs.Metadata)
		if err != nil {
			log.Error(err)
			return err
		}
		fileshares = append(fileshares, &pb.FileShare{
			Id:               fs.Id.Hex(),
			CreatedAt:        fs.CreatedAt,
			UpdatedAt:        fs.UpdatedAt,
			Name:             fs.Name,
			Description:      fs.Description,
			TenantId:         fs.TenantId,
			UserId:           fs.UserId,
			BackendId:        fs.BackendId,
			Backend:          fs.Backend,
			Size:             *fs.Size,
			Type:             fs.Type,
			Status:           fs.Status,
			Region:           fs.Region,
			AvailabilityZone: fs.AvailabilityZone,
			Tags:             tags,
			Protocols:        fs.Protocols,
			SnapshotId:       fs.SnapshotId,
			Encrypted:        *fs.Encrypted,
			EncryptionSettings: fs.EncryptionSettings,
			Metadata:         metadata,
		})
	}
	out.Fileshares = fileshares
	out.Next = in.Offset + int32(len(res))

	log.Infof("List File Share successfully, #num=%d, fileshares: %+v\n", len(fileshares), fileshares)
	return nil
}

func (f *fileService) GetFileShare(ctx context.Context, in *pb.GetFileShareRequest, out *pb.GetFileShareResponse) error {
	log.Info("Received GetFileShare request.")

	fs, err := db.DbAdapter.GetFileShare(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to get fileshare: %v\n", err)
		return err
	}
	var tags []*pb.Tag
	for _, tag := range fs.Tags {
		tags = append(tags, &pb.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	metadata, err := ToStruct(fs.Metadata)
	if err != nil {
		log.Error(err)
		return err
	}

	out.Fileshare = &pb.FileShare{
		Id:               fs.Id.Hex(),
		CreatedAt:        fs.CreatedAt,
		UpdatedAt:        fs.UpdatedAt,
		Name:             fs.Name,
		Description:      fs.Description,
		TenantId:         fs.TenantId,
		UserId:           fs.UserId,
		BackendId:        fs.BackendId,
		Backend:          fs.Backend,
		Size:             *fs.Size,
		Type:             fs.Type,
		Status:           fs.Status,
		Region:           fs.Region,
		AvailabilityZone: fs.AvailabilityZone,
		Tags:             tags,
		Protocols:        fs.Protocols,
		SnapshotId:       fs.SnapshotId,
		Encrypted:        *fs.Encrypted,
		EncryptionSettings: fs.EncryptionSettings,
		Metadata:         metadata,
	}
	log.Info("Get file share successfully.")
	return nil
}

func (f *fileService) CreateFileShare(ctx context.Context, in *pb.CreateFileShareRequest, out *pb.CreateFileShareResponse) error {
	log.Info("Received CreateFileShare request.")

	backend, err := utils.GetBackend(ctx, f.backendClient, in.Fileshare.BackendId)
	if err != nil {
		log.Errorln("failed to get backend client with err:", err)
		return err
	}
	sd, err := driver.CreateStorageDriver(backend.Backend)
	if err != nil {
		log.Errorln("Failed to create Storage driver err:", err)
		return err
	}

	fs, err := sd.CreateFileShare(ctx, in)
	if err != nil {
		log.Errorf("Received error in creating file shares at backend ", err)
		fs.Fileshare.Status = utils.FileShareStateError
	} else{
		fs.Fileshare.Status = utils.FileShareStateCreating
	}

	var tags []model.Tag
	for _, tag := range in.Fileshare.Tags {
		tags = append(tags, model.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	listFields = make(map[string]*pstruct.Value)

	fields := fs.Fileshare.Metadata.GetFields()

	metadata, err := ParseStructFields(fields)
	if err != nil {
		log.Errorf("Failed to get metadata: %v", err)
		return err
	}
	if len(listFields) != 0 {
		meta, err := ParseStructFields(listFields)
		if err != nil {
			log.Errorf("Failed to get array for metadata : %v", err)
			return err
		}
		for k, v := range meta {
			metadata[k] = v
		}
	}

	log.Infof("FS Model Metadata: %+v", metadata)

	fileshare := &model.FileShare{
		Name:               in.Fileshare.Name,
		Description:        in.Fileshare.Description,
		TenantId:           in.Fileshare.TenantId,
		UserId:             in.Fileshare.UserId,
		BackendId:          in.Fileshare.BackendId,
		Backend:            backend.Backend.Name,
		CreatedAt:          time.Now().Format(utils.TimeFormat),
		UpdatedAt:          time.Now().Format(utils.TimeFormat),
		Type:               in.Fileshare.Type,
		Status:             fs.Fileshare.Status,
		Region:             in.Fileshare.Region,
		AvailabilityZone:   in.Fileshare.AvailabilityZone,
		Tags:               tags,
		Protocols:          in.Fileshare.Protocols,
		SnapshotId:         in.Fileshare.SnapshotId,
		Size:               &fs.Fileshare.Size,
		Encrypted:          &in.Fileshare.Encrypted,
		EncryptionSettings: fs.Fileshare.EncryptionSettings,
		Metadata:           metadata,
	}

	log.Infof("FS Model: %+v", fileshare)

	res, err := db.DbAdapter.CreateFileShare(ctx, fileshare)
	if err != nil {
		log.Errorf("Failed to create file share: %v", err)
		return err
	}

	metadataFS, err := ToStruct(metadata)
	if err != nil {
		log.Error(err)
		return err
	}

	out.Fileshare = &pb.FileShare{
		Id:                 res.Id.Hex(),
		CreatedAt:          res.CreatedAt,
		UpdatedAt:          res.UpdatedAt,
		Name:               res.Name,
		Description:        res.Description,
		TenantId:           res.TenantId,
		UserId:             res.UserId,
		BackendId:          res.BackendId,
		Backend:            res.Backend,
		Size:               *res.Size,
		Type:               res.Type,
		Status:             res.Status,
		Region:             res.Region,
		AvailabilityZone:   res.AvailabilityZone,
		Tags:               in.Fileshare.Tags,
		Protocols:          res.Protocols,
		SnapshotId:         res.SnapshotId,
		Encrypted:          *res.Encrypted,
		EncryptionSettings: res.EncryptionSettings,
		Metadata:           metadataFS,
	}

	//time.Sleep(4 * time.Second)
	fsInput := &pb.GetFileShareRequest{Fileshare:out.Fileshare}

	getFs, err := sd.GetFileShare(ctx, fsInput)
	if err != nil {
		log.Errorf("Received error in getting file shares at backend ", err)
		return err
	}

	log.Infof("Get File share response = [%+v]", getFs)

	listFields = make(map[string]*pstruct.Value)

	fields = getFs.Fileshare.Metadata.GetFields()

	metadata, err = ParseStructFields(fields)
	if err != nil {
		log.Errorf("Failed to get metadata: %v", err)
		return err
	}

	if len(listFields) != 0 {
		meta, err := ParseStructFields(listFields)
		if err != nil {
			log.Errorf("Failed to get array for metadata : %v", err)
			return err
		}
		for k, v := range meta {
			metadata[k] = v
		}
	}

	log.Infof("For Update FS Model Metadata: %+v", metadata)

	fileshare.Id = res.Id
	fileshare.Status = getFs.Fileshare.Status
	fileshare.Size = &getFs.Fileshare.Size
	fileshare.Encrypted = &getFs.Fileshare.Encrypted
	fileshare.EncryptionSettings = getFs.Fileshare.EncryptionSettings
	fileshare.Metadata = metadata
	fileshare.UpdatedAt = time.Now().Format(utils.TimeFormat)

	log.Infof("For Update FS Model: %+v", fileshare)

	res, err = db.DbAdapter.UpdateFileShare(ctx, fileshare)
	if err != nil {
		log.Errorf("Failed to update file share: %v", err)
		return err
	}

	log.Info("Create file share successfully.")
	return nil
}

func (f *fileService) DeleteFileShare(ctx context.Context, in *pb.DeleteFileShareRequest, out *pb.DeleteFileShareResponse) error {
	log.Info("Received DeleteFileShare request.")
	err := db.DbAdapter.DeleteFileShare(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to delete file share: %v\n", err)
		return err
	}
	log.Info("Delete file share successfully.")
	return nil
}
