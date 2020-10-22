// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.11.1
// source: pbmsg/loumiao.proto

package msg

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type LouMiaoLoginGate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TokenId int64 `protobuf:"varint,1,opt,name=TokenId,proto3" json:"TokenId,omitempty"`
	UserId  int64 `protobuf:"varint,2,opt,name=UserId,proto3" json:"UserId,omitempty"`
}

func (x *LouMiaoLoginGate) Reset() {
	*x = LouMiaoLoginGate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbmsg_loumiao_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoLoginGate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoLoginGate) ProtoMessage() {}

func (x *LouMiaoLoginGate) ProtoReflect() protoreflect.Message {
	mi := &file_pbmsg_loumiao_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoLoginGate.ProtoReflect.Descriptor instead.
func (*LouMiaoLoginGate) Descriptor() ([]byte, []int) {
	return file_pbmsg_loumiao_proto_rawDescGZIP(), []int{0}
}

func (x *LouMiaoLoginGate) GetTokenId() int64 {
	if x != nil {
		return x.TokenId
	}
	return 0
}

func (x *LouMiaoLoginGate) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

type LouMiaoRpcRegister struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FuncName []string `protobuf:"bytes,1,rep,name=FuncName,proto3" json:"FuncName,omitempty"`
}

func (x *LouMiaoRpcRegister) Reset() {
	*x = LouMiaoRpcRegister{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbmsg_loumiao_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoRpcRegister) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoRpcRegister) ProtoMessage() {}

func (x *LouMiaoRpcRegister) ProtoReflect() protoreflect.Message {
	mi := &file_pbmsg_loumiao_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoRpcRegister.ProtoReflect.Descriptor instead.
func (*LouMiaoRpcRegister) Descriptor() ([]byte, []int) {
	return file_pbmsg_loumiao_proto_rawDescGZIP(), []int{1}
}

func (x *LouMiaoRpcRegister) GetFuncName() []string {
	if x != nil {
		return x.FuncName
	}
	return nil
}

type LouMiaoKickOut struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LouMiaoKickOut) Reset() {
	*x = LouMiaoKickOut{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbmsg_loumiao_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoKickOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoKickOut) ProtoMessage() {}

func (x *LouMiaoKickOut) ProtoReflect() protoreflect.Message {
	mi := &file_pbmsg_loumiao_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoKickOut.ProtoReflect.Descriptor instead.
func (*LouMiaoKickOut) Descriptor() ([]byte, []int) {
	return file_pbmsg_loumiao_proto_rawDescGZIP(), []int{2}
}

type LouMiaoClientOffline struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId int32 `protobuf:"varint,1,opt,name=ClientId,proto3" json:"ClientId,omitempty"`
}

func (x *LouMiaoClientOffline) Reset() {
	*x = LouMiaoClientOffline{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbmsg_loumiao_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoClientOffline) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoClientOffline) ProtoMessage() {}

func (x *LouMiaoClientOffline) ProtoReflect() protoreflect.Message {
	mi := &file_pbmsg_loumiao_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoClientOffline.ProtoReflect.Descriptor instead.
func (*LouMiaoClientOffline) Descriptor() ([]byte, []int) {
	return file_pbmsg_loumiao_proto_rawDescGZIP(), []int{3}
}

func (x *LouMiaoClientOffline) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

type LouMiaoRpcMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TargetId int32  `protobuf:"varint,1,opt,name=TargetId,proto3" json:"TargetId,omitempty"` //>0指定目标服务器uid
	FuncName string `protobuf:"bytes,2,opt,name=FuncName,proto3" json:"FuncName,omitempty"`
	Buffer   []byte `protobuf:"bytes,3,opt,name=Buffer,proto3" json:"Buffer,omitempty"`
	SourceId int32  `protobuf:"varint,4,opt,name=SourceId,proto3" json:"SourceId,omitempty"` //>0指定源服务器uid
}

func (x *LouMiaoRpcMsg) Reset() {
	*x = LouMiaoRpcMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbmsg_loumiao_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoRpcMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoRpcMsg) ProtoMessage() {}

func (x *LouMiaoRpcMsg) ProtoReflect() protoreflect.Message {
	mi := &file_pbmsg_loumiao_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoRpcMsg.ProtoReflect.Descriptor instead.
func (*LouMiaoRpcMsg) Descriptor() ([]byte, []int) {
	return file_pbmsg_loumiao_proto_rawDescGZIP(), []int{4}
}

func (x *LouMiaoRpcMsg) GetTargetId() int32 {
	if x != nil {
		return x.TargetId
	}
	return 0
}

func (x *LouMiaoRpcMsg) GetFuncName() string {
	if x != nil {
		return x.FuncName
	}
	return ""
}

func (x *LouMiaoRpcMsg) GetBuffer() []byte {
	if x != nil {
		return x.Buffer
	}
	return nil
}

func (x *LouMiaoRpcMsg) GetSourceId() int32 {
	if x != nil {
		return x.SourceId
	}
	return 0
}

type LouMiaoNetMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId int32  `protobuf:"varint,1,opt,name=ClientId,proto3" json:"ClientId,omitempty"`
	Buffer   []byte `protobuf:"bytes,2,opt,name=Buffer,proto3" json:"Buffer,omitempty"`
}

func (x *LouMiaoNetMsg) Reset() {
	*x = LouMiaoNetMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pbmsg_loumiao_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoNetMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoNetMsg) ProtoMessage() {}

func (x *LouMiaoNetMsg) ProtoReflect() protoreflect.Message {
	mi := &file_pbmsg_loumiao_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoNetMsg.ProtoReflect.Descriptor instead.
func (*LouMiaoNetMsg) Descriptor() ([]byte, []int) {
	return file_pbmsg_loumiao_proto_rawDescGZIP(), []int{5}
}

func (x *LouMiaoNetMsg) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *LouMiaoNetMsg) GetBuffer() []byte {
	if x != nil {
		return x.Buffer
	}
	return nil
}

var File_pbmsg_loumiao_proto protoreflect.FileDescriptor

var file_pbmsg_loumiao_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x62, 0x6d, 0x73, 0x67, 0x2f, 0x6c, 0x6f, 0x75, 0x6d, 0x69, 0x61, 0x6f, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x44, 0x0a, 0x10, 0x4c, 0x6f,
	0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x47, 0x61, 0x74, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x07, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72,
	0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x22, 0x30, 0x0a, 0x12, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x52, 0x70, 0x63, 0x52, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61,
	0x6d, 0x65, 0x22, 0x10, 0x0a, 0x0e, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4b, 0x69, 0x63,
	0x6b, 0x4f, 0x75, 0x74, 0x22, 0x32, 0x0a, 0x14, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x43,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4f, 0x66, 0x66, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x22, 0x7b, 0x0a, 0x0d, 0x4c, 0x6f, 0x75, 0x4d,
	0x69, 0x61, 0x6f, 0x52, 0x70, 0x63, 0x4d, 0x73, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x54, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x54, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x49, 0x64, 0x22, 0x43, 0x0a, 0x0d, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f,
	0x4e, 0x65, 0x74, 0x4d, 0x73, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6e, 0x6f, 0x77, 0x79, 0x79, 0x6a,
	0x30, 0x30, 0x31, 0x2f, 0x6c, 0x6f, 0x75, 0x6d, 0x69, 0x61, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pbmsg_loumiao_proto_rawDescOnce sync.Once
	file_pbmsg_loumiao_proto_rawDescData = file_pbmsg_loumiao_proto_rawDesc
)

