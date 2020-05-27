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

package mongo

import (
	"context"
	"errors"
	"math"
	"sync"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/opensds/multi-cloud/api/pkg/common"
	"github.com/opensds/multi-cloud/block/pkg/model"
	log "github.com/sirupsen/logrus"
)

var adapter = &mongoAdapter{}
var mutex sync.Mutex
var DataBaseName = "multi-cloud"
var VolumeCollection = "volumes"

func Init(host string) *mongoAdapter {
	mutex.Lock()
	defer mutex.Unlock()

	if adapter.session != nil {
		return adapter
	}

	session, err := mgo.Dial(host)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	adapter.session = session
	return adapter
}

func Exit() {
	adapter.session.Close()
}

type mongoAdapter struct {
	session *mgo.Session
	userID  string
}

// The implementation of Repository
func UpdateFilter(m bson.M, filter map[string]string) error {
	for k, v := range filter {
		m[k] = interface{}(v)
	}
	return nil
}

func UpdateContextFilter(ctx context.Context, m bson.M) error {
	// if context is admin, no need filter by tenantId.
	md, ok := metadata.FromContext(ctx)
	if !ok {
		log.Error("get context failed")
		return errors.New("get context failed")
	}

	isAdmin, _ := md[common.CTX_KEY_IS_ADMIN]
	if isAdmin != common.CTX_VAL_TRUE {
		tenantId, ok := md[common.CTX_KEY_TENANT_ID]
		if !ok {
			log.Error("get tenantid failed")
			return errors.New("get tenantid failed")
		}
		m["tenantid"] = tenantId
	}

	return nil
}

func (adapter *mongoAdapter) ListVolume(ctx context.Context, limit, offset int,
	query interface{}) ([]*model.Volume, error) {

	session := adapter.session.Copy()
	defer session.Close()

	if limit == 0 {
		limit = math.MinInt32
	}
	var backends []*model.Volume

	m := bson.M{}
	UpdateFilter(m, query.(map[string]string))
	err := UpdateContextFilter(ctx, m)
	if err != nil {
		return nil, err
	}
	log.Infof("ListBackend, limit=%d, offset=%d, m=%+v\n", limit, offset, m)

	err = session.DB(DataBaseName).C(VolumeCollection).Find(m).Skip(offset).Limit(limit).All(&backends)
	if err != nil {
		return nil, err
	}

	return backends, nil
}

func (adapter *mongoAdapter) GetVolume(ctx context.Context, id string) (*model.Volume,
	error) {
	session := adapter.session.Copy()
	defer session.Close()

	m := bson.M{"_id": bson.ObjectIdHex(id)}
	err := UpdateContextFilter(ctx, m)
	if err != nil {
		return nil, err
	}

	var volume = &model.Volume{}
	collection := session.DB(DataBaseName).C(VolumeCollection)
	err = collection.Find(m).One(volume)
	if err != nil {
		return nil, err
	}
	return volume, nil
}
