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

package file

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/api/pkg/policy"
	"github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/file/proto"

	log "github.com/sirupsen/logrus"
)

const (
	fileService   = "file"
	backendService = "backend"
)

type APIService struct {
	fileClient   file.FileService
	backendClient backend.BackendService
}

func NewAPIService(c client.Client) *APIService {
	return &APIService{
		fileClient:   file.NewFileService(fileService, c),
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

func (s *APIService) checkBackendExists(ctx context.Context, request *restful.Request, response *restful.Response,
	backendId string){
	backendResp, err := s.backendClient.GetBackend(ctx, &backend.GetBackendRequest{Id: backendId})
	if err != nil {
		WriteError(response, "get backend details failed: %v\n", err)
		return
	}
	log.Infof("backend response = [%+v]\n", backendResp)
}

func (s *APIService) getBackendDetails(ctx context.Context, request *restful.Request, response *restful.Response,
	backendId string) (*backend.GetBackendResponse, error){
	backendResp, err := s.backendClient.GetBackend(ctx, &backend.GetBackendRequest{Id: backendId})
	if err != nil {
		WriteError(response, "get backend details failed: %v\n", err)
		return nil, err
	}
	log.Infof("backend response = [%+v]\n", backendResp)
	return backendResp, nil
}

func (s *APIService) ListFileShare(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "fileshare:list") {
		return
	}
	log.Info("received request for fileshare list.")

	ctx := common.InitCtxWithAuthInfo(request)
	backendId := request.PathParameter(common.REQUEST_PATH_BACKEND_ID)
	if backendId != "" {
		s.listfileshareByBackend(ctx, request, response, backendId)
	} else {
		s.listfileshareDefault(ctx, request, response)
	}
}

func (s *APIService) preparefileshareRequest(request *restful.Request, response *restful.Response) *file.ListFileShareRequest {

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

	return &file.ListFileShareRequest{
		Limit:    limit,
		Offset:   offset,
		SortKeys: sortKeys,
		SortDirs: sortDirs,
		Filter:   filter,
	}
}

func (s *APIService) listfileshareDefault(ctx context.Context, request *restful.Request, response *restful.Response) {
	listfileshareRequest := s.preparefileshareRequest(request, response)

	res, err := s.fileClient.ListFileShare(ctx, listfileshareRequest)
	if err != nil {
		WriteError(response, "list fileshares failed: %v\n", err)
		return
	}

	log.Info("list fileshares successfully.")

	err = response.WriteEntity(res)
	if err != nil {
		log.Errorf("response write entity failed: %v\n", err)
		return
	}
}

func (s *APIService) listfileshareByBackend(ctx context.Context, request *restful.Request, response *restful.Response,
	backendId string) {

	s.checkBackendExists(ctx, request, response, backendId)

	filterPathOpts := []string{"backendId"}
	filterPath, err := common.GetFilterPathParams(request, filterPathOpts)
	if err != nil {
		WriteError(response, "get filter failed: %v\n", err)
		return
	}

	listfileshareRequest := s.preparefileshareRequest(request, response)

	err = UpdateFilter(listfileshareRequest.Filter, filterPath)
	if err != nil {
		log.Errorf("update filter failed: %v\n", err)
		return
	}

	res, err := s.fileClient.ListFileShare(ctx, listfileshareRequest)
	if err != nil {
		WriteError(response, "list fileshares failed: %v\n", err)
		return
	}

	log.Info("list fileshares successfully.")

	err = response.WriteEntity(res)
	if err != nil {
		log.Errorf("response write entity failed: %v\n", err)
		return
	}
}