func file_pbmsg_loumiao_proto_rawDescGZIP() []byte {
	file_pbmsg_loumiao_proto_rawDescOnce.Do(func() {
		file_pbmsg_loumiao_proto_rawDescData = protoimpl.X.CompressGZIP(file_pbmsg_loumiao_proto_rawDescData)
	})
	return file_pbmsg_loumiao_proto_rawDescData
}

var file_pbmsg_loumiao_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_pbmsg_loumiao_proto_goTypes = []interface{}{
	(*LouMiaoLoginGate)(nil),     // 0: msg.LouMiaoLoginGate
	(*LouMiaoRpcRegister)(nil),   // 1: msg.LouMiaoRpcRegister
	(*LouMiaoKickOut)(nil),       // 2: msg.LouMiaoKickOut
	(*LouMiaoClientOffline)(nil), // 3: msg.LouMiaoClientOffline
	(*LouMiaoRpcMsg)(nil),        // 4: msg.LouMiaoRpcMsg
	(*LouMiaoNetMsg)(nil),        // 5: msg.LouMiaoNetMsg
}
var file_pbmsg_loumiao_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pbmsg_loumiao_proto_init() }
func file_pbmsg_loumiao_proto_init() {
	if File_pbmsg_loumiao_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pbmsg_loumiao_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoLoginGate); i {
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
		file_pbmsg_loumiao_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoRpcRegister); i {
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
		file_pbmsg_loumiao_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoKickOut); i {
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
		file_pbmsg_loumiao_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoClientOffline); i {
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
		file_pbmsg_loumiao_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoRpcMsg); i {
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
		file_pbmsg_loumiao_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoNetMsg); i {
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
			RawDescriptor: file_pbmsg_loumiao_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pbmsg_loumiao_proto_goTypes,
		DependencyIndexes: file_pbmsg_loumiao_proto_depIdxs,
		MessageInfos:      file_pbmsg_loumiao_proto_msgTypes,
	}.Build()
	File_pbmsg_loumiao_proto = out.File
	file_pbmsg_loumiao_proto_rawDesc = nil
	file_pbmsg_loumiao_proto_goTypes = nil
	file_pbmsg_loumiao_proto_depIdxs = nil
}
