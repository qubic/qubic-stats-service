// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.27.2
// source: stats-api.proto

package protobuff

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type QubicData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp                int64   `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	CirculatingSupply        int64   `protobuf:"varint,2,opt,name=circulating_supply,json=circulatingSupply,proto3" json:"circulating_supply,omitempty"`
	ActiveAddresses          int32   `protobuf:"varint,3,opt,name=active_addresses,json=activeAddresses,proto3" json:"active_addresses,omitempty"`
	Price                    float32 `protobuf:"fixed32,4,opt,name=price,proto3" json:"price,omitempty"`
	MarketCap                int64   `protobuf:"varint,5,opt,name=market_cap,json=marketCap,proto3" json:"market_cap,omitempty"`
	Epoch                    uint32  `protobuf:"varint,6,opt,name=epoch,proto3" json:"epoch,omitempty"`
	CurrentTick              uint32  `protobuf:"varint,7,opt,name=current_tick,json=currentTick,proto3" json:"current_tick,omitempty"`
	TicksInCurrentEpoch      uint32  `protobuf:"varint,8,opt,name=ticks_in_current_epoch,json=ticksInCurrentEpoch,proto3" json:"ticks_in_current_epoch,omitempty"`
	EmptyTicksInCurrentEpoch uint32  `protobuf:"varint,9,opt,name=empty_ticks_in_current_epoch,json=emptyTicksInCurrentEpoch,proto3" json:"empty_ticks_in_current_epoch,omitempty"`
	EpochTickQuality         float32 `protobuf:"fixed32,10,opt,name=epoch_tick_quality,json=epochTickQuality,proto3" json:"epoch_tick_quality,omitempty"`
	BurnedQus                uint64  `protobuf:"varint,11,opt,name=burned_qus,json=burnedQus,proto3" json:"burned_qus,omitempty"`
}

func (x *QubicData) Reset() {
	*x = QubicData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QubicData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QubicData) ProtoMessage() {}

func (x *QubicData) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QubicData.ProtoReflect.Descriptor instead.
func (*QubicData) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{0}
}

func (x *QubicData) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *QubicData) GetCirculatingSupply() int64 {
	if x != nil {
		return x.CirculatingSupply
	}
	return 0
}

func (x *QubicData) GetActiveAddresses() int32 {
	if x != nil {
		return x.ActiveAddresses
	}
	return 0
}

func (x *QubicData) GetPrice() float32 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *QubicData) GetMarketCap() int64 {
	if x != nil {
		return x.MarketCap
	}
	return 0
}

func (x *QubicData) GetEpoch() uint32 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *QubicData) GetCurrentTick() uint32 {
	if x != nil {
		return x.CurrentTick
	}
	return 0
}

func (x *QubicData) GetTicksInCurrentEpoch() uint32 {
	if x != nil {
		return x.TicksInCurrentEpoch
	}
	return 0
}

func (x *QubicData) GetEmptyTicksInCurrentEpoch() uint32 {
	if x != nil {
		return x.EmptyTicksInCurrentEpoch
	}
	return 0
}

func (x *QubicData) GetEpochTickQuality() float32 {
	if x != nil {
		return x.EpochTickQuality
	}
	return 0
}

func (x *QubicData) GetBurnedQus() uint64 {
	if x != nil {
		return x.BurnedQus
	}
	return 0
}

type GetLatestDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data *QubicData `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *GetLatestDataResponse) Reset() {
	*x = GetLatestDataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetLatestDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLatestDataResponse) ProtoMessage() {}

func (x *GetLatestDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetLatestDataResponse.ProtoReflect.Descriptor instead.
func (*GetLatestDataResponse) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{1}
}

