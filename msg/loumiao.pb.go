// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.11.1
// source: loumiao.proto

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

type LouMiaoWatchKey_OpCode int32

const (
	LouMiaoWatchKey_ADD LouMiaoWatchKey_OpCode = 0 //添加
	LouMiaoWatchKey_DEL LouMiaoWatchKey_OpCode = 1 //删除
)

// Enum value maps for LouMiaoWatchKey_OpCode.
var (
	LouMiaoWatchKey_OpCode_name = map[int32]string{
		0: "ADD",
		1: "DEL",
	}
	LouMiaoWatchKey_OpCode_value = map[string]int32{
		"ADD": 0,
		"DEL": 1,
	}
)

func (x LouMiaoWatchKey_OpCode) Enum() *LouMiaoWatchKey_OpCode {
	p := new(LouMiaoWatchKey_OpCode)
	*p = x
	return p
}

func (x LouMiaoWatchKey_OpCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LouMiaoWatchKey_OpCode) Descriptor() protoreflect.EnumDescriptor {
	return file_loumiao_proto_enumTypes[0].Descriptor()
}

func (LouMiaoWatchKey_OpCode) Type() protoreflect.EnumType {
	return &file_loumiao_proto_enumTypes[0]
}

func (x LouMiaoWatchKey_OpCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LouMiaoWatchKey_OpCode.Descriptor instead.
func (LouMiaoWatchKey_OpCode) EnumDescriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{9, 0}
}

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
		mi := &file_loumiao_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoLoginGate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoLoginGate) ProtoMessage() {}

func (x *LouMiaoLoginGate) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[0]
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
	return file_loumiao_proto_rawDescGZIP(), []int{0}
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

	Uid      int32    `protobuf:"varint,1,opt,name=Uid,proto3" json:"Uid,omitempty"`          //server的uid
	FuncName []string `protobuf:"bytes,2,rep,name=FuncName,proto3" json:"FuncName,omitempty"` //rpc func集合
}

func (x *LouMiaoRpcRegister) Reset() {
	*x = LouMiaoRpcRegister{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoRpcRegister) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoRpcRegister) ProtoMessage() {}

func (x *LouMiaoRpcRegister) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[1]
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
	return file_loumiao_proto_rawDescGZIP(), []int{1}
}

func (x *LouMiaoRpcRegister) GetUid() int32 {
	if x != nil {
		return x.Uid
	}
	return 0
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
		mi := &file_loumiao_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoKickOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoKickOut) ProtoMessage() {}

func (x *LouMiaoKickOut) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[2]
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
	return file_loumiao_proto_rawDescGZIP(), []int{2}
}

type LouMiaoClientConnect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId int64 `protobuf:"varint,1,opt,name=ClientId,proto3" json:"ClientId,omitempty"`
	GateId   int32 `protobuf:"varint,2,opt,name=GateId,proto3" json:"GateId,omitempty"`
	State    int32 `protobuf:"varint,3,opt,name=State,proto3" json:"State,omitempty"` //0:连接，1:断开
}

func (x *LouMiaoClientConnect) Reset() {
	*x = LouMiaoClientConnect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoClientConnect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoClientConnect) ProtoMessage() {}

func (x *LouMiaoClientConnect) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoClientConnect.ProtoReflect.Descriptor instead.
func (*LouMiaoClientConnect) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{3}
}

func (x *LouMiaoClientConnect) GetClientId() int64 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *LouMiaoClientConnect) GetGateId() int32 {
	if x != nil {
		return x.GateId
	}
	return 0
}

func (x *LouMiaoClientConnect) GetState() int32 {
	if x != nil {
		return x.State
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
	Flag     int32  `protobuf:"varint,5,opt,name=Flag,proto3" json:"Flag,omitempty"`
}

func (x *LouMiaoRpcMsg) Reset() {
	*x = LouMiaoRpcMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoRpcMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoRpcMsg) ProtoMessage() {}

func (x *LouMiaoRpcMsg) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[4]
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
	return file_loumiao_proto_rawDescGZIP(), []int{4}
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

func (x *LouMiaoRpcMsg) GetFlag() int32 {
	if x != nil {
		return x.Flag
	}
	return 0
}

type LouMiaoNetMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId int64  `protobuf:"varint,1,opt,name=ClientId,proto3" json:"ClientId,omitempty"`
	Buffer   []byte `protobuf:"bytes,2,opt,name=Buffer,proto3" json:"Buffer,omitempty"`
}

