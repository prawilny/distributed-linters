// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.19.2
// source: linter_proto/linter_proto.proto

package linter_proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SetConfigResponse_Code int32

const (
	SetConfigResponse_SUCCESS SetConfigResponse_Code = 0
)

// Enum value maps for SetConfigResponse_Code.
var (
	SetConfigResponse_Code_name = map[int32]string{
		0: "SUCCESS",
	}
	SetConfigResponse_Code_value = map[string]int32{
		"SUCCESS": 0,
	}
)

func (x SetConfigResponse_Code) Enum() *SetConfigResponse_Code {
	p := new(SetConfigResponse_Code)
	*p = x
	return p
}

func (x SetConfigResponse_Code) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SetConfigResponse_Code) Descriptor() protoreflect.EnumDescriptor {
	return file_linter_proto_linter_proto_proto_enumTypes[0].Descriptor()
}

func (SetConfigResponse_Code) Type() protoreflect.EnumType {
	return &file_linter_proto_linter_proto_proto_enumTypes[0]
}

func (x SetConfigResponse_Code) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SetConfigResponse_Code.Descriptor instead.
func (SetConfigResponse_Code) EnumDescriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{3, 0}
}

type LintRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Language string `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	Content  []byte `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *LintRequest) Reset() {
	*x = LintRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LintRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LintRequest) ProtoMessage() {}

func (x *LintRequest) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LintRequest.ProtoReflect.Descriptor instead.
func (*LintRequest) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{0}
}

func (x *LintRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *LintRequest) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type LintResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hints []*LintResponse_Hint `protobuf:"bytes,1,rep,name=hints,proto3" json:"hints,omitempty"`
}

func (x *LintResponse) Reset() {
	*x = LintResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LintResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LintResponse) ProtoMessage() {}

func (x *LintResponse) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LintResponse.ProtoReflect.Descriptor instead.
func (*LintResponse) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{1}
}

func (x *LintResponse) GetHints() []*LintResponse_Hint {
	if x != nil {
		return x.Hints
	}
	return nil
}

type SetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Workers []*SetConfigRequest_Worker `protobuf:"bytes,1,rep,name=workers,proto3" json:"workers,omitempty"`
	Weights []*SetConfigRequest_Weight `protobuf:"bytes,2,rep,name=weights,proto3" json:"weights,omitempty"`
}

func (x *SetConfigRequest) Reset() {
	*x = SetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetConfigRequest) ProtoMessage() {}

func (x *SetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetConfigRequest.ProtoReflect.Descriptor instead.
func (*SetConfigRequest) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{2}
}

func (x *SetConfigRequest) GetWorkers() []*SetConfigRequest_Worker {
	if x != nil {
		return x.Workers
	}
	return nil
}

func (x *SetConfigRequest) GetWeights() []*SetConfigRequest_Weight {
	if x != nil {
		return x.Weights
	}
	return nil
}

type SetConfigResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code SetConfigResponse_Code `protobuf:"varint,1,opt,name=code,proto3,enum=SetConfigResponse_Code" json:"code,omitempty"`
}

func (x *SetConfigResponse) Reset() {
	*x = SetConfigResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetConfigResponse) ProtoMessage() {}

func (x *SetConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetConfigResponse.ProtoReflect.Descriptor instead.
func (*SetConfigResponse) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{3}
}

func (x *SetConfigResponse) GetCode() SetConfigResponse_Code {
	if x != nil {
		return x.Code
	}
	return SetConfigResponse_SUCCESS
}

type LintResponse_Hint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HintText  string `protobuf:"bytes,1,opt,name=hintText,proto3" json:"hintText,omitempty"`
	StartByte int32  `protobuf:"varint,2,opt,name=startByte,proto3" json:"startByte,omitempty"`
	EndByte   int32  `protobuf:"varint,3,opt,name=endByte,proto3" json:"endByte,omitempty"`
}