func (x *GetLatestDataResponse) GetData() *QubicData {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_stats_api_proto protoreflect.FileDescriptor

var file_stats_api_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x12, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x70, 0x62, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xb3, 0x03, 0x0a, 0x09, 0x51, 0x75, 0x62, 0x69, 0x63, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1c,
	0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x2d, 0x0a, 0x12,
	0x63, 0x69, 0x72, 0x63, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x75, 0x70, 0x70,
	0x6c, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x63, 0x69, 0x72, 0x63, 0x75, 0x6c,
	0x61, 0x74, 0x69, 0x6e, 0x67, 0x53, 0x75, 0x70, 0x70, 0x6c, 0x79, 0x12, 0x29, 0x0a, 0x10, 0x61,
	0x63, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x12, 0x1d, 0x0a, 0x0a,
	0x6d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x5f, 0x63, 0x61, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x43, 0x61, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x70, 0x6f, 0x63, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63,
	0x68, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x69, 0x63,
	0x6b, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
	0x54, 0x69, 0x63, 0x6b, 0x12, 0x33, 0x0a, 0x16, 0x74, 0x69, 0x63, 0x6b, 0x73, 0x5f, 0x69, 0x6e,
	0x5f, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x13, 0x74, 0x69, 0x63, 0x6b, 0x73, 0x49, 0x6e, 0x43, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x74, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x3e, 0x0a, 0x1c, 0x65, 0x6d, 0x70,
	0x74, 0x79, 0x5f, 0x74, 0x69, 0x63, 0x6b, 0x73, 0x5f, 0x69, 0x6e, 0x5f, 0x63, 0x75, 0x72, 0x72,
	0x65, 0x6e, 0x74, 0x5f, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x18, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x54, 0x69, 0x63, 0x6b, 0x73, 0x49, 0x6e, 0x43, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x74, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x2c, 0x0a, 0x12, 0x65, 0x70, 0x6f,
	0x63, 0x68, 0x5f, 0x74, 0x69, 0x63, 0x6b, 0x5f, 0x71, 0x75, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x02, 0x52, 0x10, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x54, 0x69, 0x63, 0x6b,
	0x51, 0x75, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x75, 0x72, 0x6e, 0x65,
	0x64, 0x5f, 0x71, 0x75, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x62, 0x75, 0x72,
	0x6e, 0x65, 0x64, 0x51, 0x75, 0x73, 0x22, 0x4a, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74,
	0x65, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x31, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e,
	0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x70, 0x62, 0x2e, 0x51, 0x75, 0x62, 0x69, 0x63, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x32, 0x7c, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x6c, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x29, 0x2e, 0x71, 0x75,
	0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62,
	0x2e, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10,
	0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2d, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74,
	0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x71,
	0x75, 0x62, 0x69, 0x63, 0x2f, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2d, 0x73, 0x74, 0x61, 0x74, 0x73,
	0x2d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x66, 0x2f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_stats_api_proto_rawDescOnce sync.Once
	file_stats_api_proto_rawDescData = file_stats_api_proto_rawDesc
)

func file_stats_api_proto_rawDescGZIP() []byte {
	file_stats_api_proto_rawDescOnce.Do(func() {
		file_stats_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_stats_api_proto_rawDescData)
	})
	return file_stats_api_proto_rawDescData
}

var file_stats_api_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_stats_api_proto_goTypes = []interface{}{
	(*QubicData)(nil),             // 0: qubic.stats.api.pb.QubicData
	(*GetLatestDataResponse)(nil), // 1: qubic.stats.api.pb.GetLatestDataResponse
	(*emptypb.Empty)(nil),         // 2: google.protobuf.Empty
}
var file_stats_api_proto_depIdxs = []int32{
	0, // 0: qubic.stats.api.pb.GetLatestDataResponse.data:type_name -> qubic.stats.api.pb.QubicData
	2, // 1: qubic.stats.api.pb.StatsService.GetLatestData:input_type -> google.protobuf.Empty
	1, // 2: qubic.stats.api.pb.StatsService.GetLatestData:output_type -> qubic.stats.api.pb.GetLatestDataResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_stats_api_proto_init() }
func file_stats_api_proto_init() {
	if File_stats_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_stats_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QubicData); i {
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
		file_stats_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetLatestDataResponse); i {
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
			RawDescriptor: file_stats_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_stats_api_proto_goTypes,
		DependencyIndexes: file_stats_api_proto_depIdxs,
		MessageInfos:      file_stats_api_proto_msgTypes,
	}.Build()
	File_stats_api_proto = out.File
	file_stats_api_proto_rawDesc = nil
	file_stats_api_proto_goTypes = nil
	file_stats_api_proto_depIdxs = nil
}
