// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.0
// source: commentservice.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	CommentService_CreateComment_FullMethodName = "/commentservice.CommentService/CreateComment"
	CommentService_GetComment_FullMethodName    = "/commentservice.CommentService/GetComment"
	CommentService_ListComments_FullMethodName  = "/commentservice.CommentService/ListComments"
	CommentService_UpdateComment_FullMethodName = "/commentservice.CommentService/UpdateComment"
	CommentService_DeleteComment_FullMethodName = "/commentservice.CommentService/DeleteComment"
)

// CommentServiceClient is the client API for CommentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CommentServiceClient interface {
	CreateComment(ctx context.Context, in *CreateCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	GetComment(ctx context.Context, in *GetCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	ListComments(ctx context.Context, in *ListCommentsRequest, opts ...grpc.CallOption) (*ListCommentsResponse, error)
	UpdateComment(ctx context.Context, in *UpdateCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	DeleteComment(ctx context.Context, in *DeleteCommentRequest, opts ...grpc.CallOption) (*Empty, error)
}

type commentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCommentServiceClient(cc grpc.ClientConnInterface) CommentServiceClient {
	return &commentServiceClient{cc}
}

func (c *commentServiceClient) CreateComment(ctx context.Context, in *CreateCommentRequest, opts ...grpc.CallOption) (*Comment, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Comment)
	err := c.cc.Invoke(ctx, CommentService_CreateComment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commentServiceClient) GetComment(ctx context.Context, in *GetCommentRequest, opts ...grpc.CallOption) (*Comment, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Comment)
	err := c.cc.Invoke(ctx, CommentService_GetComment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commentServiceClient) ListComments(ctx context.Context, in *ListCommentsRequest, opts ...grpc.CallOption) (*ListCommentsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListCommentsResponse)
	err := c.cc.Invoke(ctx, CommentService_ListComments_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commentServiceClient) UpdateComment(ctx context.Context, in *UpdateCommentRequest, opts ...grpc.CallOption) (*Comment, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Comment)
	err := c.cc.Invoke(ctx, CommentService_UpdateComment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *commentServiceClient) DeleteComment(ctx context.Context, in *DeleteCommentRequest, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, CommentService_DeleteComment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CommentServiceServer is the server API for CommentService service.
// All implementations must embed UnimplementedCommentServiceServer
// for forward compatibility.
type CommentServiceServer interface {
	CreateComment(context.Context, *CreateCommentRequest) (*Comment, error)
	GetComment(context.Context, *GetCommentRequest) (*Comment, error)
	ListComments(context.Context, *ListCommentsRequest) (*ListCommentsResponse, error)
	UpdateComment(context.Context, *UpdateCommentRequest) (*Comment, error)
	DeleteComment(context.Context, *DeleteCommentRequest) (*Empty, error)
	mustEmbedUnimplementedCommentServiceServer()
}

// UnimplementedCommentServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCommentServiceServer struct{}

func (UnimplementedCommentServiceServer) CreateComment(context.Context, *CreateCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateComment not implemented")
}
func (UnimplementedCommentServiceServer) GetComment(context.Context, *GetCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetComment not implemented")
}
func (UnimplementedCommentServiceServer) ListComments(context.Context, *ListCommentsRequest) (*ListCommentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListComments not implemented")
}
func (UnimplementedCommentServiceServer) UpdateComment(context.Context, *UpdateCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateComment not implemented")
}
func (UnimplementedCommentServiceServer) DeleteComment(context.Context, *DeleteCommentRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteComment not implemented")
}
func (UnimplementedCommentServiceServer) mustEmbedUnimplementedCommentServiceServer() {}
func (UnimplementedCommentServiceServer) testEmbeddedByValue()                        {}

// UnsafeCommentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CommentServiceServer will
// result in compilation errors.
type UnsafeCommentServiceServer interface {
	mustEmbedUnimplementedCommentServiceServer()
}

func RegisterCommentServiceServer(s grpc.ServiceRegistrar, srv CommentServiceServer) {
	// If the following call pancis, it indicates UnimplementedCommentServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CommentService_ServiceDesc, srv)
}

func _CommentService_CreateComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommentServiceServer).CreateComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommentService_CreateComment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommentServiceServer).CreateComment(ctx, req.(*CreateCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommentService_GetComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommentServiceServer).GetComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommentService_GetComment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommentServiceServer).GetComment(ctx, req.(*GetCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommentService_ListComments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCommentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommentServiceServer).ListComments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommentService_ListComments_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommentServiceServer).ListComments(ctx, req.(*ListCommentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommentService_UpdateComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommentServiceServer).UpdateComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommentService_UpdateComment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommentServiceServer).UpdateComment(ctx, req.(*UpdateCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CommentService_DeleteComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommentServiceServer).DeleteComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CommentService_DeleteComment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommentServiceServer).DeleteComment(ctx, req.(*DeleteCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CommentService_ServiceDesc is the grpc.ServiceDesc for CommentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CommentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "commentservice.CommentService",
	HandlerType: (*CommentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateComment",
			Handler:    _CommentService_CreateComment_Handler,
		},
		{
			MethodName: "GetComment",
			Handler:    _CommentService_GetComment_Handler,
		},
		{
			MethodName: "ListComments",
			Handler:    _CommentService_ListComments_Handler,
		},
		{
			MethodName: "UpdateComment",
			Handler:    _CommentService_UpdateComment_Handler,
		},
		{
			MethodName: "DeleteComment",
			Handler:    _CommentService_DeleteComment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "commentservice.proto",
}