func (x *LouMiaoNetMsg) Reset() {
	*x = LouMiaoNetMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoNetMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoNetMsg) ProtoMessage() {}

func (x *LouMiaoNetMsg) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[5]
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
	return file_loumiao_proto_rawDescGZIP(), []int{5}
}

func (x *LouMiaoNetMsg) GetClientId() int64 {
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

type LouMiaoBindGate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid    int32 `protobuf:"varint,1,opt,name=Uid,proto3" json:"Uid,omitempty"`       //gate的uid
	UserId int64 `protobuf:"varint,2,opt,name=UserId,proto3" json:"UserId,omitempty"` //userid
}

func (x *LouMiaoBindGate) Reset() {
	*x = LouMiaoBindGate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoBindGate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoBindGate) ProtoMessage() {}

func (x *LouMiaoBindGate) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoBindGate.ProtoReflect.Descriptor instead.
func (*LouMiaoBindGate) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{6}
}

func (x *LouMiaoBindGate) GetUid() int32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *LouMiaoBindGate) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

type LouMiaoBroadCastMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     int32  `protobuf:"varint,1,opt,name=Type,proto3" json:"Type,omitempty"` //指定目标服务器类型
	FuncName string `protobuf:"bytes,2,opt,name=FuncName,proto3" json:"FuncName,omitempty"`
	Buffer   []byte `protobuf:"bytes,3,opt,name=Buffer,proto3" json:"Buffer,omitempty"`
}

func (x *LouMiaoBroadCastMsg) Reset() {
	*x = LouMiaoBroadCastMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoBroadCastMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoBroadCastMsg) ProtoMessage() {}

func (x *LouMiaoBroadCastMsg) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoBroadCastMsg.ProtoReflect.Descriptor instead.
func (*LouMiaoBroadCastMsg) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{7}
}

func (x *LouMiaoBroadCastMsg) GetType() int32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *LouMiaoBroadCastMsg) GetFuncName() string {
	if x != nil {
		return x.FuncName
	}
	return ""
}

func (x *LouMiaoBroadCastMsg) GetBuffer() []byte {
	if x != nil {
		return x.Buffer
	}
	return nil
}

type TestEncodeMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     string `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	FuncName string `protobuf:"bytes,2,opt,name=FuncName,proto3" json:"FuncName,omitempty"`
	Buffer   string `protobuf:"bytes,3,opt,name=Buffer,proto3" json:"Buffer,omitempty"`
}

func (x *TestEncodeMsg) Reset() {
	*x = TestEncodeMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestEncodeMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestEncodeMsg) ProtoMessage() {}

func (x *TestEncodeMsg) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestEncodeMsg.ProtoReflect.Descriptor instead.
func (*TestEncodeMsg) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{8}
}

func (x *TestEncodeMsg) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *TestEncodeMsg) GetFuncName() string {
	if x != nil {
		return x.FuncName
	}
	return ""
}

func (x *TestEncodeMsg) GetBuffer() string {
	if x != nil {
		return x.Buffer
	}
	return ""
}

type LouMiaoWatchKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string                 `protobuf:"bytes,1,opt,name=Prefix,proto3" json:"Prefix,omitempty"`
	Opcode LouMiaoWatchKey_OpCode `protobuf:"varint,2,opt,name=opcode,proto3,enum=msg.LouMiaoWatchKey_OpCode" json:"opcode,omitempty"`
}

func (x *LouMiaoWatchKey) Reset() {
	*x = LouMiaoWatchKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoWatchKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoWatchKey) ProtoMessage() {}

func (x *LouMiaoWatchKey) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoWatchKey.ProtoReflect.Descriptor instead.
func (*LouMiaoWatchKey) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{9}
}

func (x *LouMiaoWatchKey) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *LouMiaoWatchKey) GetOpcode() LouMiaoWatchKey_OpCode {
	if x != nil {
		return x.Opcode
	}
	return LouMiaoWatchKey_ADD
}

type LouMiaoPutValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=Prefix,proto3" json:"Prefix,omitempty"`
	Value  string `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"` //value为空，代表删除
}

