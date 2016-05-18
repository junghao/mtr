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
	DataSite
	DataSiteResult
	DataLatencyTag
	DataLatencyTagResult
	DataLatencyThreshold
	DataLatencyThresholdResult
	FieldMetricSummary
	FieldMetricSummaryResult
	FieldMetricTag
	FieldMetricTagResult
	FieldMetricThreshold
	FieldMetricThresholdResult
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

type DataSite struct {
	// The siteID for the metric e.g., TAUP
	SiteID string `protobuf:"bytes,1,opt,name=site_iD,json=siteID" json:"site_iD,omitempty"`
	// The site latitude - not usually accurate enough for meta data
	Latitude float64 `protobuf:"fixed64,2,opt,name=latitude" json:"latitude,omitempty"`
	// The site longitude - not usually accurate enough for meta data
	Longitude float64 `protobuf:"fixed64,3,opt,name=longitude" json:"longitude,omitempty"`
}

func (m *DataSite) Reset()                    { *m = DataSite{} }
func (m *DataSite) String() string            { return proto.CompactTextString(m) }
func (*DataSite) ProtoMessage()               {}
func (*DataSite) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type DataSiteResult struct {
	Result []*DataSite `protobuf:"bytes,1,rep,name=result" json:"result,omitempty"`
}

func (m *DataSiteResult) Reset()                    { *m = DataSiteResult{} }
func (m *DataSiteResult) String() string            { return proto.CompactTextString(m) }
func (*DataSiteResult) ProtoMessage()               {}
func (*DataSiteResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *DataSiteResult) GetResult() []*DataSite {
	if m != nil {
		return m.Result
	}
	return nil
}

type DataLatencyTag struct {
	// The siteID for the latency e.g., TAUP
	SiteID string `protobuf:"bytes,1,opt,name=site_iD,json=siteID" json:"site_iD,omitempty"`
	// The typeID for the latency e.g., latency.gnss.1hz
	TypeID string `protobuf:"bytes,2,opt,name=type_iD,json=typeID" json:"type_iD,omitempty"`
	// The tag for the latency e.g., TAUP
	Tag string `protobuf:"bytes,3,opt,name=tag" json:"tag,omitempty"`
}

func (m *DataLatencyTag) Reset()                    { *m = DataLatencyTag{} }
func (m *DataLatencyTag) String() string            { return proto.CompactTextString(m) }
func (*DataLatencyTag) ProtoMessage()               {}
func (*DataLatencyTag) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type DataLatencyTagResult struct {
	Result []*DataLatencyTag `protobuf:"bytes,1,rep,name=result" json:"result,omitempty"`
}

func (m *DataLatencyTagResult) Reset()                    { *m = DataLatencyTagResult{} }
func (m *DataLatencyTagResult) String() string            { return proto.CompactTextString(m) }
func (*DataLatencyTagResult) ProtoMessage()               {}
func (*DataLatencyTagResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *DataLatencyTagResult) GetResult() []*DataLatencyTag {
	if m != nil {
		return m.Result
	}
	return nil
}

type DataLatencyThreshold struct {
	// The siteID for the latency e.g., TAUP
	SiteID string `protobuf:"bytes,1,opt,name=site_iD,json=siteID" json:"site_iD,omitempty"`
	// The typeID for the latency e.g., latency.gnss.1hz
	TypeID string `protobuf:"bytes,2,opt,name=type_iD,json=typeID" json:"type_iD,omitempty"`
	// The lower threshold for the latency to be good.
	Lower int32 `protobuf:"varint,3,opt,name=lower" json:"lower,omitempty"`
	// The upper threshold for the latency to be good.
	Upper int32 `protobuf:"varint,4,opt,name=upper" json:"upper,omitempty"`
}

func (m *DataLatencyThreshold) Reset()                    { *m = DataLatencyThreshold{} }
func (m *DataLatencyThreshold) String() string            { return proto.CompactTextString(m) }
func (*DataLatencyThreshold) ProtoMessage()               {}
func (*DataLatencyThreshold) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type DataLatencyThresholdResult struct {
	Result []*DataLatencyThreshold `protobuf:"bytes,1,rep,name=result" json:"result,omitempty"`
}

func (m *DataLatencyThresholdResult) Reset()                    { *m = DataLatencyThresholdResult{} }
func (m *DataLatencyThresholdResult) String() string            { return proto.CompactTextString(m) }
func (*DataLatencyThresholdResult) ProtoMessage()               {}
func (*DataLatencyThresholdResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *DataLatencyThresholdResult) GetResult() []*DataLatencyThreshold {
	if m != nil {
		return m.Result
	}
	return nil
}

func init() {
	proto.RegisterType((*DataLatencySummary)(nil), "mtrpb.DataLatencySummary")
	proto.RegisterType((*DataLatencySummaryResult)(nil), "mtrpb.DataLatencySummaryResult")
	proto.RegisterType((*DataSite)(nil), "mtrpb.DataSite")
	proto.RegisterType((*DataSiteResult)(nil), "mtrpb.DataSiteResult")
	proto.RegisterType((*DataLatencyTag)(nil), "mtrpb.DataLatencyTag")
	proto.RegisterType((*DataLatencyTagResult)(nil), "mtrpb.DataLatencyTagResult")
	proto.RegisterType((*DataLatencyThreshold)(nil), "mtrpb.DataLatencyThreshold")
	proto.RegisterType((*DataLatencyThresholdResult)(nil), "mtrpb.DataLatencyThresholdResult")
}

