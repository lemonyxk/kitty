// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hello.proto

package awesomepackage

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

type AwesomeMessage struct {
	AwesomeField         string   `protobuf:"bytes,1,opt,name=awesomeField,proto3" json:"awesomeField,omitempty"`
	AwesomeKey           string   `protobuf:"bytes,2,opt,name=awesomeKey,proto3" json:"awesomeKey,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AwesomeMessage) Reset()         { *m = AwesomeMessage{} }
func (m *AwesomeMessage) String() string { return proto.CompactTextString(m) }
func (*AwesomeMessage) ProtoMessage()    {}
func (*AwesomeMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_61ef911816e0a8ce, []int{0}
}

func (m *AwesomeMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AwesomeMessage.Unmarshal(m, b)
}
func (m *AwesomeMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AwesomeMessage.Marshal(b, m, deterministic)
}
func (m *AwesomeMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AwesomeMessage.Merge(m, src)
}
func (m *AwesomeMessage) XXX_Size() int {
	return xxx_messageInfo_AwesomeMessage.Size(m)
}
func (m *AwesomeMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_AwesomeMessage.DiscardUnknown(m)
}

var xxx_messageInfo_AwesomeMessage proto.InternalMessageInfo

func (m *AwesomeMessage) GetAwesomeField() string {
	if m != nil {
		return m.AwesomeField
	}
	return ""
}

func (m *AwesomeMessage) GetAwesomeKey() string {
	if m != nil {
		return m.AwesomeKey
	}
	return ""
}

func init() {
	proto.RegisterType((*AwesomeMessage)(nil), "awesomepackage.AwesomeMessage")
}

func init() { proto.RegisterFile("hello.proto", fileDescriptor_61ef911816e0a8ce) }

var fileDescriptor_61ef911816e0a8ce = []byte{
	// 107 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xce, 0x48, 0xcd, 0xc9,
	0xc9, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4b, 0x2c, 0x4f, 0x2d, 0xce, 0xcf, 0x4d,
	0x2d, 0x48, 0x4c, 0xce, 0x4e, 0x4c, 0x4f, 0x55, 0x0a, 0xe1, 0xe2, 0x73, 0x84, 0x88, 0xf8, 0xa6,
	0x16, 0x17, 0x27, 0xa6, 0xa7, 0x0a, 0x29, 0x71, 0xf1, 0x40, 0xd5, 0xb8, 0x65, 0xa6, 0xe6, 0xa4,
	0x48, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0xa1, 0x88, 0x09, 0xc9, 0x71, 0x71, 0x41, 0xf9, 0xde,
	0xa9, 0x95, 0x12, 0x4c, 0x60, 0x15, 0x48, 0x22, 0x49, 0x6c, 0x60, 0xcb, 0x8c, 0x01, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x22, 0x34, 0xa1, 0x40, 0x7b, 0x00, 0x00, 0x00,
}
