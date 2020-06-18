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

package azure

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-file-go/azfile"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	backendpb "github.com/opensds/multi-cloud/backend/proto"
	file "github.com/opensds/multi-cloud/file/proto"
	log "github.com/sirupsen/logrus"
)

// TryTimeout indicates the maximum time allowed for any single try of an HTTP request.
var MaxTimeForSingleHttpRequest = 50 * time.Minute

type AzureAdapter struct {
	backend      *backendpb.BackendDetail
	pipeline      pipeline.Pipeline
}

func ToAzureMetadata(pbs *pstruct.Struct) (azfile.Metadata, error) {
	fields := pbs.GetFields()

	valuesMap := make(map[string]string)

	for key, value := range fields {
		if v, ok := value.GetKind().(*pstruct.Value_StringValue); ok {
			valuesMap[key] = v.StringValue
		} else {
			msg := fmt.Sprintf("Failed to parse field for key = [%+v], value = [%+v]", key, value)
			err := errors.New(msg)
			log.Errorf(msg)
			return nil, err
		}
	}
	return valuesMap, nil
}


func (ad *AzureAdapter) createPipeline(endpoint string, acountName string, accountKey string) (pipeline.Pipeline, error) {
	credential, err := azfile.NewSharedKeyCredential(acountName, accountKey)

	if err != nil {
		log.Infof("create credential[Azure File Share] failed, err:%v\n", err)
		return nil, err
	}

	//create Azure Pipeline
	p := azfile.NewPipeline(credential, azfile.PipelineOptions{
		Retry: azfile.RetryOptions{
			TryTimeout: MaxTimeForSingleHttpRequest,
		},
	})

	return p, nil
}

func (ad *AzureAdapter) createFileShareURL(fileshareName string) (azfile.ShareURL, error) {

	//create fileShareURL
	URL, _ := url.Parse(ad.backend.Endpoint + fileshareName)

	return azfile.NewShareURL(*URL, ad.pipeline), nil
}

func (ad *AzureAdapter) CreateFileShare(ctx context.Context, fs *file.CreateFileShareRequest) (*file.CreateFileShareResponse, error) {

	shareURL, err := ad.createFileShareURL(fs.Fileshare.Name)

	if err != nil {
		log.Infof("create Azure File Share URL failed, err:%v\n", err)
		return nil, err
	}

	//result, err := shareURL.Create(ctx, azfile.Metadata{"createdby": "AKS"}, int32(fs.Fileshare.Size))

	metadata, err := ToAzureMetadata(fs.Fileshare.Metadata)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	result, err := shareURL.Create(ctx, metadata, int32(fs.Fileshare.Size))
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Create File share response = %+v", result)

	return nil, nil
}

func (ad *AzureAdapter) GetFileShare(ctx context.Context, fs *file.GetFileShareRequest) (*file.GetFileShareResponse, error) {
	panic("implement me")
}

func (ad *AzureAdapter) ListFileShare(ctx context.Context, fs *file.ListFileShareRequest) (*file.ListFileShareResponse, error) {
	panic("implement me")
}

/*
func ListFileshares(accountName, accountKey string) {
	credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
	u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net/aks-share2", accountName))
	if err != nil {
		log.Infof("create credential[Azure Blob] failed, err:%v\n", err)
		return
	}
	shareURL := azfile.NewShareURL(*u, azfile.NewPipeline(credential, azfile.PipelineOptions{}))
	ctx := context.Background()
	_, err = shareURL.Create(ctx, azfile.Metadata{"createdby": "AKS"}, 2)
	if err != nil {
		log.Fatal(err)
	}
	// List file share
	basicClient, client_err := storage.NewBasicClient(accountName, accountKey)
	if client_err != nil {
		fmt.Println("Error in getting client")
		return
	}
	fsc := basicClient.GetFileService()
	rsp, rsp_err := fsc.ListShares(storage.ListSharesParameters{})
	if rsp_err != nil {
		fmt.Println("Error in response")
		return
	}
	fmt.Println(rsp)
}
*/

func (ad *AzureAdapter) Close() error {
	// TODO:
	return nil
}
