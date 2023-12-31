// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.15.8
// source: archive.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ArchiveServiceClient is the client API for ArchiveService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ArchiveServiceClient interface {
	GetRecentMessages(ctx context.Context, in *GetRecentMessagesRequest, opts ...grpc.CallOption) (*GetRecentMessagesResponse, error)
}

type archiveServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewArchiveServiceClient(cc grpc.ClientConnInterface) ArchiveServiceClient {
	return &archiveServiceClient{cc}
}

func (c *archiveServiceClient) GetRecentMessages(ctx context.Context, in *GetRecentMessagesRequest, opts ...grpc.CallOption) (*GetRecentMessagesResponse, error) {
	out := new(GetRecentMessagesResponse)
	err := c.cc.Invoke(ctx, "/pb.ArchiveService/GetRecentMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ArchiveServiceServer is the server API for ArchiveService service.
// All implementations must embed UnimplementedArchiveServiceServer
// for forward compatibility
type ArchiveServiceServer interface {
	GetRecentMessages(context.Context, *GetRecentMessagesRequest) (*GetRecentMessagesResponse, error)
	mustEmbedUnimplementedArchiveServiceServer()
}

// UnimplementedArchiveServiceServer must be embedded to have forward compatible implementations.
type UnimplementedArchiveServiceServer struct {
}

func (UnimplementedArchiveServiceServer) GetRecentMessages(context.Context, *GetRecentMessagesRequest) (*GetRecentMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRecentMessages not implemented")
}
func (UnimplementedArchiveServiceServer) mustEmbedUnimplementedArchiveServiceServer() {}

// UnsafeArchiveServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ArchiveServiceServer will
// result in compilation errors.
type UnsafeArchiveServiceServer interface {
	mustEmbedUnimplementedArchiveServiceServer()
}

func RegisterArchiveServiceServer(s grpc.ServiceRegistrar, srv ArchiveServiceServer) {
	s.RegisterService(&ArchiveService_ServiceDesc, srv)
}

func _ArchiveService_GetRecentMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRecentMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArchiveServiceServer).GetRecentMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.ArchiveService/GetRecentMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArchiveServiceServer).GetRecentMessages(ctx, req.(*GetRecentMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ArchiveService_ServiceDesc is the grpc.ServiceDesc for ArchiveService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ArchiveService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.ArchiveService",
	HandlerType: (*ArchiveServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRecentMessages",
			Handler:    _ArchiveService_GetRecentMessages_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "archive.proto",
}
