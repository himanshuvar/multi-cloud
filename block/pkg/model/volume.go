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

package model

import "github.com/globalsign/mgo/bson"

// VolumeSpec is an block device created by storage service, it can be attached
// to physical machine or virtual machine instance.
type Volume struct {
	// The uuid of the volume, it's unique in the context and generated by system
	// on successful creation of the volume. It's not allowed to be modified by
	// the user.
	// +readOnly
	Id  bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`

	// The uuid of the project that the volume belongs to.
	TenantId string `json:"tenantId,omitempty" bson:"tenantId,omitempty"`

	// The uuid of the user that the volume belongs to.
	// +optional
	UserId string `json:"userId,omitempty" bson:"userId,omitempty"`

	// The uuid of the backend that the volume belongs to.
	BackendId string `json:"backendId,omitempty" bson:"backendId,omitempty"`

	// The uuid of the snapshot which the volume is created.
	// +optional
	SnapshotId string `json:"snapshotId,omitempty bson:"snapshotId,omitempty"`

	// The name of the volume.
	Name string `json:"name,omitempty" bson:"name,omitempty"`

	// The description of the volume.
	// +optional
	Description string `json:"description,omitempty" bson:"description,omitempty"`

	// The type of the volume.
	Type string `json:"type,omitempty" bson:"type,omitempty"`

	// The size of the volume requested by the user.
	// Default unit of volume Size is GB.
	Size int64 `json:"size,omitempty" bson:"size,omitempty"`

	// The location that volume belongs to.
	Region string `json:"region,omitempty bson:"region,omitempty"`

	// The locality that volume belongs to.
	AvailabilityZone string `json:"availabilityZone,omitempty bson:"availabilityZone,omitempty"`

	// The status of the volume.
	Status string `json:"status,omitempty" bson:"status,omitempty"`

	// Any tags assigned to the volume.
	Tags []Tag `json:"tags,omitempty" bson:"tags,omitempty"`

	// Indicates whether Multi-Attach is enabled.
	// +optional
	MultiAttach bool `json:"multiAttachEnabled,omitempty" bson:"multiAttachEnabled,omitempty"`

	// Indicates whether the volume is encrypted.
	// +optional
	Encrypted bool `json:"encrypted,omitempty" bson:"encrypted,omitempty"`

	// EncryptionSettings that was used to protect the volume encryption.
	// +optional
	EncryptionSettings map[string]string `json:"encryptionSettings,omitempty bson:"encryptionSettings,omitempty"`

	// Metadata should be kept until the semantics between volume and backend storage resource.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty bson:"metadata,omitempty"`
}

type Tag struct {
	// The key of the tag.
	Key string `json:"key,omitempty bson:"key,omitempty"`

	// The value of the tag.
	Value string `json:"value,omitempty bson:"value,omitempty"`
}
