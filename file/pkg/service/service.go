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

	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/file/pkg/db"
	pb "github.com/opensds/multi-cloud/file/proto"
	log "github.com/sirupsen/logrus"
)

type fileService struct {
	fileClient pb.FileService
}

func NewFileService() pb.FileHandler {

	log.Infof("Init file service finished.\n")
	return &fileService{
		fileClient: pb.NewFileService("file", client.DefaultClient),
	}
}

func (f *fileService) ListFileShare(ctx context.Context, in *pb.ListFileShareRequest, out *pb.ListFileShareResponse) error {
	log.Info("Received ListFileShare request.")

	if in.Limit < 0 || in.Offset < 0 {
		msg := fmt.Sprintf("invalid pagination parameter, limit = %d and offset = %d.", in.Limit, in.Offset)
		log.Info(msg)
		return errors.New(msg)
	}

	res, err := db.DbAdapter.ListFileShare(ctx, int(in.Limit), int(in.Offset), in.Filter)
	if err != nil {
		log.Errorf("failed to list fileshares: %v\n", err)
		return err
	}

	var fileshares []*pb.FileShare
	for _, item := range res {
		fileshares = append(fileshares, &pb.FileShare{
			Id:               item.Id.Hex(),
			Name:             item.Name,
			Description:      item.Description,
			TenantId:         item.TenantId,
			UserId:           item.UserId,
			BackendId:        item.BackendId,
			SnapshotId:       item.SnapshotId,
			Size:             item.Size,
			Type:             item.Type,
			Status:           item.Status,
			Region:           item.Region,
			AvailabilityZone: item.AvailabilityZone,
			Encrypted:        item.Encrypted,
			Metadata:         item.Metadata,
		})
	}
	out.Fileshares = fileshares
	out.Next = in.Offset + int32(len(res))

	log.Infof("List file share successfully, #num=%d, fileshares: %+v\n", len(fileshares), fileshares)
	return nil
}
