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
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/golang/protobuf/jsonpb"
	"github.com/micro/go-micro/v2/client"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/api/pkg/policy"
	"github.com/opensds/multi-cloud/backend/proto"
	"github.com/opensds/multi-cloud/file/pkg/model"
	"github.com/opensds/multi-cloud/file/proto"

	pstruct "github.com/golang/protobuf/ptypes/struct"
	c "github.com/opensds/multi-cloud/api/pkg/context"
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

func WriteError(response *restful.Response, msg string, errCode int, err error) {
	log.Errorf(msg, err)
	error := response.WriteError(errCode, err)
	if error != nil {
		log.Errorf("Response write status and http error failed: %v\n", err)
		return
	}
}

func (s *APIService) checkBackendExists(ctx context.Context, request *restful.Request, response *restful.Response,
	backendId string){
	backendResp, err := s.backendClient.GetBackend(ctx, &backend.GetBackendRequest{Id: backendId})
	if err != nil {
		WriteError(response, "Get Backend details failed: %v\n", http.StatusNotFound, err)
		return
	}
	log.Infof("Backend response = [%+v]\n", backendResp)
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

	listfileshareRequest := s.prepareFileShareRequest(request, response)

	res, err := s.fileClient.ListFileShare(ctx, listfileshareRequest)
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

	s.checkBackendExists(ctx, request, response, backendId)

	filterPathOpts := []string{"backendId"}
	filterPath, err := common.GetFilterPathParams(request, filterPathOpts)
	if err != nil {
		WriteError(response, "Get filter failed: %v\n", http.StatusBadRequest, err)
		return
	}

	listfileshareRequest := s.prepareFileShareRequest(request, response)

	err = UpdateFilter(listfileshareRequest.Filter, filterPath)
	if err != nil {
		log.Errorf("Update filter failed: %v\n", err)
		return
	}

	res, err := s.fileClient.ListFileShare(ctx, listfileshareRequest)
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

func (s *APIService) GetFileShare(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "fileshare:get") {
		return
	}
	log.Infof("Received request for file share details: %s\n", request.PathParameter("id"))
	id := request.PathParameter("id")

	ctx := common.InitCtxWithAuthInfo(request)

	backendId := request.PathParameter(common.REQUEST_PATH_BACKEND_ID)
	s.checkBackendExists(ctx, request, response, backendId)

	res, err := s.fileClient.GetFileShare(ctx, &file.GetFileShareRequest{Id: id})
	if err != nil {
		log.Errorf("failed to get file share details: %v\n", err)
		response.WriteError(http.StatusNotFound, err)
		return
	}

	log.Info("Get file share details successfully.")
	response.WriteEntity(res.Fileshare)
}

func (s *APIService) CreateFileShare(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "fileshare:create") {
		return
	}
	log.Info("Received request for creating file share.")

	fileshare := &model.FileShare{}

	err := request.ReadEntity(&fileshare)
	if err != nil {
		log.Errorf("failed to read request body: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	var tags []*file.Tag
	for _, tag := range fileshare.Tags {
		tags = append(tags, &file.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	metadata, err := ToStruct(fileshare.Metadata)

	fs := &file.FileShare{
		Name:                 fileshare.Name,
		Description:          fileshare.Description,
		Region:               fileshare.Region,
		AvailabilityZone:     fileshare.AvailabilityZone,
		Encrypted:            *fileshare.Encrypted,
		EncryptionSettings:   fileshare.EncryptionSettings,
		Tags:                 tags,
		Metadata:             metadata,
	}

	ctx := common.InitCtxWithAuthInfo(request)

	backendId := request.PathParameter(common.REQUEST_PATH_BACKEND_ID)
	s.checkBackendExists(ctx, request, response, backendId)
	fs.BackendId = backendId

	actx := request.Attribute(c.KContext).(*c.Context)
	fs.TenantId = actx.TenantId
	fs.UserId = actx.UserId
	res, err := s.fileClient.CreateFileShare(ctx, &file.CreateFileShareRequest{Fileshare:fs})
	if err != nil {
		log.Errorf("failed to create file share: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Info("Create file share successfully.")
	response.WriteEntity(res.Fileshare)
}

func (s *APIService) DeleteFileShare(request *restful.Request, response *restful.Response) {
	if !policy.Authorize(request, response, "fileshare:delete") {
		return
	}

	id := request.PathParameter("id")
	log.Infof("Received request for deleting file share: %s\n", id)

	ctx := common.InitCtxWithAuthInfo(request)

	backendId := request.PathParameter(common.REQUEST_PATH_BACKEND_ID)
	s.checkBackendExists(ctx, request, response, backendId)

	res, err := s.fileClient.DeleteFileShare(ctx, &file.DeleteFileShareRequest{Id: id})
	if err != nil {
		log.Errorf("failed to delete file share: %v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Info("Delete file share successfully.")
	response.WriteEntity(res)
}