func (x *LouMiaoPutValue) Reset() {
	*x = LouMiaoPutValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoPutValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoPutValue) ProtoMessage() {}

func (x *LouMiaoPutValue) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoPutValue.ProtoReflect.Descriptor instead.
func (*LouMiaoPutValue) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{10}
}

func (x *LouMiaoPutValue) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *LouMiaoPutValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type LouMiaoNoticeValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=Prefix,proto3" json:"Prefix,omitempty"`
	Value  string `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"` //value为空，代表删除
}

func (x *LouMiaoNoticeValue) Reset() {
	*x = LouMiaoNoticeValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoNoticeValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoNoticeValue) ProtoMessage() {}

func (x *LouMiaoNoticeValue) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoNoticeValue.ProtoReflect.Descriptor instead.
func (*LouMiaoNoticeValue) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{11}
}

func (x *LouMiaoNoticeValue) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *LouMiaoNoticeValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type LouMiaoGetValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix  string   `protobuf:"bytes,1,opt,name=Prefix,proto3" json:"Prefix,omitempty"`
	Prefixs []string `protobuf:"bytes,2,rep,name=Prefixs,proto3" json:"Prefixs,omitempty"` //返回prefix
	Values  []string `protobuf:"bytes,3,rep,name=Values,proto3" json:"Values,omitempty"`   //返回值
}

func (x *LouMiaoGetValue) Reset() {
	*x = LouMiaoGetValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoGetValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoGetValue) ProtoMessage() {}

func (x *LouMiaoGetValue) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoGetValue.ProtoReflect.Descriptor instead.
func (*LouMiaoGetValue) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{12}
}

func (x *LouMiaoGetValue) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *LouMiaoGetValue) GetPrefixs() []string {
	if x != nil {
		return x.Prefixs
	}
	return nil
}

func (x *LouMiaoGetValue) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type LouMiaoAquireLock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix  string `protobuf:"bytes,1,opt,name=Prefix,proto3" json:"Prefix,omitempty"`
	TimeOut int32  `protobuf:"varint,2,opt,name=TimeOut,proto3" json:"TimeOut,omitempty"` //超时时间，毫秒；返回时代表剩余时间，<=0代表锁超时
}

func (x *LouMiaoAquireLock) Reset() {
	*x = LouMiaoAquireLock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoAquireLock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoAquireLock) ProtoMessage() {}

func (x *LouMiaoAquireLock) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoAquireLock.ProtoReflect.Descriptor instead.
func (*LouMiaoAquireLock) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{13}
}

func (x *LouMiaoAquireLock) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *LouMiaoAquireLock) GetTimeOut() int32 {
	if x != nil {
		return x.TimeOut
	}
	return 0
}

type LouMiaoReleaseLock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=Prefix,proto3" json:"Prefix,omitempty"`
}

func (x *LouMiaoReleaseLock) Reset() {
	*x = LouMiaoReleaseLock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoReleaseLock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoReleaseLock) ProtoMessage() {}

func (x *LouMiaoReleaseLock) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoReleaseLock.ProtoReflect.Descriptor instead.
func (*LouMiaoReleaseLock) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{14}
}

func (x *LouMiaoReleaseLock) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

type LouMiaoLease struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid int32 `protobuf:"varint,1,opt,name=Uid,proto3" json:"Uid,omitempty"` //服务器的uid
}

func (x *LouMiaoLease) Reset() {
	*x = LouMiaoLease{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loumiao_proto_msgTypes[15]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LouMiaoLease) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LouMiaoLease) ProtoMessage() {}

