// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: file.proto

package file

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/struct"
	math "math"
)

import (
	context "context"
	api "github.com/micro/go-micro/v2/api"
	client "github.com/micro/go-micro/v2/client"
	server "github.com/micro/go-micro/v2/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for File service

func NewFileEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for File service

type FileService interface {
	ListFileShare(ctx context.Context, in *ListFileShareRequest, opts ...client.CallOption) (*ListFileShareResponse, error)
	GetFileShare(ctx context.Context, in *GetFileShareRequest, opts ...client.CallOption) (*GetFileShareResponse, error)
	CreateFileShare(ctx context.Context, in *CreateFileShareRequest, opts ...client.CallOption) (*CreateFileShareResponse, error)
	UpdateFileShare(ctx context.Context, in *UpdateFileShareRequest, opts ...client.CallOption) (*UpdateFileShareResponse, error)
	DeleteFileShare(ctx context.Context, in *DeleteFileShareRequest, opts ...client.CallOption) (*DeleteFileShareResponse, error)
}

type fileService struct {
	c    client.Client
	name string
}

func NewFileService(name string, c client.Client) FileService {
	return &fileService{
		c:    c,
		name: name,
	}
}

func (c *fileService) ListFileShare(ctx context.Context, in *ListFileShareRequest, opts ...client.CallOption) (*ListFileShareResponse, error) {
	req := c.c.NewRequest(c.name, "File.ListFileShare", in)
	out := new(ListFileShareResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileService) GetFileShare(ctx context.Context, in *GetFileShareRequest, opts ...client.CallOption) (*GetFileShareResponse, error) {
	req := c.c.NewRequest(c.name, "File.GetFileShare", in)
	out := new(GetFileShareResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileService) CreateFileShare(ctx context.Context, in *CreateFileShareRequest, opts ...client.CallOption) (*CreateFileShareResponse, error) {
	req := c.c.NewRequest(c.name, "File.CreateFileShare", in)
	out := new(CreateFileShareResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileService) UpdateFileShare(ctx context.Context, in *UpdateFileShareRequest, opts ...client.CallOption) (*UpdateFileShareResponse, error) {
	req := c.c.NewRequest(c.name, "File.UpdateFileShare", in)
	out := new(UpdateFileShareResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileService) DeleteFileShare(ctx context.Context, in *DeleteFileShareRequest, opts ...client.CallOption) (*DeleteFileShareResponse, error) {
	req := c.c.NewRequest(c.name, "File.DeleteFileShare", in)
	out := new(DeleteFileShareResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for File service

type FileHandler interface {
	ListFileShare(context.Context, *ListFileShareRequest, *ListFileShareResponse) error
	GetFileShare(context.Context, *GetFileShareRequest, *GetFileShareResponse) error
	CreateFileShare(context.Context, *CreateFileShareRequest, *CreateFileShareResponse) error
	UpdateFileShare(context.Context, *UpdateFileShareRequest, *UpdateFileShareResponse) error
	DeleteFileShare(context.Context, *DeleteFileShareRequest, *DeleteFileShareResponse) error
}

func RegisterFileHandler(s server.Server, hdlr FileHandler, opts ...server.HandlerOption) error {
	type file interface {
		ListFileShare(ctx context.Context, in *ListFileShareRequest, out *ListFileShareResponse) error
		GetFileShare(ctx context.Context, in *GetFileShareRequest, out *GetFileShareResponse) error
		CreateFileShare(ctx context.Context, in *CreateFileShareRequest, out *CreateFileShareResponse) error
		UpdateFileShare(ctx context.Context, in *UpdateFileShareRequest, out *UpdateFileShareResponse) error
		DeleteFileShare(ctx context.Context, in *DeleteFileShareRequest, out *DeleteFileShareResponse) error
	}
	type File struct {
		file
	}
	h := &fileHandler{hdlr}
	return s.Handle(s.NewHandler(&File{h}, opts...))
}

type fileHandler struct {
	FileHandler
}

func (h *fileHandler) ListFileShare(ctx context.Context, in *ListFileShareRequest, out *ListFileShareResponse) error {
	return h.FileHandler.ListFileShare(ctx, in, out)
}

func (h *fileHandler) GetFileShare(ctx context.Context, in *GetFileShareRequest, out *GetFileShareResponse) error {
	return h.FileHandler.GetFileShare(ctx, in, out)
}

func (h *fileHandler) CreateFileShare(ctx context.Context, in *CreateFileShareRequest, out *CreateFileShareResponse) error {
	return h.FileHandler.CreateFileShare(ctx, in, out)
}

func (h *fileHandler) UpdateFileShare(ctx context.Context, in *UpdateFileShareRequest, out *UpdateFileShareResponse) error {
	return h.FileHandler.UpdateFileShare(ctx, in, out)
}

func (h *fileHandler) DeleteFileShare(ctx context.Context, in *DeleteFileShareRequest, out *DeleteFileShareResponse) error {
	return h.FileHandler.DeleteFileShare(ctx, in, out)
}
