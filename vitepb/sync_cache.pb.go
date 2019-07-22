// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vitepb/sync_cache.proto

package vitepb

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

type CacheItem struct {
	From                 uint64        `protobuf:"varint,1,opt,name=from,proto3" json:"from,omitempty"`
	To                   uint64        `protobuf:"varint,2,opt,name=to,proto3" json:"to,omitempty"`
	PrevHash             []byte        `protobuf:"bytes,3,opt,name=prevHash,proto3" json:"prevHash,omitempty"`
	Hash                 []byte        `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	Points               []*HashHeight `protobuf:"bytes,5,rep,name=points,proto3" json:"points,omitempty"`
	Verified             bool          `protobuf:"varint,6,opt,name=verified,proto3" json:"verified,omitempty"`
	Filename             string        `protobuf:"bytes,7,opt,name=filename,proto3" json:"filename,omitempty"`
	Done                 bool          `protobuf:"varint,8,opt,name=done,proto3" json:"done,omitempty"`
	Size                 int64         `protobuf:"varint,9,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *CacheItem) Reset()         { *m = CacheItem{} }
func (m *CacheItem) String() string { return proto.CompactTextString(m) }
func (*CacheItem) ProtoMessage()    {}
func (*CacheItem) Descriptor() ([]byte, []int) {
	return fileDescriptor_1f427b08fc6065e7, []int{0}
}

func (m *CacheItem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CacheItem.Unmarshal(m, b)
}
func (m *CacheItem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CacheItem.Marshal(b, m, deterministic)
}
func (m *CacheItem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CacheItem.Merge(m, src)
}
func (m *CacheItem) XXX_Size() int {
	return xxx_messageInfo_CacheItem.Size(m)
}
func (m *CacheItem) XXX_DiscardUnknown() {
	xxx_messageInfo_CacheItem.DiscardUnknown(m)
}

var xxx_messageInfo_CacheItem proto.InternalMessageInfo

func (m *CacheItem) GetFrom() uint64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *CacheItem) GetTo() uint64 {
	if m != nil {
		return m.To
	}
	return 0
}

func (m *CacheItem) GetPrevHash() []byte {
	if m != nil {
		return m.PrevHash
	}
	return nil
}

func (m *CacheItem) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *CacheItem) GetPoints() []*HashHeight {
	if m != nil {
		return m.Points
	}
	return nil
}

func (m *CacheItem) GetVerified() bool {
	if m != nil {
		return m.Verified
	}
	return false
}

func (m *CacheItem) GetFilename() string {
	if m != nil {
		return m.Filename
	}
	return ""
}

func (m *CacheItem) GetDone() bool {
	if m != nil {
		return m.Done
	}
	return false
}

func (m *CacheItem) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

type CacheItems struct {
	Items                []*CacheItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *CacheItems) Reset()         { *m = CacheItems{} }
func (m *CacheItems) String() string { return proto.CompactTextString(m) }
func (*CacheItems) ProtoMessage()    {}
func (*CacheItems) Descriptor() ([]byte, []int) {
	return fileDescriptor_1f427b08fc6065e7, []int{1}
}

func (m *CacheItems) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CacheItems.Unmarshal(m, b)
}
func (m *CacheItems) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CacheItems.Marshal(b, m, deterministic)
}
func (m *CacheItems) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CacheItems.Merge(m, src)
}
func (m *CacheItems) XXX_Size() int {
	return xxx_messageInfo_CacheItems.Size(m)
}
func (m *CacheItems) XXX_DiscardUnknown() {
	xxx_messageInfo_CacheItems.DiscardUnknown(m)
}

var xxx_messageInfo_CacheItems proto.InternalMessageInfo

func (m *CacheItems) GetItems() []*CacheItem {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
	proto.RegisterType((*CacheItem)(nil), "vitepb.cacheItem")
	proto.RegisterType((*CacheItems)(nil), "vitepb.cacheItems")
}

func init() { proto.RegisterFile("vitepb/sync_cache.proto", fileDescriptor_1f427b08fc6065e7) }

var fileDescriptor_1f427b08fc6065e7 = []byte{
	// 245 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x50, 0x4d, 0x4a, 0xc4, 0x30,
	0x14, 0x26, 0x6d, 0xa7, 0xb6, 0x4f, 0x11, 0x0c, 0x82, 0x8f, 0x59, 0x85, 0xd9, 0x18, 0x5c, 0x54,
	0x50, 0xbc, 0xc3, 0xb8, 0xcd, 0x05, 0xa4, 0xd3, 0x79, 0x9d, 0x06, 0x6c, 0x53, 0x9a, 0x50, 0xd0,
	0x2b, 0x7b, 0x09, 0xc9, 0x6b, 0xa7, 0xbb, 0xef, 0x37, 0xf9, 0x12, 0x78, 0x9a, 0x6d, 0xa0, 0xf1,
	0xf4, 0xea, 0x7f, 0x86, 0xe6, 0xab, 0xa9, 0x9b, 0x8e, 0xaa, 0x71, 0x72, 0xc1, 0xc9, 0x7c, 0x31,
	0xf6, 0x8f, 0x6b, 0xa0, 0x27, 0xef, 0xeb, 0xcb, 0xea, 0x1e, 0xfe, 0x04, 0x94, 0x9c, 0xfe, 0x0c,
	0xd4, 0x4b, 0x09, 0x59, 0x3b, 0xb9, 0x1e, 0x85, 0x12, 0x3a, 0x33, 0x8c, 0xe5, 0x3d, 0x24, 0xc1,
	0x61, 0xc2, 0x4a, 0x12, 0x9c, 0xdc, 0x43, 0x31, 0x4e, 0x34, 0x1f, 0x6b, 0xdf, 0x61, 0xaa, 0x84,
	0xbe, 0x33, 0x1b, 0x8f, 0xfd, 0x2e, 0xea, 0x19, 0xeb, 0x8c, 0xe5, 0x0b, 0xe4, 0xa3, 0xb3, 0x43,
	0xf0, 0xb8, 0x53, 0xa9, 0xbe, 0x7d, 0x93, 0xd5, 0x32, 0xa4, 0x8a, 0x8d, 0x23, 0xd9, 0x4b, 0x17,
	0xcc, 0x9a, 0x88, 0x67, 0xcf, 0x34, 0xd9, 0xd6, 0xd2, 0x19, 0x73, 0x25, 0x74, 0x61, 0x36, 0x1e,
	0xbd, 0xd6, 0x7e, 0xd3, 0x50, 0xf7, 0x84, 0x37, 0x4a, 0xe8, 0xd2, 0x6c, 0x3c, 0xde, 0x7b, 0x76,
	0x03, 0x61, 0xc1, 0x1d, 0xc6, 0x51, 0xf3, 0xf6, 0x97, 0xb0, 0x54, 0x42, 0xa7, 0x86, 0xf1, 0xe1,
	0x03, 0x60, 0x7b, 0xac, 0x97, 0xcf, 0xb0, 0xb3, 0x11, 0xa0, 0xe0, 0x61, 0x0f, 0xd7, 0x61, 0x5b,
	0xc4, 0x2c, 0xfe, 0x29, 0xe7, 0xbf, 0x7a, 0xff, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x90, 0xf9, 0xf7,
	0xd9, 0x64, 0x01, 0x00, 0x00,
}