func (x *LintResponse_Hint) Reset() {
	*x = LintResponse_Hint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LintResponse_Hint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LintResponse_Hint) ProtoMessage() {}

func (x *LintResponse_Hint) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LintResponse_Hint.ProtoReflect.Descriptor instead.
func (*LintResponse_Hint) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{1, 0}
}

func (x *LintResponse_Hint) GetHintText() string {
	if x != nil {
		return x.HintText
	}
	return ""
}

func (x *LintResponse_Hint) GetStartByte() int32 {
	if x != nil {
		return x.StartByte
	}
	return 0
}

func (x *LintResponse_Hint) GetEndByte() int32 {
	if x != nil {
		return x.EndByte
	}
	return 0
}

type SetConfigRequest_Weight struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Language string  `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	Version  string  `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Weight   float32 `protobuf:"fixed32,3,opt,name=weight,proto3" json:"weight,omitempty"`
}

func (x *SetConfigRequest_Weight) Reset() {
	*x = SetConfigRequest_Weight{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetConfigRequest_Weight) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetConfigRequest_Weight) ProtoMessage() {}

func (x *SetConfigRequest_Weight) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetConfigRequest_Weight.ProtoReflect.Descriptor instead.
func (*SetConfigRequest_Weight) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{2, 0}
}

func (x *SetConfigRequest_Weight) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *SetConfigRequest_Weight) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *SetConfigRequest_Weight) GetWeight() float32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

