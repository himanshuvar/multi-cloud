/*
 * EVS
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowVolumeTagsRequest struct {
	VolumeId string `json:"volume_id"`
}

func (o ShowVolumeTagsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowVolumeTagsRequest", string(data)}, " ")
}
