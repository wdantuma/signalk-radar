// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.12
// source: RadarMessage.proto

package radar

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

type RadarMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Radar  uint32                `protobuf:"varint,1,opt,name=radar,proto3" json:"radar,omitempty"`
	Spokes []*RadarMessage_Spoke `protobuf:"bytes,2,rep,name=spokes,proto3" json:"spokes,omitempty"`
}

func (x *RadarMessage) Reset() {
	*x = RadarMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_RadarMessage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RadarMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RadarMessage) ProtoMessage() {}

func (x *RadarMessage) ProtoReflect() protoreflect.Message {
	mi := &file_RadarMessage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RadarMessage.ProtoReflect.Descriptor instead.
func (*RadarMessage) Descriptor() ([]byte, []int) {
	return file_RadarMessage_proto_rawDescGZIP(), []int{0}
}

func (x *RadarMessage) GetRadar() uint32 {
	if x != nil {
		return x.Radar
	}
	return 0
}

func (x *RadarMessage) GetSpokes() []*RadarMessage_Spoke {
	if x != nil {
		return x.Spokes
	}
	return nil
}

type RadarMessage_Spoke struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Angle   uint32 `protobuf:"varint,1,opt,name=angle,proto3" json:"angle,omitempty"`
	Bearing uint32 `protobuf:"varint,2,opt,name=bearing,proto3" json:"bearing,omitempty"`
	Range   uint32 `protobuf:"varint,3,opt,name=range,proto3" json:"range,omitempty"`
	Time    uint64 `protobuf:"varint,4,opt,name=time,proto3" json:"time,omitempty"`
	Data    []byte `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *RadarMessage_Spoke) Reset() {
	*x = RadarMessage_Spoke{}
	if protoimpl.UnsafeEnabled {
		mi := &file_RadarMessage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RadarMessage_Spoke) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RadarMessage_Spoke) ProtoMessage() {}

func (x *RadarMessage_Spoke) ProtoReflect() protoreflect.Message {
	mi := &file_RadarMessage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RadarMessage_Spoke.ProtoReflect.Descriptor instead.
func (*RadarMessage_Spoke) Descriptor() ([]byte, []int) {
	return file_RadarMessage_proto_rawDescGZIP(), []int{0, 0}
}

func (x *RadarMessage_Spoke) GetAngle() uint32 {
	if x != nil {
		return x.Angle
	}
	return 0
}

func (x *RadarMessage_Spoke) GetBearing() uint32 {
	if x != nil {
		return x.Bearing
	}
	return 0
}

func (x *RadarMessage_Spoke) GetRange() uint32 {
	if x != nil {
		return x.Range
	}
	return 0
}

func (x *RadarMessage_Spoke) GetTime() uint64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *RadarMessage_Spoke) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_RadarMessage_proto protoreflect.FileDescriptor

var file_RadarMessage_proto_rawDesc = []byte{
	0x0a, 0x12, 0x52, 0x61, 0x64, 0x61, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc8, 0x01, 0x0a, 0x0c, 0x52, 0x61, 0x64, 0x61, 0x72, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x61, 0x64, 0x61, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x72, 0x61, 0x64, 0x61, 0x72, 0x12, 0x2b, 0x0a, 0x06, 0x73,
	0x70, 0x6f, 0x6b, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x52, 0x61,
	0x64, 0x61, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x53, 0x70, 0x6f, 0x6b, 0x65,
	0x52, 0x06, 0x73, 0x70, 0x6f, 0x6b, 0x65, 0x73, 0x1a, 0x75, 0x0a, 0x05, 0x53, 0x70, 0x6f, 0x6b,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x6e, 0x67, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x05, 0x61, 0x6e, 0x67, 0x6c, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x65, 0x61, 0x72, 0x69,
	0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x62, 0x65, 0x61, 0x72, 0x69, 0x6e,
	0x67, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42,
	0x0a, 0x5a, 0x08, 0x2e, 0x2e, 0x2f, 0x72, 0x61, 0x64, 0x61, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_RadarMessage_proto_rawDescOnce sync.Once
	file_RadarMessage_proto_rawDescData = file_RadarMessage_proto_rawDesc
)

func file_RadarMessage_proto_rawDescGZIP() []byte {
	file_RadarMessage_proto_rawDescOnce.Do(func() {
		file_RadarMessage_proto_rawDescData = protoimpl.X.CompressGZIP(file_RadarMessage_proto_rawDescData)
	})
	return file_RadarMessage_proto_rawDescData
}

var file_RadarMessage_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_RadarMessage_proto_goTypes = []interface{}{
	(*RadarMessage)(nil),       // 0: RadarMessage
	(*RadarMessage_Spoke)(nil), // 1: RadarMessage.Spoke
}
var file_RadarMessage_proto_depIdxs = []int32{
	1, // 0: RadarMessage.spokes:type_name -> RadarMessage.Spoke
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_RadarMessage_proto_init() }
func file_RadarMessage_proto_init() {
	if File_RadarMessage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_RadarMessage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RadarMessage); i {
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
		file_RadarMessage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RadarMessage_Spoke); i {
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
			RawDescriptor: file_RadarMessage_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_RadarMessage_proto_goTypes,
		DependencyIndexes: file_RadarMessage_proto_depIdxs,
		MessageInfos:      file_RadarMessage_proto_msgTypes,
	}.Build()
	File_RadarMessage_proto = out.File
	file_RadarMessage_proto_rawDesc = nil
	file_RadarMessage_proto_goTypes = nil
	file_RadarMessage_proto_depIdxs = nil
}