func (x *LouMiaoLease) ProtoReflect() protoreflect.Message {
	mi := &file_loumiao_proto_msgTypes[15]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LouMiaoLease.ProtoReflect.Descriptor instead.
func (*LouMiaoLease) Descriptor() ([]byte, []int) {
	return file_loumiao_proto_rawDescGZIP(), []int{15}
}

func (x *LouMiaoLease) GetUid() int32 {
	if x != nil {
		return x.Uid
	}
	return 0
}

var File_loumiao_proto protoreflect.FileDescriptor

var file_loumiao_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6c, 0x6f, 0x75, 0x6d, 0x69, 0x61, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x03, 0x6d, 0x73, 0x67, 0x22, 0x44, 0x0a, 0x10, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x47, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x42, 0x0a, 0x12, 0x4c, 0x6f,
	0x75, 0x4d, 0x69, 0x61, 0x6f, 0x52, 0x70, 0x63, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x12, 0x10, 0x0a, 0x03, 0x55, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x55,
	0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x10,
	0x0a, 0x0e, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4b, 0x69, 0x63, 0x6b, 0x4f, 0x75, 0x74,
	0x22, 0x60, 0x0a, 0x14, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x47, 0x61, 0x74, 0x65, 0x49, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x47, 0x61, 0x74, 0x65, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x22, 0x8f, 0x01, 0x0a, 0x0d, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x52, 0x70,
	0x63, 0x4d, 0x73, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x64,
	0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x42, 0x75,
	0x66, 0x66, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x46, 0x6c, 0x61, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04,
	0x46, 0x6c, 0x61, 0x67, 0x22, 0x43, 0x0a, 0x0d, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4e,
	0x65, 0x74, 0x4d, 0x73, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49,
	0x64, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x22, 0x3b, 0x0a, 0x0f, 0x4c, 0x6f, 0x75,
	0x4d, 0x69, 0x61, 0x6f, 0x42, 0x69, 0x6e, 0x64, 0x47, 0x61, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x55, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x55, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x5d, 0x0a, 0x13, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61,
	0x6f, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x43, 0x61, 0x73, 0x74, 0x4d, 0x73, 0x67, 0x12, 0x12, 0x0a,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x46, 0x75, 0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a,
	0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x42,
	0x75, 0x66, 0x66, 0x65, 0x72, 0x22, 0x57, 0x0a, 0x0d, 0x54, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x63,
	0x6f, 0x64, 0x65, 0x4d, 0x73, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75,
	0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x46, 0x75,
	0x6e, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x22, 0x7a,
	0x0a, 0x0f, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x57, 0x61, 0x74, 0x63, 0x68, 0x4b, 0x65,
	0x79, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x33, 0x0a, 0x06, 0x6f, 0x70, 0x63,
	0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x6d, 0x73, 0x67, 0x2e,
	0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x57, 0x61, 0x74, 0x63, 0x68, 0x4b, 0x65, 0x79, 0x2e,
	0x4f, 0x70, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x06, 0x6f, 0x70, 0x63, 0x6f, 0x64, 0x65, 0x22, 0x1a,
	0x0a, 0x06, 0x4f, 0x70, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x44, 0x44, 0x10,
	0x00, 0x12, 0x07, 0x0a, 0x03, 0x44, 0x45, 0x4c, 0x10, 0x01, 0x22, 0x3f, 0x0a, 0x0f, 0x4c, 0x6f,
	0x75, 0x4d, 0x69, 0x61, 0x6f, 0x50, 0x75, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a,
	0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x50,
	0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x42, 0x0a, 0x12, 0x4c,
	0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4e, 0x6f, 0x74, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22,
	0x5b, 0x0a, 0x0f, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x47, 0x65, 0x74, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x50, 0x72,
	0x65, 0x66, 0x69, 0x78, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x50, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x45, 0x0a, 0x11,
	0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x41, 0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x6f, 0x63,
	0x6b, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x54, 0x69, 0x6d,
	0x65, 0x4f, 0x75, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x54, 0x69, 0x6d, 0x65,
	0x4f, 0x75, 0x74, 0x22, 0x2c, 0x0a, 0x12, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x52, 0x65,
	0x6c, 0x65, 0x61, 0x73, 0x65, 0x4c, 0x6f, 0x63, 0x6b, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x50, 0x72, 0x65, 0x66, 0x69,
	0x78, 0x22, 0x20, 0x0a, 0x0c, 0x4c, 0x6f, 0x75, 0x4d, 0x69, 0x61, 0x6f, 0x4c, 0x65, 0x61, 0x73,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x55, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03,
	0x55, 0x69, 0x64, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x73, 0x6e, 0x6f, 0x77, 0x79, 0x79, 0x6a, 0x30, 0x30, 0x31, 0x2f, 0x6c, 0x6f, 0x75,
	0x6d, 0x69, 0x61, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_loumiao_proto_rawDescOnce sync.Once
	file_loumiao_proto_rawDescData = file_loumiao_proto_rawDesc
)

