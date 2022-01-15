// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package linter_proto

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

// LinterClient is the client API for Linter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LinterClient interface {
	Lint(ctx context.Context, in *LintRequest, opts ...grpc.CallOption) (*LintResponse, error)
}

type linterClient struct {
	cc grpc.ClientConnInterface
}

func NewLinterClient(cc grpc.ClientConnInterface) LinterClient {
	return &linterClient{cc}
}

func (c *linterClient) Lint(ctx context.Context, in *LintRequest, opts ...grpc.CallOption) (*LintResponse, error) {
	out := new(LintResponse)
	err := c.cc.Invoke(ctx, "/Linter/Lint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LinterServer is the server API for Linter service.
// All implementations must embed UnimplementedLinterServer
// for forward compatibility
type LinterServer interface {
	Lint(context.Context, *LintRequest) (*LintResponse, error)
	mustEmbedUnimplementedLinterServer()
}

// UnimplementedLinterServer must be embedded to have forward compatible implementations.
type UnimplementedLinterServer struct {
}

func (UnimplementedLinterServer) Lint(context.Context, *LintRequest) (*LintResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lint not implemented")
}
func (UnimplementedLinterServer) mustEmbedUnimplementedLinterServer() {}

// UnsafeLinterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LinterServer will
// result in compilation errors.
type UnsafeLinterServer interface {
	mustEmbedUnimplementedLinterServer()
}

func RegisterLinterServer(s grpc.ServiceRegistrar, srv LinterServer) {
	s.RegisterService(&Linter_ServiceDesc, srv)
}

func _Linter_Lint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LinterServer).Lint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Linter/Lint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LinterServer).Lint(ctx, req.(*LintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Linter_ServiceDesc is the grpc.ServiceDesc for Linter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Linter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Linter",
	HandlerType: (*LinterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Lint",
			Handler:    _Linter_Lint_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "linter_proto/linter_proto.proto",
}

// LoadBalancerClient is the client API for LoadBalancer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LoadBalancerClient interface {
	SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigResponse, error)
}

type loadBalancerClient struct {
	cc grpc.ClientConnInterface
}

func NewLoadBalancerClient(cc grpc.ClientConnInterface) LoadBalancerClient {
	return &loadBalancerClient{cc}
}

