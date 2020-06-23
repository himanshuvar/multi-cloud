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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/backend/pkg/utils/constants"
	"github.com/opensds/multi-cloud/contrib/datastore/drivers"
	"github.com/opensds/multi-cloud/contrib/datastore/file/aws"
	"github.com/opensds/multi-cloud/file/pkg/db"
	"github.com/opensds/multi-cloud/file/pkg/model"
	"github.com/opensds/multi-cloud/file/pkg/utils"

	pstruct "github.com/golang/protobuf/ptypes/struct"
	backend "github.com/opensds/multi-cloud/backend/proto"
	driverutils "github.com/opensds/multi-cloud/contrib/utils"
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

func (f *fileService) UpdateFileShareStruct(fsModel *model.FileShare, fs *pb.FileShare) error {
	var tags []*pb.Tag
	for _, tag := range fsModel.Tags {
		tags = append(tags, &pb.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	fs.Tags = tags

	metadata, err := driverutils.ConvertMapToStruct(fsModel.Metadata)
	if err != nil {
		log.Error(err)
		return err
	}
	fs.Metadata = metadata

	return nil
}

func (f *fileService) UpdateFileShareModel(fs *pb.FileShare, fsModel *model.FileShare) error {

	tags, err := f.ConvertTags(fs.Tags)
	if err != nil {
		log.Errorf("Failed to get conversions for tags: %v", fs.Tags, err)
		return err
	}
	fsModel.Tags = tags

	metadata := make(map[string]interface{})
	for k, v := range fsModel.Metadata {
		metadata[k] = v
	}

	metaMap, err := f.ConvertMetadataStructToMap(fs.Metadata)
	if err != nil {
		log.Errorf("Failed to get metaMap: %v", err)
		return err
	}

	for k, v := range metaMap {
		metadata[k] = v
	}
	fsModel.Metadata = metadata

	return nil
}

func (f *fileService) ConvertTags(pbtags []*pb.Tag) ([]model.Tag, error) {
	var tags []model.Tag
	for _, tag := range pbtags {
		tags = append(tags, model.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return tags, nil
}

func (f *fileService) ConvertMetadataStructToMap(metaStruct *pstruct.Struct) (map[string]interface{}, error) {

	var metaMap map[string]interface{}
	metaMap = make(map[string]interface{})
	listFields = make(map[string]*pstruct.Value)

	fields := metaStruct.GetFields()

	metaMap, err := ParseStructFields(fields)
	if err != nil {
		log.Errorf("Failed to get metadata: %v", err)
		return metaMap, err
	}

	if len(listFields) != 0 {
		meta, err := ParseStructFields(listFields)
		if err != nil {
			log.Errorf("Failed to get array for metadata : %v", err)
			return metaMap, err
		}
		for k, v := range meta {
			metaMap[k] = v
		}
	}
	return metaMap, nil
}


func (f *fileService) MergeFileShareData(fs *pb.FileShare, fsFinal *pb.FileShare) error {

	if fs.Tags != nil || len(fs.Tags) != 0{
		fsFinal.Tags = fs.Tags
	}
	var metaMapFinal map[string]interface{}

	metaMapFinal, err := f.ConvertMetadataStructToMap(fsFinal.Metadata)
	if err != nil {
		log.Errorf("Failed to get metaMap: %v", err)
		return err
	}

	metaMap, err := f.ConvertMetadataStructToMap(fs.Metadata)
	if err != nil {
		log.Errorf("Failed to get metaMap: %v", err)
		return err
	}

	for k, v := range metaMap {
		metaMapFinal[k] = v
	}

	fsFinal.Status = fs.Status
	fsFinal.Size = fs.Size
	fsFinal.Encrypted = fs.Encrypted
	fsFinal.EncryptionSettings = fs.EncryptionSettings

	metadataFS, err := driverutils.ConvertMapToStruct(metaMapFinal)
	if err != nil {
		log.Error(err)
		return err
	}
	fsFinal.Metadata = metadataFS

	return nil
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
		fileshare :=  &pb.FileShare{
			Id:                 fs.Id.Hex(),
			CreatedAt:          fs.CreatedAt,
			UpdatedAt:          fs.UpdatedAt,
			Name:               fs.Name,
			Description:        fs.Description,
			TenantId:           fs.TenantId,
			UserId:             fs.UserId,
			BackendId:          fs.BackendId,
			Backend:            fs.Backend,
			Size:               *fs.Size,
			Type:               fs.Type,
			Status:             fs.Status,
			Region:             fs.Region,
			AvailabilityZone:   fs.AvailabilityZone,
			Protocols:          fs.Protocols,
			SnapshotId:         fs.SnapshotId,
			Encrypted:          *fs.Encrypted,
			EncryptionSettings: fs.EncryptionSettings,
		}

		if f.UpdateFileShareStruct(fs, fileshare) != nil {
			log.Errorf("Failed to update fileshare struct: %v\n", fs, err)
			return err
		}

		fileshares = append(fileshares,fileshare)
	}
	out.Fileshares = fileshares
	out.Next = in.Offset + int32(len(res))

	log.Debugf("Listed File Share successfully, #num=%d, fileshares: %+v\n", len(fileshares), fileshares)

	return nil
}

func (f *fileService) GetFileShare(ctx context.Context, in *pb.GetFileShareRequest, out *pb.GetFileShareResponse) error {
	log.Info("Received GetFileShare request.")

	fs, err := db.DbAdapter.GetFileShare(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to get fileshare: %v\n", err)
		return err
	}

	fileshare :=  &pb.FileShare{
		Id:                 fs.Id.Hex(),
		CreatedAt:          fs.CreatedAt,
		UpdatedAt:          fs.UpdatedAt,
		Name:               fs.Name,
		Description:        fs.Description,
		TenantId:           fs.TenantId,
		UserId:             fs.UserId,
		BackendId:          fs.BackendId,
		Backend:            fs.Backend,
		Size:               *fs.Size,
		Type:               fs.Type,
		Status:             fs.Status,
		Region:             fs.Region,
		AvailabilityZone:   fs.AvailabilityZone,
		Protocols:          fs.Protocols,
		SnapshotId:         fs.SnapshotId,
		Encrypted:          *fs.Encrypted,
		EncryptionSettings: fs.EncryptionSettings,
	}

	if f.UpdateFileShareStruct(fs, fileshare) != nil {
		log.Errorf("Failed to update fileshare struct: %v\n", fs, err)
		return err
	}
	log.Debugf("Get File share response, fileshare: %+v\n", fileshare)

	out.Fileshare = fileshare

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
	} else {
		if backend.Backend.Type == constants.BackendTypeAwsFile {
			time.Sleep(4 * time.Second)
		}

		fsGetInput := &pb.GetFileShareRequest{Fileshare:fs.Fileshare}

		getFs, err := sd.GetFileShare(ctx, fsGetInput)
		if err != nil {
			fs.Fileshare.Status = utils.FileShareStateError
			log.Errorf("Received error in getting file shares at backend ", err)
			return err
		}
		log.Debugf("Get File share response= [%+v] from backend", getFs)

		if f.MergeFileShareData(getFs.Fileshare, fs.Fileshare) != nil {
			log.Errorf("Failed to merge file share create data: [%+v] and get data: [%+v] %v\n", fs, err)
			return err
		}
	}

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
		Protocols:          in.Fileshare.Protocols,
		SnapshotId:         in.Fileshare.SnapshotId,
		Size:               &fs.Fileshare.Size,
		Encrypted:          &fs.Fileshare.Encrypted,
		EncryptionSettings: fs.Fileshare.EncryptionSettings,
	}

	if f.UpdateFileShareModel(fs.Fileshare, fileshare) != nil {
		log.Errorf("Failed to update fileshare model: %v\n", fileshare, err)
		return err
	}

	if in.Fileshare.Id != "" {
		fileshare.Id = bson.ObjectIdHex(in.Fileshare.Id)
	}
	log.Debugf("Create File Share Model: %+v", fileshare)

	res, err := db.DbAdapter.CreateFileShare(ctx, fileshare)
	if err != nil {
		log.Errorf("Failed to create file share: %v", err)
		return err
	}

	metadataFS, err := driverutils.ConvertMapToStruct(fileshare.Metadata)
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
		Tags:               in.Fileshare.Tags,       // Need to revisit
		Protocols:          res.Protocols,
		SnapshotId:         res.SnapshotId,
		Encrypted:          *res.Encrypted,
		EncryptionSettings: res.EncryptionSettings,
		Metadata:           metadataFS,
	}

	log.Info("Create file share successfully.")
	return nil
}

