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

package block

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/api/pkg/policy"
	"github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/block/proto"

	log "github.com/sirupsen/logrus"
)

const (
	blockService   = "block"
	backendService = "backend"
)

type APIService struct {
	blockClient   block.BlockService
	backendClient backend.BackendService
}

func NewAPIService(c client.Client) *APIService {
	return &APIService{
		blockClient:   block.NewBlockService(blockService, c),
		backendClient: backend.NewBackendService(backendService, c),
	}
}

func UpdateFilter(reqFilter map[string]string, filter map[string]string) error {
	for k, v := range filter {
		reqFilter[k] = v
	}
	return nil
}

func WriteError(response *restful.Response, msg string, err error) {
	log.Errorf(msg, err)
	error := response.WriteError(http.StatusInternalServerError, err)
	if error != nil {
		log.Errorf("response write status and http error failed: %v\n", err)
		return
	}
}

func (s *APIService) ListVolume(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "volume:list") {
		return
	}
	log.Info("received request for volume list.")

	ctx := common.InitCtxWithAuthInfo(request)
	backendId := request.PathParameter(common.REQUEST_PATH_BACKEND_ID)
	if backendId != "" {
		s.listVolumeByBackend(ctx, request, response, backendId)
	} else {
		s.listVolumeDefault(ctx, request, response)
	}
}

func (s *APIService) prepareVolumeRequest(request *restful.Request, response *restful.Response) *block.ListVolumeRequest {

	limit, offset, err := common.GetPaginationParam(request)
	if err != nil {
		WriteError(response, "get pagination parameters failed: %v\n", err)
		return nil
	}

	sortKeys, sortDirs, err := common.GetSortParam(request)
	if err != nil {
		WriteError(response, "get sort parameters failed: %v\n", err)
		return nil
	}

	filterOpts := []string{"name", "type", "status"}
	filter, err := common.GetFilter(request, filterOpts)
	if err != nil {
		WriteError(response, "get filter failed: %v\n", err)
		return nil
	}

	return &block.ListVolumeRequest{
		Limit:    limit,
		Offset:   offset,
		SortKeys: sortKeys,
		SortDirs: sortDirs,
		Filter:   filter,
	}
}

func (s *APIService) listVolumeDefault(ctx context.Context, request *restful.Request, response *restful.Response) {
	listVolumeRequest := s.prepareVolumeRequest(request, response)

	res, err := s.blockClient.ListVolume(ctx, listVolumeRequest)
	if err != nil {
		WriteError(response, "list volumes failed: %v\n", err)
		return
	}

	log.Info("list volumes successfully.")

	err = response.WriteEntity(res)
	if err != nil {
		log.Errorf("response write entity failed: %v\n", err)
		return
	}
}

func (s *APIService) listVolumeByBackend(ctx context.Context, request *restful.Request, response *restful.Response,
	backendId string) {

	backendResp, err := s.backendClient.GetBackend(ctx, &backend.GetBackendRequest{Id: backendId})
	if err != nil {
		WriteError(response, "get backend details failed: %v\n", err)
		return
	}
	log.Infof("backend response = [%+v]\n", backendResp)

	filterPathOpts := []string{"backendId"}
	filterPath, err := common.GetFilterPathParams(request, filterPathOpts)
	if err != nil {
		WriteError(response, "get filter failed: %v\n", err)
		return
	}

	listVolumeRequest := s.prepareVolumeRequest(request, response)

	err = UpdateFilter(listVolumeRequest.Filter, filterPath)
	if err != nil {
		log.Errorf("update filter failed: %v\n", err)
		return
	}

	res, err := s.blockClient.ListVolume(ctx, listVolumeRequest)
	if err != nil {
		WriteError(response, "list volumes failed: %v\n", err)
		return
	}

	log.Info("list volumes successfully.")

	err = response.WriteEntity(res)
	if err != nil {
		log.Errorf("response write entity failed: %v\n", err)
		return
	}
}

func (s *APIService) GetVolume(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "volume:get") {
		return
	}
	log.Infof("Received request for volume details: %s\n", request.PathParameter("id"))
	id := request.PathParameter("id")

	ctx := common.InitCtxWithAuthInfo(request)
	res, err := s.blockClient.GetVolume(ctx, &block.GetVolumeRequest{Id: id})
	if err != nil {
		log.Errorf("failed to get volume details: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Info("Get backend details successfully.")
	response.WriteEntity(res.Volume)
}

func (s *APIService) DeleteVolume(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "volume:delete") {
		return
	}

	id := request.PathParameter("id")
	log.Infof("Received request for deleting volume: %s\n", id)

	ctx := common.InitCtxWithAuthInfo(request)
	res, err := s.blockClient.DeleteVolume(ctx, &block.DeleteVolumeRequest{Id: id})
	if err != nil {
		log.Errorf("failed to delete volume: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Info("Delete volume successfully.")
	response.WriteEntity(res)
}