func (c *loadBalancerClient) SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigResponse, error) {
	out := new(SetConfigResponse)
	err := c.cc.Invoke(ctx, "/LoadBalancer/SetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LoadBalancerServer is the server API for LoadBalancer service.
// All implementations must embed UnimplementedLoadBalancerServer
// for forward compatibility
type LoadBalancerServer interface {
	SetConfig(context.Context, *SetConfigRequest) (*SetConfigResponse, error)
	mustEmbedUnimplementedLoadBalancerServer()
}

// UnimplementedLoadBalancerServer must be embedded to have forward compatible implementations.
type UnimplementedLoadBalancerServer struct {
}

func (UnimplementedLoadBalancerServer) SetConfig(context.Context, *SetConfigRequest) (*SetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConfig not implemented")
}
func (UnimplementedLoadBalancerServer) mustEmbedUnimplementedLoadBalancerServer() {}

// UnsafeLoadBalancerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LoadBalancerServer will
// result in compilation errors.
type UnsafeLoadBalancerServer interface {
	mustEmbedUnimplementedLoadBalancerServer()
}

func RegisterLoadBalancerServer(s grpc.ServiceRegistrar, srv LoadBalancerServer) {
	s.RegisterService(&LoadBalancer_ServiceDesc, srv)
}

func _LoadBalancer_SetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LoadBalancerServer).SetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/LoadBalancer/SetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LoadBalancerServer).SetConfig(ctx, req.(*SetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LoadBalancer_ServiceDesc is the grpc.ServiceDesc for LoadBalancer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LoadBalancer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "LoadBalancer",
	HandlerType: (*LoadBalancerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetConfig",
			Handler:    _LoadBalancer_SetConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "linter_proto/linter_proto.proto",
}

// MachineManagerClient is the client API for MachineManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MachineManagerClient interface {
	AppendLinter(ctx context.Context, in *LinterAttributes, opts ...grpc.CallOption) (*LinterResponse, error)
	RemoveLinter(ctx context.Context, in *LinterAttributes, opts ...grpc.CallOption) (*LinterResponse, error)
	SetProportions(ctx context.Context, in *LoadBalancingProportions, opts ...grpc.CallOption) (*LinterResponse, error)
}

type machineManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewMachineManagerClient(cc grpc.ClientConnInterface) MachineManagerClient {
	return &machineManagerClient{cc}
}

func (c *machineManagerClient) AppendLinter(ctx context.Context, in *LinterAttributes, opts ...grpc.CallOption) (*LinterResponse, error) {
	out := new(LinterResponse)
	err := c.cc.Invoke(ctx, "/MachineManager/AppendLinter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *machineManagerClient) RemoveLinter(ctx context.Context, in *LinterAttributes, opts ...grpc.CallOption) (*LinterResponse, error) {
	out := new(LinterResponse)
	err := c.cc.Invoke(ctx, "/MachineManager/RemoveLinter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *machineManagerClient) SetProportions(ctx context.Context, in *LoadBalancingProportions, opts ...grpc.CallOption) (*LinterResponse, error) {
	out := new(LinterResponse)
	err := c.cc.Invoke(ctx, "/MachineManager/SetProportions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MachineManagerServer is the server API for MachineManager service.
// All implementations must embed UnimplementedMachineManagerServer
// for forward compatibility
type MachineManagerServer interface {
	AppendLinter(context.Context, *LinterAttributes) (*LinterResponse, error)
	RemoveLinter(context.Context, *LinterAttributes) (*LinterResponse, error)
	SetProportions(context.Context, *LoadBalancingProportions) (*LinterResponse, error)
	mustEmbedUnimplementedMachineManagerServer()
}

// UnimplementedMachineManagerServer must be embedded to have forward compatible implementations.
type UnimplementedMachineManagerServer struct {
}

func (UnimplementedMachineManagerServer) AppendLinter(context.Context, *LinterAttributes) (*LinterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendLinter not implemented")
}
func (UnimplementedMachineManagerServer) RemoveLinter(context.Context, *LinterAttributes) (*LinterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveLinter not implemented")
}
func (UnimplementedMachineManagerServer) SetProportions(context.Context, *LoadBalancingProportions) (*LinterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetProportions not implemented")
}
func (UnimplementedMachineManagerServer) mustEmbedUnimplementedMachineManagerServer() {}

// UnsafeMachineManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MachineManagerServer will
// result in compilation errors.
type UnsafeMachineManagerServer interface {
	mustEmbedUnimplementedMachineManagerServer()
}

func RegisterMachineManagerServer(s grpc.ServiceRegistrar, srv MachineManagerServer) {
	s.RegisterService(&MachineManager_ServiceDesc, srv)
}

func _MachineManager_AppendLinter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LinterAttributes)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MachineManagerServer).AppendLinter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MachineManager/AppendLinter",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MachineManagerServer).AppendLinter(ctx, req.(*LinterAttributes))
	}
	return interceptor(ctx, in, info, handler)
}

func _MachineManager_RemoveLinter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LinterAttributes)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MachineManagerServer).RemoveLinter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MachineManager/RemoveLinter",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MachineManagerServer).RemoveLinter(ctx, req.(*LinterAttributes))
	}
	return interceptor(ctx, in, info, handler)
}

func _MachineManager_SetProportions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadBalancingProportions)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MachineManagerServer).SetProportions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MachineManager/SetProportions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MachineManagerServer).SetProportions(ctx, req.(*LoadBalancingProportions))
	}
	return interceptor(ctx, in, info, handler)
}

// MachineManager_ServiceDesc is the grpc.ServiceDesc for MachineManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MachineManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "MachineManager",
	HandlerType: (*MachineManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AppendLinter",
			Handler:    _MachineManager_AppendLinter_Handler,
		},
		{
			MethodName: "RemoveLinter",
			Handler:    _MachineManager_RemoveLinter_Handler,
		},
		{
			MethodName: "SetProportions",
			Handler:    _MachineManager_SetProportions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "linter_proto/linter_proto.proto",
}