func (f *fileService) UpdateFileShare(ctx context.Context, in *pb.UpdateFileShareRequest,
	out *pb.UpdateFileShareResponse) error {

	log.Info("Received UpdateFileShare request.")

	res, err := db.DbAdapter.GetFileShare(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to get fileshare: [%v] from db\n", res, err)
		return err
	}

	backend, err := utils.GetBackend(ctx, f.backendClient, res.BackendId)
	if err != nil {
		log.Errorln("failed to get backend client with err:", err)
		return err
	}

	sd, err := driver.CreateStorageDriver(backend.Backend)
	if err != nil {
		log.Errorln("Failed to create Storage driver err:", err)
		return err
	}

	in.Fileshare.Name = res.Name
	if backend.Backend.Type == constants.BackendTypeAwsFile {
		metaMap, err := f.ConvertMetadataStructToMap(in.Fileshare.Metadata)
		if err != nil {
			log.Errorf("Failed to get metaMap: %v", err)
			return err
		}
		metaMap[aws.FileSystemId] = res.Metadata[aws.FileSystemId]
		metaStruct, err := driverutils.ConvertMapToStruct(metaMap)
		if err != nil {
			log.Errorf("failed to convert metadata map to struct: %v\n", err)
			return err
		}
		in.Fileshare.Metadata = metaStruct
	}

	fs, err := sd.UpdatefileShare(ctx, in)
	if err != nil {
		log.Errorf("Received error in creating file shares at backend ", err)
		fs.Fileshare.Status = utils.FileShareStateError
	} else {
		if backend.Backend.Type == constants.BackendTypeAwsFile {
			time.Sleep(4 * time.Second)
		}

		fsGetInput := &pb.GetFileShareRequest{Fileshare:fs.Fileshare}

		getFs, err := sd.GetFileShare(ctx, fsGetInput)
		if err != nil {
			log.Errorf("Received error in getting file shares at backend ", err)
			return err
		}
		log.Debugf("Get File share response= [%+v] from backend", getFs)

		if f.MergeFileShareData(getFs.Fileshare, fs.Fileshare) != nil {
			log.Errorf("Failed to merge file share create data: [%+v] and get data: [%+v] %v\n", fs, err)
			return err
		}
	}

	fileshare := &model.FileShare{
		Id: 				res.Id,
		Name:               res.Name,
		Description:        in.Fileshare.Description,
		TenantId:           res.TenantId,
		UserId:             res.UserId,
		BackendId:          res.BackendId,
		Backend:            backend.Backend.Name,
		CreatedAt:          res.CreatedAt,
		UpdatedAt:          time.Now().Format(utils.TimeFormat),
		Type:               res.Type,
		Status:             fs.Fileshare.Status,
		Region:             res.Region,
		AvailabilityZone:   res.AvailabilityZone,
		Protocols:          res.Protocols,
		SnapshotId:         res.SnapshotId,
		Size:               &fs.Fileshare.Size,
		Encrypted:          res.Encrypted,
		EncryptionSettings: res.EncryptionSettings,
		Tags: res.Tags,
		Metadata: res.Metadata,
	}

	if f.UpdateFileShareModel(fs.Fileshare, fileshare) != nil {
		log.Errorf("Failed to update fileshare model: %v\n", fileshare, err)
		return err
	}
	log.Debugf("Update File Share Model: %+v", fileshare)

	upateRes, err := db.DbAdapter.UpdateFileShare(ctx, fileshare)
	if err != nil {
		log.Errorf("Failed to update file share: %v", err)
		return err
	}

	out.Fileshare = &pb.FileShare{
		Id:                 upateRes.Id.Hex(),
		CreatedAt:          upateRes.CreatedAt,
		UpdatedAt:          upateRes.UpdatedAt,
		Name:               upateRes.Name,
		Description:        upateRes.Description,
		TenantId:           upateRes.TenantId,
		UserId:             upateRes.UserId,
		BackendId:          upateRes.BackendId,
		Backend:            upateRes.Backend,
		Size:               *upateRes.Size,
		Type:               upateRes.Type,
		Status:             upateRes.Status,
		Region:             upateRes.Region,
		AvailabilityZone:   upateRes.AvailabilityZone,
		Protocols:          res.Protocols,
		SnapshotId:         upateRes.SnapshotId,
		Encrypted:          *upateRes.Encrypted,
		EncryptionSettings: upateRes.EncryptionSettings,
	}

	if f.UpdateFileShareStruct(upateRes, out.Fileshare) != nil {
		log.Errorf("Failed to update fileshare struct: %v\n", fs, err)
		return err
	}

	log.Info("Update file share successfully.")
	return nil
}

