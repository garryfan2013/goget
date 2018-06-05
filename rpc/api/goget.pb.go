// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rpc/api/goget.proto

package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// A job represents a download, provided with neccessary information
type Job struct {
	// The url of the resource to be downloaded
	Url string `protobuf:"bytes,1,opt,name=url" json:"url,omitempty"`
	// The file path to be saved
	Path string `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	// Username for auth, such as ftp download, an account is always needed
	Username string `protobuf:"bytes,3,opt,name=username" json:"username,omitempty"`
	// Password for auth
	Passwd string `protobuf:"bytes,4,opt,name=passwd" json:"passwd,omitempty"`
	// The cnt indicates the thread count for cocurrent downloading
	Cnt                  int64    `protobuf:"varint,5,opt,name=cnt" json:"cnt,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Job) Reset()         { *m = Job{} }
func (m *Job) String() string { return proto.CompactTextString(m) }
func (*Job) ProtoMessage()    {}
func (*Job) Descriptor() ([]byte, []int) {
	return fileDescriptor_goget_675b2d4cb891dbeb, []int{0}
}
func (m *Job) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Job.Unmarshal(m, b)
}
func (m *Job) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Job.Marshal(b, m, deterministic)
}
func (dst *Job) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Job.Merge(dst, src)
}
func (m *Job) XXX_Size() int {
	return xxx_messageInfo_Job.Size(m)
}
func (m *Job) XXX_DiscardUnknown() {
	xxx_messageInfo_Job.DiscardUnknown(m)
}

var xxx_messageInfo_Job proto.InternalMessageInfo

func (m *Job) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *Job) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Job) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *Job) GetPasswd() string {
	if m != nil {
		return m.Passwd
	}
	return ""
}

func (m *Job) GetCnt() int64 {
	if m != nil {
		return m.Cnt
	}
	return 0
}

// A Id represent a specified job
type Id struct {
	// The string format of given uuid
	Uuid                 string   `protobuf:"bytes,1,opt,name=uuid" json:"uuid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Id) Reset()         { *m = Id{} }
func (m *Id) String() string { return proto.CompactTextString(m) }
func (*Id) ProtoMessage()    {}
func (*Id) Descriptor() ([]byte, []int) {
	return fileDescriptor_goget_675b2d4cb891dbeb, []int{1}
}
func (m *Id) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Id.Unmarshal(m, b)
}
func (m *Id) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Id.Marshal(b, m, deterministic)
}
func (dst *Id) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Id.Merge(dst, src)
}
func (m *Id) XXX_Size() int {
	return xxx_messageInfo_Id.Size(m)
}
func (m *Id) XXX_DiscardUnknown() {
	xxx_messageInfo_Id.DiscardUnknown(m)
}

var xxx_messageInfo_Id proto.InternalMessageInfo

func (m *Id) GetUuid() string {
	if m != nil {
		return m.Uuid
	}
	return ""
}

// A Stats consists of progress information for a specified job
type Stats struct {
	// The name of the feature.
	Size int64 `protobuf:"varint,1,opt,name=size" json:"size,omitempty"`
	// The point where the feature is detected.
	Done                 int64    `protobuf:"varint,2,opt,name=done" json:"done,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Stats) Reset()         { *m = Stats{} }
func (m *Stats) String() string { return proto.CompactTextString(m) }
func (*Stats) ProtoMessage()    {}
func (*Stats) Descriptor() ([]byte, []int) {
	return fileDescriptor_goget_675b2d4cb891dbeb, []int{2}
}
func (m *Stats) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Stats.Unmarshal(m, b)
}
func (m *Stats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Stats.Marshal(b, m, deterministic)
}
func (dst *Stats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stats.Merge(dst, src)
}
func (m *Stats) XXX_Size() int {
	return xxx_messageInfo_Stats.Size(m)
}
func (m *Stats) XXX_DiscardUnknown() {
	xxx_messageInfo_Stats.DiscardUnknown(m)
}

var xxx_messageInfo_Stats proto.InternalMessageInfo

func (m *Stats) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *Stats) GetDone() int64 {
	if m != nil {
		return m.Done
	}
	return 0
}

