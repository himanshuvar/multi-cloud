package driver

import (
	"context"

	pb "github.com/opensds/multi-cloud/file/proto"
)

// define the common driver interface for io.

type StorageDriver interface {
	CreateFileShare(ctx context.Context, fs *pb.CreateFileShareRequest) (*pb.CreateFileShareResponse, error)
	// Close: cleanup when driver needs to be stopped.
	Close() error
}