var fileDescriptor0 = []byte{
	// 366 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x53, 0x3d, 0x4f, 0xc3, 0x30,
	0x10, 0x55, 0x48, 0x93, 0xb6, 0x87, 0x04, 0xc8, 0x2a, 0x60, 0x0a, 0x03, 0xca, 0x42, 0x17, 0x2a,
	0x41, 0x27, 0x56, 0x54, 0x06, 0x24, 0x18, 0x48, 0x3b, 0x21, 0x21, 0xe4, 0x36, 0x6e, 0x1b, 0x29,
	0x5f, 0x4a, 0x1d, 0xa1, 0xfc, 0x49, 0x7e, 0x13, 0xf6, 0xd9, 0xe9, 0x67, 0x58, 0xba, 0xf9, 0xbd,
	0x77, 0xe7, 0xbc, 0x77, 0x17, 0x03, 0x04, 0x4c, 0xb0, 0x7e, 0x96, 0xa7, 0x22, 0x25, 0x4e, 0x2c,
	0xf2, 0x6c, 0xe2, 0xfd, 0x5a, 0x40, 0x86, 0x92, 0x7d, 0x63, 0x82, 0x27, 0xd3, 0x72, 0x54, 0xc4,
	0x31, 0xcb, 0x4b, 0x72, 0x09, 0xcd, 0x65, 0x28, 0xf8, 0x77, 0x38, 0xa4, 0xd6, 0xad, 0xd5, 0x6b,
	0xfb, 0xae, 0x82, 0xaf, 0x43, 0x25, 0x88, 0x32, 0x43, 0xe1, 0x48, 0x0b, 0x0a, 0x4a, 0x81, 0xca,
	0x0e, 0x3e, 0x4d, 0x93, 0x60, 0x49, 0x6d, 0x29, 0xd8, 0x7e, 0x05, 0x09, 0x81, 0x46, 0xcc, 0x59,
	0x42, 0x1b, 0x92, 0x76, 0x7c, 0x3c, 0x93, 0x0e, 0x38, 0xb3, 0x70, 0x26, 0x4a, 0xea, 0x20, 0xa9,
	0x01, 0xb9, 0x00, 0x37, 0x09, 0x13, 0x2e, 0x69, 0x17, 0x69, 0x83, 0x54, 0x75, 0x91, 0x65, 0x3c,
	0xa7, 0x4d, 0x5d, 0x8d, 0x40, 0xb1, 0x51, 0xfa, 0x23, 0xd9, 0x96, 0x66, 0x11, 0x78, 0xef, 0x40,
	0xf7, 0xf3, 0xf8, 0x7c, 0x59, 0x44, 0x82, 0x3c, 0x80, 0x9b, 0xe3, 0x49, 0x86, 0xb2, 0x7b, 0xc7,
	0x8f, 0x57, 0x7d, 0x1c, 0x42, 0xbf, 0xa6, 0xc1, 0x14, 0x7a, 0x5f, 0xd0, 0x52, 0xea, 0x48, 0xa6,
	0xff, 0x7f, 0x28, 0x5d, 0x68, 0x45, 0x4c, 0x84, 0xa2, 0x08, 0x38, 0x4e, 0xc5, 0xf2, 0x57, 0x98,
	0xdc, 0x40, 0x3b, 0x4a, 0x93, 0xb9, 0x16, 0x6d, 0x14, 0xd7, 0x84, 0xf7, 0x04, 0x27, 0xd5, 0xf5,
	0xc6, 0xe3, 0xdd, 0x8e, 0xc7, 0xd3, 0x0d, 0x8f, 0x58, 0x56, 0x39, 0x1b, 0xeb, 0x56, 0xe3, 0x7b,
	0xcc, 0xe6, 0x07, 0x2c, 0xed, 0x0c, 0x6c, 0xc1, 0xe6, 0x68, 0xab, 0xed, 0xab, 0xa3, 0xf7, 0x02,
	0x9d, 0xed, 0x5b, 0x8d, 0xad, 0xfb, 0x1d, 0x5b, 0xe7, 0xfb, 0xa3, 0x53, 0xc5, 0x95, 0x39, 0xb1,
	0x7d, 0xcd, 0x42, 0xd2, 0x8b, 0x34, 0x0a, 0x0e, 0xb0, 0xb8, 0xda, 0xb2, 0xbd, 0xb1, 0xe5, 0xf5,
	0x1f, 0xd1, 0xd8, 0xf8, 0x23, 0xbc, 0x0f, 0xe8, 0xd6, 0x7d, 0xd5, 0x44, 0x18, 0xec, 0x44, 0xb8,
	0xae, 0x89, 0xb0, 0x6a, 0x31, 0xa5, 0xcf, 0xcd, 0x4f, 0xfd, 0x50, 0x26, 0x2e, 0x3e, 0x9b, 0xc1,
	0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xe0, 0x1f, 0xd4, 0xda, 0x44, 0x03, 0x00, 0x00,
}