// A Id represent a specified job
type Result struct {
	// The string format of given uuid
	Uuid                 string   `protobuf:"bytes,1,opt,name=uuid" json:"uuid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Result) Reset()         { *m = Result{} }
func (m *Result) String() string { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()    {}
func (*Result) Descriptor() ([]byte, []int) {
	return fileDescriptor_goget_675b2d4cb891dbeb, []int{3}
}
func (m *Result) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Result.Unmarshal(m, b)
}
func (m *Result) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Result.Marshal(b, m, deterministic)
}
func (dst *Result) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Result.Merge(dst, src)
}
func (m *Result) XXX_Size() int {
	return xxx_messageInfo_Result.Size(m)
}
func (m *Result) XXX_DiscardUnknown() {
	xxx_messageInfo_Result.DiscardUnknown(m)
}

var xxx_messageInfo_Result proto.InternalMessageInfo

func (m *Result) GetUuid() string {
	if m != nil {
		return m.Uuid
	}
	return ""
}

func init() {
	proto.RegisterType((*Job)(nil), "api.Job")
	proto.RegisterType((*Id)(nil), "api.Id")
	proto.RegisterType((*Stats)(nil), "api.Stats")
	proto.RegisterType((*Result)(nil), "api.Result")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GoGetClient is the client API for GoGet service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GoGetClient interface {
	// A simple RPC.
	//
	// Add a goget job.
	//
	Add(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Id, error)
	// A server-to-client RPC.
	//
	// Obtains the Current job progress
	//
	Progress(ctx context.Context, in *Id, opts ...grpc.CallOption) (*Stats, error)
	// A client-to-server RPC.
	//
	// Accepts a job id
	//
	Stop(ctx context.Context, in *Id, opts ...grpc.CallOption) (*Stats, error)
}

type goGetClient struct {
	cc *grpc.ClientConn
}

func NewGoGetClient(cc *grpc.ClientConn) GoGetClient {
	return &goGetClient{cc}
}

func (c *goGetClient) Add(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Id, error) {
	out := new(Id)
	err := c.cc.Invoke(ctx, "/api.GoGet/Add", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goGetClient) Progress(ctx context.Context, in *Id, opts ...grpc.CallOption) (*Stats, error) {
	out := new(Stats)
	err := c.cc.Invoke(ctx, "/api.GoGet/Progress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goGetClient) Stop(ctx context.Context, in *Id, opts ...grpc.CallOption) (*Stats, error) {
	out := new(Stats)
	err := c.cc.Invoke(ctx, "/api.GoGet/Stop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoGetServer is the server API for GoGet service.
type GoGetServer interface {
	// A simple RPC.
	//
	// Add a goget job.
	//
	Add(context.Context, *Job) (*Id, error)
	// A server-to-client RPC.
	//
	// Obtains the Current job progress
	//
	Progress(context.Context, *Id) (*Stats, error)
	// A client-to-server RPC.
	//
	// Accepts a job id
	//
	Stop(context.Context, *Id) (*Stats, error)
}

func RegisterGoGetServer(s *grpc.Server, srv GoGetServer) {
	s.RegisterService(&_GoGet_serviceDesc, srv)
}

func _GoGet_Add_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoGetServer).Add(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.GoGet/Add",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoGetServer).Add(ctx, req.(*Job))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoGet_Progress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Id)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoGetServer).Progress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.GoGet/Progress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoGetServer).Progress(ctx, req.(*Id))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoGet_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Id)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoGetServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.GoGet/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoGetServer).Stop(ctx, req.(*Id))
	}
	return interceptor(ctx, in, info, handler)
}

var _GoGet_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.GoGet",
	HandlerType: (*GoGetServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Add",
			Handler:    _GoGet_Add_Handler,
		},
		{
			MethodName: "Progress",
			Handler:    _GoGet_Progress_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _GoGet_Stop_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/api/goget.proto",
}

func init() { proto.RegisterFile("rpc/api/goget.proto", fileDescriptor_goget_675b2d4cb891dbeb) }

var fileDescriptor_goget_675b2d4cb891dbeb = []byte{
	// 245 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0xcd, 0x4a, 0x03, 0x31,
	0x14, 0x85, 0x3b, 0xcd, 0xcc, 0x38, 0xde, 0x95, 0x5c, 0x41, 0xc2, 0xa0, 0x50, 0xb3, 0xea, 0x6a,
	0x06, 0xf4, 0x09, 0x5c, 0x95, 0x76, 0x25, 0xe9, 0x13, 0x64, 0x9a, 0x30, 0x06, 0xea, 0x24, 0xe6,
	0x07, 0xc1, 0xa7, 0x97, 0x5c, 0xaa, 0x2b, 0xdd, 0x7d, 0xf7, 0x7c, 0x70, 0x72, 0x08, 0xdc, 0x06,
	0x7f, 0x1a, 0x95, 0xb7, 0xe3, 0xec, 0x66, 0x93, 0x06, 0x1f, 0x5c, 0x72, 0xc8, 0x94, 0xb7, 0xe2,
	0x03, 0xd8, 0xc1, 0x4d, 0x78, 0x03, 0x2c, 0x87, 0x33, 0xaf, 0x36, 0xd5, 0xf6, 0x5a, 0x16, 0x44,
	0x84, 0xda, 0xab, 0xf4, 0xc6, 0xd7, 0x14, 0x11, 0x63, 0x0f, 0x5d, 0x8e, 0x26, 0x2c, 0xea, 0xdd,
	0x70, 0x46, 0xf9, 0xef, 0x8d, 0x77, 0xd0, 0x7a, 0x15, 0xe3, 0xa7, 0xe6, 0x35, 0x99, 0xcb, 0x55,
	0x9a, 0x4f, 0x4b, 0xe2, 0xcd, 0xa6, 0xda, 0x32, 0x59, 0x50, 0x70, 0x58, 0xef, 0x75, 0xe9, 0xcf,
	0xd9, 0xea, 0xcb, 0x93, 0xc4, 0x62, 0x84, 0xe6, 0x98, 0x54, 0x8a, 0x45, 0x46, 0xfb, 0x65, 0x48,
	0x32, 0x49, 0x5c, 0x32, 0xed, 0x16, 0x43, 0x83, 0x98, 0x24, 0x16, 0xf7, 0xd0, 0x4a, 0x13, 0xf3,
	0x39, 0xfd, 0x55, 0xf7, 0x64, 0xa0, 0xd9, 0xb9, 0x9d, 0x49, 0xd8, 0x03, 0x7b, 0xd1, 0x1a, 0xbb,
	0x41, 0x79, 0x3b, 0x1c, 0xdc, 0xd4, 0x5f, 0x11, 0xed, 0xb5, 0x58, 0xe1, 0x23, 0x74, 0xaf, 0xc1,
	0xcd, 0xc1, 0xc4, 0x88, 0x3f, 0x71, 0x0f, 0x04, 0xb4, 0x45, 0xac, 0xf0, 0x01, 0xea, 0x63, 0x72,
	0xfe, 0x1f, 0x3d, 0xb5, 0xf4, 0x9d, 0xcf, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xe0, 0x29, 0x93,
	0x0c, 0x65, 0x01, 0x00, 0x00,
}