func (f *fileService) DeleteFileShare(ctx context.Context, in *pb.DeleteFileShareRequest, out *pb.DeleteFileShareResponse) error {

	log.Info("Received DeleteFileShare request.")

	res, err := db.DbAdapter.GetFileShare(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to get fileshare: [%v] from db\n", res, err)
		return err
	}

	backend, err := utils.GetBackend(ctx, f.backendClient, res.BackendId)
	if err != nil {
		log.Errorln("failed to get backend client with err:", err)
		return err
	}

	sd, err := driver.CreateStorageDriver(backend.Backend)
	if err != nil {
		log.Errorln("Failed to create Storage driver err:", err)
		return err
	}

	fileshare := &pb.FileShare{
		Name: res.Name,
	}

	metadata, err := driverutils.ConvertMapToStruct(res.Metadata)
	if err != nil {
		log.Error(err)
		return err
	}
	fileshare.Metadata = metadata

	in.Fileshare = fileshare

	_, err = sd.DeleteFileShare(ctx, in)
	if err != nil {
		log.Errorf("Received error in deleting file shares at backend ", err)
		return err
	}

	err = db.DbAdapter.DeleteFileShare(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to delete file share: %v\n", err)
		return err
	}
	log.Info("Delete file share successfully.")
	return nil
}