func file_loumiao_proto_rawDescGZIP() []byte {
	file_loumiao_proto_rawDescOnce.Do(func() {
		file_loumiao_proto_rawDescData = protoimpl.X.CompressGZIP(file_loumiao_proto_rawDescData)
	})
	return file_loumiao_proto_rawDescData
}

var file_loumiao_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_loumiao_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_loumiao_proto_goTypes = []interface{}{
	(LouMiaoWatchKey_OpCode)(0),  // 0: msg.LouMiaoWatchKey.OpCode
	(*LouMiaoLoginGate)(nil),     // 1: msg.LouMiaoLoginGate
	(*LouMiaoRpcRegister)(nil),   // 2: msg.LouMiaoRpcRegister
	(*LouMiaoKickOut)(nil),       // 3: msg.LouMiaoKickOut
	(*LouMiaoClientConnect)(nil), // 4: msg.LouMiaoClientConnect
	(*LouMiaoRpcMsg)(nil),        // 5: msg.LouMiaoRpcMsg
	(*LouMiaoNetMsg)(nil),        // 6: msg.LouMiaoNetMsg
	(*LouMiaoBindGate)(nil),      // 7: msg.LouMiaoBindGate
	(*LouMiaoBroadCastMsg)(nil),  // 8: msg.LouMiaoBroadCastMsg
	(*TestEncodeMsg)(nil),        // 9: msg.TestEncodeMsg
	(*LouMiaoWatchKey)(nil),      // 10: msg.LouMiaoWatchKey
	(*LouMiaoPutValue)(nil),      // 11: msg.LouMiaoPutValue
	(*LouMiaoNoticeValue)(nil),   // 12: msg.LouMiaoNoticeValue
	(*LouMiaoGetValue)(nil),      // 13: msg.LouMiaoGetValue
	(*LouMiaoAquireLock)(nil),    // 14: msg.LouMiaoAquireLock
	(*LouMiaoReleaseLock)(nil),   // 15: msg.LouMiaoReleaseLock
	(*LouMiaoLease)(nil),         // 16: msg.LouMiaoLease
}
var file_loumiao_proto_depIdxs = []int32{
	0, // 0: msg.LouMiaoWatchKey.opcode:type_name -> msg.LouMiaoWatchKey.OpCode
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_loumiao_proto_init() }
func file_loumiao_proto_init() {
	if File_loumiao_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_loumiao_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_loumiao_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_loumiao_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_loumiao_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoClientConnect); i {
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
		file_loumiao_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
		file_loumiao_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
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
		file_loumiao_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoBindGate); i {
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
		file_loumiao_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoBroadCastMsg); i {
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
		file_loumiao_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestEncodeMsg); i {
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
		file_loumiao_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoWatchKey); i {
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
		file_loumiao_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoPutValue); i {
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
		file_loumiao_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoNoticeValue); i {
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
		file_loumiao_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoGetValue); i {
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
		file_loumiao_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoAquireLock); i {
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
		file_loumiao_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoReleaseLock); i {
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
		file_loumiao_proto_msgTypes[15].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LouMiaoLease); i {
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
			RawDescriptor: file_loumiao_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_loumiao_proto_goTypes,
		DependencyIndexes: file_loumiao_proto_depIdxs,
		EnumInfos:         file_loumiao_proto_enumTypes,
		MessageInfos:      file_loumiao_proto_msgTypes,
	}.Build()
	File_loumiao_proto = out.File
	file_loumiao_proto_rawDesc = nil
	file_loumiao_proto_goTypes = nil
	file_loumiao_proto_depIdxs = nil
}