type SetConfigRequest_Worker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address  string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Port     int32  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Language string `protobuf:"bytes,3,opt,name=language,proto3" json:"language,omitempty"`
	Version  string `protobuf:"bytes,4,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *SetConfigRequest_Worker) Reset() {
	*x = SetConfigRequest_Worker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_linter_proto_linter_proto_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetConfigRequest_Worker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetConfigRequest_Worker) ProtoMessage() {}

func (x *SetConfigRequest_Worker) ProtoReflect() protoreflect.Message {
	mi := &file_linter_proto_linter_proto_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetConfigRequest_Worker.ProtoReflect.Descriptor instead.
func (*SetConfigRequest_Worker) Descriptor() ([]byte, []int) {
	return file_linter_proto_linter_proto_proto_rawDescGZIP(), []int{2, 1}
}

func (x *SetConfigRequest_Worker) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *SetConfigRequest_Worker) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *SetConfigRequest_Worker) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *SetConfigRequest_Worker) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

var File_linter_proto_linter_proto_proto protoreflect.FileDescriptor

var file_linter_proto_linter_proto_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6c, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x43, 0x0a, 0x0b, 0x4c, 0x69, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x94, 0x01, 0x0a, 0x0c, 0x4c, 0x69, 0x6e, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x68, 0x69, 0x6e, 0x74, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x4c, 0x69, 0x6e, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x48, 0x69, 0x6e, 0x74, 0x52, 0x05, 0x68, 0x69, 0x6e, 0x74,
	0x73, 0x1a, 0x5a, 0x0a, 0x04, 0x48, 0x69, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x69, 0x6e,
	0x74, 0x54, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x69, 0x6e,
	0x74, 0x54, 0x65, 0x78, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x42, 0x79,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x42,
	0x79, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x64, 0x42, 0x79, 0x74, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x42, 0x79, 0x74, 0x65, 0x22, 0xc0, 0x02,
	0x0a, 0x10, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x32, 0x0a, 0x07, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x52, 0x07, 0x77,
	0x6f, 0x72, 0x6b, 0x65, 0x72, 0x73, 0x12, 0x32, 0x0a, 0x07, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x57, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x52, 0x07, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x73, 0x1a, 0x56, 0x0a, 0x06, 0x57, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x02, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x1a, 0x6c, 0x0a, 0x06, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61,
	0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x61,
	0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x22, 0x55, 0x0a, 0x11, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x63, 0x6f,
	0x64, 0x65, 0x22, 0x13, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55,
	0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x32, 0x2f, 0x0a, 0x06, 0x4c, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x12, 0x25, 0x0a, 0x04, 0x4c, 0x69, 0x6e, 0x74, 0x12, 0x0c, 0x2e, 0x4c, 0x69, 0x6e, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x4c, 0x69, 0x6e, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x32, 0x44, 0x0a, 0x0c, 0x4c, 0x6f, 0x61, 0x64,
	0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x12, 0x34, 0x0a, 0x09, 0x53, 0x65, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x11, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x12,
	0x5a, 0x10, 0x61, 0x2f, 0x62, 0x2f, 0x6c, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x5f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_linter_proto_linter_proto_proto_rawDescOnce sync.Once
	file_linter_proto_linter_proto_proto_rawDescData = file_linter_proto_linter_proto_proto_rawDesc
)

func file_linter_proto_linter_proto_proto_rawDescGZIP() []byte {
	file_linter_proto_linter_proto_proto_rawDescOnce.Do(func() {
		file_linter_proto_linter_proto_proto_rawDescData = protoimpl.X.CompressGZIP(file_linter_proto_linter_proto_proto_rawDescData)
	})
	return file_linter_proto_linter_proto_proto_rawDescData
}

var file_linter_proto_linter_proto_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_linter_proto_linter_proto_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_linter_proto_linter_proto_proto_goTypes = []interface{}{
	(SetConfigResponse_Code)(0),     // 0: SetConfigResponse.Code
	(*LintRequest)(nil),             // 1: LintRequest
	(*LintResponse)(nil),            // 2: LintResponse
	(*SetConfigRequest)(nil),        // 3: SetConfigRequest
	(*SetConfigResponse)(nil),       // 4: SetConfigResponse
	(*LintResponse_Hint)(nil),       // 5: LintResponse.Hint
	(*SetConfigRequest_Weight)(nil), // 6: SetConfigRequest.Weight
	(*SetConfigRequest_Worker)(nil), // 7: SetConfigRequest.Worker
}
var file_linter_proto_linter_proto_proto_depIdxs = []int32{
	5, // 0: LintResponse.hints:type_name -> LintResponse.Hint
	7, // 1: SetConfigRequest.workers:type_name -> SetConfigRequest.Worker
	6, // 2: SetConfigRequest.weights:type_name -> SetConfigRequest.Weight
	0, // 3: SetConfigResponse.code:type_name -> SetConfigResponse.Code
	1, // 4: Linter.Lint:input_type -> LintRequest
	3, // 5: LoadBalancer.SetConfig:input_type -> SetConfigRequest
	2, // 6: Linter.Lint:output_type -> LintResponse
	4, // 7: LoadBalancer.SetConfig:output_type -> SetConfigResponse
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_linter_proto_linter_proto_proto_init() }
func file_linter_proto_linter_proto_proto_init() {
	if File_linter_proto_linter_proto_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_linter_proto_linter_proto_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LintRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_linter_proto_linter_proto_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LintResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_linter_proto_linter_proto_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_linter_proto_linter_proto_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetConfigResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_linter_proto_linter_proto_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LintResponse_Hint); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_linter_proto_linter_proto_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetConfigRequest_Weight); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_linter_proto_linter_proto_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetConfigRequest_Worker); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_linter_proto_linter_proto_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_linter_proto_linter_proto_proto_goTypes,
		DependencyIndexes: file_linter_proto_linter_proto_proto_depIdxs,
		EnumInfos:         file_linter_proto_linter_proto_proto_enumTypes,
		MessageInfos:      file_linter_proto_linter_proto_proto_msgTypes,
	}.Build()
	File_linter_proto_linter_proto_proto = out.File
	file_linter_proto_linter_proto_proto_rawDesc = nil
	file_linter_proto_linter_proto_proto_goTypes = nil
	file_linter_proto_linter_proto_proto_depIdxs = nil
}
