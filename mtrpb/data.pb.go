// Code generated by protoc-gen-go.
// source: data.proto
// DO NOT EDIT!

/*
Package mtrpb is a generated protocol buffer package.

It is generated from these files:
	data.proto
	field.proto
	tag.proto

It has these top-level messages:
	DataLatencySummary
	DataLatencySummaryResult
	FieldMetricSummary
	FieldMetricSummaryResult
	FieldMetricTag
	FieldMetricTagResult
	Tag
	TagResult
	TagSearchResult
*/
package mtrpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

// DataLatencySummary is a summary of data latency metrics for each site.
// mean should not be 0.  fifty and ninety may be unknown (0).
// If upper == lower == 0 then no threshold has been set on the metric.
type DataLatencySummary struct {
	// The siteID for the metric e.g., TAUP
	SiteID string `protobuf:"bytes,1,opt,name=site_iD,json=siteID" json:"site_iD,omitempty"`
	// The typeID for the metric e.g., latency.strong
	TypeID string `protobuf:"bytes,2,opt,name=type_iD,json=typeID" json:"type_iD,omitempty"`
	// Unix time in seconds for the metric value (don't need nanos).
	Seconds int64 `protobuf:"varint,3,opt,name=seconds" json:"seconds,omitempty"`
	// The mean latency
	Mean int32 `protobuf:"varint,4,opt,name=mean" json:"mean,omitempty"`
	// The fiftieth percentile value.  Might be unknown (0)
	Fifty int32 `protobuf:"varint,5,opt,name=fifty" json:"fifty,omitempty"`
	// The ninetieth percentile value.  Might be unknown (0)
	Ninety int32 `protobuf:"varint,6,opt,name=ninety" json:"ninety,omitempty"`
	// The upper threshold for the metric to be good.
	Upper int32 `protobuf:"varint,7,opt,name=upper" json:"upper,omitempty"`
	// The lower threshold for the metric to be good.
	Lower int32 `protobuf:"varint,8,opt,name=lower" json:"lower,omitempty"`
}

func (m *DataLatencySummary) Reset()                    { *m = DataLatencySummary{} }
func (m *DataLatencySummary) String() string            { return proto.CompactTextString(m) }
func (*DataLatencySummary) ProtoMessage()               {}
func (*DataLatencySummary) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type DataLatencySummaryResult struct {
	Result []*DataLatencySummary `protobuf:"bytes,1,rep,name=result" json:"result,omitempty"`
}

func (m *DataLatencySummaryResult) Reset()                    { *m = DataLatencySummaryResult{} }
func (m *DataLatencySummaryResult) String() string            { return proto.CompactTextString(m) }
func (*DataLatencySummaryResult) ProtoMessage()               {}
func (*DataLatencySummaryResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *DataLatencySummaryResult) GetResult() []*DataLatencySummary {
	if m != nil {
		return m.Result
	}
	return nil
}

func init() {
	proto.RegisterType((*DataLatencySummary)(nil), "mtrpb.DataLatencySummary")
	proto.RegisterType((*DataLatencySummaryResult)(nil), "mtrpb.DataLatencySummaryResult")
}

var fileDescriptor0 = []byte{
	// 226 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x15, 0xd2, 0x38, 0x70, 0x6c, 0x27, 0x04, 0x66, 0x43, 0x9d, 0x3a, 0x45, 0x02, 0xde,
	0x00, 0x65, 0x41, 0x82, 0x25, 0x6c, 0x2c, 0xc8, 0x6d, 0x5d, 0x29, 0x52, 0x63, 0x5b, 0xce, 0x45,
	0xc8, 0x2f, 0xc9, 0x33, 0x71, 0x3e, 0x97, 0x89, 0x6e, 0xf7, 0x7d, 0xff, 0xdd, 0x70, 0x3f, 0xc0,
	0xde, 0x90, 0xe9, 0x42, 0xf4, 0xe4, 0xb1, 0x99, 0x28, 0x86, 0xed, 0xfa, 0xa7, 0x02, 0xec, 0xd9,
	0xbe, 0x19, 0xb2, 0x6e, 0x97, 0x3e, 0x96, 0x69, 0x32, 0x31, 0xe1, 0x1d, 0xb4, 0xf3, 0x48, 0xf6,
	0x6b, 0xec, 0x75, 0xf5, 0x50, 0x6d, 0xae, 0x06, 0x95, 0xf1, 0xb5, 0xcf, 0x01, 0xa5, 0x20, 0xc1,
	0x45, 0x09, 0x32, 0x72, 0xa0, 0xf9, 0xc2, 0xee, 0xbc, 0xdb, 0xcf, 0xba, 0xe6, 0xa0, 0x1e, 0xfe,
	0x10, 0x11, 0x56, 0x93, 0x35, 0x4e, 0xaf, 0x58, 0x37, 0x83, 0xcc, 0x78, 0x03, 0xcd, 0x61, 0x3c,
	0x50, 0xd2, 0x8d, 0xc8, 0x02, 0x78, 0x0b, 0xca, 0x8d, 0xce, 0xb2, 0x56, 0xa2, 0x4f, 0x94, 0xb7,
	0x97, 0x10, 0x6c, 0xd4, 0x6d, 0xd9, 0x16, 0xc8, 0xf6, 0xe8, 0xbf, 0xd9, 0x5e, 0x16, 0x2b, 0xb0,
	0x7e, 0x07, 0xfd, 0xff, 0x9f, 0xc1, 0xce, 0xcb, 0x91, 0xf0, 0x11, 0x54, 0x94, 0x89, 0x9f, 0xaa,
	0x37, 0xd7, 0x4f, 0xf7, 0x9d, 0x94, 0xd0, 0x9d, 0x39, 0x38, 0x2d, 0xbe, 0xb4, 0x9f, 0xa5, 0xa8,
	0xad, 0x92, 0xda, 0x9e, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x7b, 0x90, 0x91, 0xba, 0x44, 0x01,
	0x00, 0x00,
}
