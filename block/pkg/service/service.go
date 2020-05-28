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

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/opensds/multi-cloud/block/pkg/model"
	_ "strings"

	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/block/pkg/db"
	block "github.com/opensds/multi-cloud/block/proto"
	pb "github.com/opensds/multi-cloud/block/proto"
	log "github.com/sirupsen/logrus"
)

type blockService struct {
	blockClient block.BlockService
}

func NewBlockService() pb.BlockHandler {
	return &blockService{
		blockClient: block.NewBlockService("block", client.DefaultClient),
	}
}

func (b *blockService) ListVolume(ctx context.Context, in *pb.ListVolumeRequest, out *pb.ListVolumeResponse) error {
	log.Info("Received ListVolume request.")

	if in.Limit < 0 || in.Offset < 0 {
		msg := fmt.Sprintf("invalid pagination parameter, limit = %d and offset = %d.", in.Limit, in.Offset)
		log.Info(msg)
		return errors.New(msg)
	}

	res, err := db.DbAdapter.ListVolume(ctx, int(in.Limit), int(in.Offset), in.Filter)
	if err != nil {
		log.Errorf("failed to list volumes: %v\n", err)
		return err
	}

	var volumes []*pb.Volume
	for _, item := range res {
		volumes = append(volumes, &pb.Volume{
			Id:                 item.Id.Hex(),
			Name:               item.Name,
			Description:        item.Description,
			TenantId:           item.TenantId,
			UserId:             item.UserId,
			BackendId:          item.BackendId,
			SnapshotId:         item.SnapshotId,
			Size:               item.Size,
			Type:               item.Type,
			Status:             item.Status,
			Region:             item.Region,
			AvailabilityZone:   item.AvailabilityZone,
			MultiAttachEnabled: item.MultiAttach,
			Encrypted:          item.Encrypted,
			Metadata:           item.Metadata,
		})
	}
	out.Volumes = volumes
	out.Next = in.Offset + int32(len(res))

	log.Infof("List volume successfully, #num=%d, volumes: %+v\n", len(volumes), volumes)
	return nil
}

func (b *blockService) GetVolume(ctx context.Context, in *pb.GetVolumeRequest, out *pb.GetVolumeResponse) error {
	log.Info("Received GetVolume request.")
	res, err := db.DbAdapter.GetVolume(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to get volume: %v\n", err)
		return err
	}
	out.Volume = &pb.Volume{
		Id:                 res.Id.Hex(),
		Name:               res.Name,
		Description:        res.Description,
		TenantId:           res.TenantId,
		UserId:             res.UserId,
		BackendId:          res.BackendId,
		SnapshotId:         res.SnapshotId,
		Size:               res.Size,
		Type:               res.Type,
		Status:             res.Status,
		Region:             res.Region,
		AvailabilityZone:   res.AvailabilityZone,
		MultiAttachEnabled: res.MultiAttach,
		Encrypted:          res.Encrypted,
		Metadata:           res.Metadata,
	}
	log.Info("Get volume successfully.")
	return nil
}

func (b *blockService) CreateVolume(ctx context.Context, in *pb.CreateVolumeRequest, out *pb.CreateVolumeResponse) error {
	log.Info("Received CreateVolume request.")
	volume := &model.Volume{
		Name:               in.Volume.Name,
		Description:        in.Volume.Description,
		TenantId:           in.Volume.TenantId,
		UserId:             in.Volume.UserId,
		BackendId:          in.Volume.BackendId,
		SnapshotId:         in.Volume.SnapshotId,
		Size:               in.Volume.Size,
		Type:               in.Volume.Type,
		Status:             in.Volume.Status,
		Region:             in.Volume.Region,
		AvailabilityZone:   in.Volume.AvailabilityZone,
		MultiAttach:        in.Volume.MultiAttachEnabled,
		Encrypted:          in.Volume.Encrypted,
		Metadata:           in.Volume.Metadata,
	}

	res, err := db.DbAdapter.CreateVolume(ctx, volume)
	if err != nil {
		log.Errorf("Failed to create volume: %v", err)
		return err
	}

	out.Volume = &pb.Volume{
		Id:                 res.Id.Hex(),
		Name:               res.Name,
		Description:        res.Description,
		TenantId:           res.TenantId,
		UserId:             res.UserId,
		BackendId:          res.BackendId,
		SnapshotId:         res.SnapshotId,
		Size:               res.Size,
		Type:               res.Type,
		Status:             res.Status,
		Region:             res.Region,
		AvailabilityZone:   res.AvailabilityZone,
		MultiAttachEnabled: res.MultiAttach,
		Encrypted:          res.Encrypted,
		Metadata:           res.Metadata,
	}
	log.Info("Create volume successfully.")
	return nil

}

func (b *blockService) DeleteVolume(ctx  context.Context, in *pb.DeleteVolumeRequest, out *pb.DeleteVolumeResponse) error {
	log.Info("Received DeleteVolume request.")
	err := db.DbAdapter.DeleteVolume(ctx, in.Id)
	if err != nil {
		log.Errorf("failed to delete volume: %v\n", err)
		return err
	}
	log.Info("Delete volume successfully.")
	return nil
}
