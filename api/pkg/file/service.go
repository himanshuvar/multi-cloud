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
	log.Debugf("Getting the filters updated for the request.")

	for k, v := range filter {
		log.Debugf("Filter Key = [%+v], Key = [%+v]\n", k, v)
		reqFilter[k] = v
	}
	return nil
}

func WriteError(response *restful.Response, msg string, errCode int, err error) {
	log.Errorf(msg, err)
	error := response.WriteError(errCode, err)
	if error != nil {
		log.Errorf("Response write status and http error failed: %v\n", err)
		return
	}
}

func (s *APIService) checkBackendExists(ctx context.Context, request *restful.Request, response *restful.Response,
	backendId string) error{

	backendResp, err := s.backendClient.GetBackend(ctx, &backend.GetBackendRequest{Id: backendId})
	if err != nil {
		WriteError(response, "Get Backend details failed: %v\n", http.StatusNotFound, err)
		return err
	}
	log.Infof("Backend response = [%+v]\n", backendResp)

	return nil
}

func (s *APIService) prepareFileShareRequest(request *restful.Request, response *restful.Response) *file.ListFileShareRequest {

	limit, offset, err := common.GetPaginationParam(request)
	if err != nil {
		WriteError(response, "Get pagination parameters failed: %v\n", http.StatusInternalServerError, err)
		return nil
	}

	sortKeys, sortDirs, err := common.GetSortParam(request)
	if err != nil {
		WriteError(response, "Get sort parameters failed: %v\n", http.StatusInternalServerError, err)
		return nil
	}

	filterOpts := []string{"name", "type", "status"}
	filter, err := common.GetFilter(request, filterOpts)
	if err != nil {
		WriteError(response, "Get filter failed: %v\n", http.StatusInternalServerError, err)
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

func (s *APIService) ListFileShare(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "fileshare:list") {
		return
	}
	log.Info("Received request for File Share List.")

	ctx := common.InitCtxWithAuthInfo(request)

	listFileShareRequest := s.prepareFileShareRequest(request, response)

	if listFileShareRequest == nil {
		return
	}

	res, err := s.fileClient.ListFileShare(ctx, listFileShareRequest)
	if err != nil {
		WriteError(response, "List FileShares failed: %v\n", http.StatusInternalServerError, err)
		return
	}

	log.Info("List FileShares successfully.")

	err = response.WriteEntity(res)
	if err != nil {
		log.Errorf("Response write entity failed: %v\n", err)
		return
	}
}

func (s *APIService) ListFileShareByBackend(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "fileshare:list") {
		return
	}
	log.Info("Received request for File Share List for a Backend.")

	ctx := common.InitCtxWithAuthInfo(request)
	backendId := request.PathParameter(common.REQUEST_PATH_BACKEND_ID)

	if s.checkBackendExists(ctx, request, response, backendId) != nil {
		return
	}

	filterPathOpts := []string{"backendId"}
	filterPath, err := common.GetFilterPathParams(request, filterPathOpts)
	if err != nil {
		WriteError(response, "Get filter failed: %v\n", http.StatusBadRequest, err)
		return
	}

	listFileShareRequest := s.prepareFileShareRequest(request, response)

	if listFileShareRequest == nil {
		return
	}

	if UpdateFilter(listFileShareRequest.Filter, filterPath) != nil {
		log.Errorf("Update filter failed: %v\n", err)
		return
	}

	res, err := s.fileClient.ListFileShare(ctx, listFileShareRequest)
	if err != nil {
		WriteError(response, "List FileShares failed: %v\n", http.StatusInternalServerError, err)
		return
	}

	log.Info("List FileShares successfully.")

	err = response.WriteEntity(res)
	if err != nil {
		WriteError(response, "Response write entity failed: %v\n", http.StatusInternalServerError, err)
		return
	}
}
