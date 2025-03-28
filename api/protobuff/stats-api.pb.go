// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.21.12
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

type Pagination struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TotalRecords int32 `protobuf:"varint,1,opt,name=total_records,json=totalRecords,proto3" json:"total_records,omitempty"`
	CurrentPage  int32 `protobuf:"varint,2,opt,name=current_page,json=currentPage,proto3" json:"current_page,omitempty"`
	TotalPages   int32 `protobuf:"varint,3,opt,name=total_pages,json=totalPages,proto3" json:"total_pages,omitempty"`
	PageSize     int32 `protobuf:"varint,4,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

func (x *Pagination) Reset() {
	*x = Pagination{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Pagination) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pagination) ProtoMessage() {}

func (x *Pagination) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Pagination.ProtoReflect.Descriptor instead.
func (*Pagination) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{1}
}

func (x *Pagination) GetTotalRecords() int32 {
	if x != nil {
		return x.TotalRecords
	}
	return 0
}

func (x *Pagination) GetCurrentPage() int32 {
	if x != nil {
		return x.CurrentPage
	}
	return 0
}

func (x *Pagination) GetTotalPages() int32 {
	if x != nil {
		return x.TotalPages
	}
	return 0
}

func (x *Pagination) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type RichListEntity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Identity string `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity,omitempty"`
	Balance  int64  `protobuf:"varint,2,opt,name=balance,proto3" json:"balance,omitempty"`
}

func (x *RichListEntity) Reset() {
	*x = RichListEntity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RichListEntity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RichListEntity) ProtoMessage() {}

func (x *RichListEntity) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RichListEntity.ProtoReflect.Descriptor instead.
func (*RichListEntity) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{2}
}

func (x *RichListEntity) GetIdentity() string {
	if x != nil {
		return x.Identity
	}
	return ""
}

func (x *RichListEntity) GetBalance() int64 {
	if x != nil {
		return x.Balance
	}
	return 0
}

type RichList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entities []*RichListEntity `protobuf:"bytes,1,rep,name=entities,proto3" json:"entities,omitempty"`
}

func (x *RichList) Reset() {
	*x = RichList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RichList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RichList) ProtoMessage() {}

func (x *RichList) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RichList.ProtoReflect.Descriptor instead.
func (*RichList) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{3}
}

func (x *RichList) GetEntities() []*RichListEntity {
	if x != nil {
		return x.Entities
	}
	return nil
}

type GetRichListSliceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page     int32 `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

func (x *GetRichListSliceRequest) Reset() {
	*x = GetRichListSliceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRichListSliceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRichListSliceRequest) ProtoMessage() {}

func (x *GetRichListSliceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRichListSliceRequest.ProtoReflect.Descriptor instead.
func (*GetRichListSliceRequest) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{4}
}

func (x *GetRichListSliceRequest) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *GetRichListSliceRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type GetRichListSliceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pagination *Pagination `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
	Epoch      uint32      `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	RichList   *RichList   `protobuf:"bytes,3,opt,name=rich_list,json=richList,proto3" json:"rich_list,omitempty"`
}

func (x *GetRichListSliceResponse) Reset() {
	*x = GetRichListSliceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRichListSliceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRichListSliceResponse) ProtoMessage() {}

func (x *GetRichListSliceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRichListSliceResponse.ProtoReflect.Descriptor instead.
func (*GetRichListSliceResponse) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{5}
}

func (x *GetRichListSliceResponse) GetPagination() *Pagination {
	if x != nil {
		return x.Pagination
	}
	return nil
}

func (x *GetRichListSliceResponse) GetEpoch() uint32 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *GetRichListSliceResponse) GetRichList() *RichList {
	if x != nil {
		return x.RichList
	}
	return nil
}

type GetAssetOwnershipRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IssuerIdentity string `protobuf:"bytes,1,opt,name=issuer_identity,json=issuerIdentity,proto3" json:"issuer_identity,omitempty"`
	AssetName      string `protobuf:"bytes,2,opt,name=asset_name,json=assetName,proto3" json:"asset_name,omitempty"`
	Page           uint32 `protobuf:"varint,3,opt,name=page,proto3" json:"page,omitempty"`
	PageSize       uint32 `protobuf:"varint,4,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

func (x *GetAssetOwnershipRequest) Reset() {
	*x = GetAssetOwnershipRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAssetOwnershipRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAssetOwnershipRequest) ProtoMessage() {}

func (x *GetAssetOwnershipRequest) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAssetOwnershipRequest.ProtoReflect.Descriptor instead.
func (*GetAssetOwnershipRequest) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{6}
}

func (x *GetAssetOwnershipRequest) GetIssuerIdentity() string {
	if x != nil {
		return x.IssuerIdentity
	}
	return ""
}

func (x *GetAssetOwnershipRequest) GetAssetName() string {
	if x != nil {
		return x.AssetName
	}
	return ""
}

func (x *GetAssetOwnershipRequest) GetPage() uint32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *GetAssetOwnershipRequest) GetPageSize() uint32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type GetAssetOwnershipResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pagination *Pagination       `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
	Tick       uint32            `protobuf:"varint,2,opt,name=tick,proto3" json:"tick,omitempty"`
	Owners     []*AssetOwnership `protobuf:"bytes,3,rep,name=owners,proto3" json:"owners,omitempty"`
}

func (x *GetAssetOwnershipResponse) Reset() {
	*x = GetAssetOwnershipResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAssetOwnershipResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAssetOwnershipResponse) ProtoMessage() {}

func (x *GetAssetOwnershipResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAssetOwnershipResponse.ProtoReflect.Descriptor instead.
func (*GetAssetOwnershipResponse) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{7}
}

func (x *GetAssetOwnershipResponse) GetPagination() *Pagination {
	if x != nil {
		return x.Pagination
	}
	return nil
}

func (x *GetAssetOwnershipResponse) GetTick() uint32 {
	if x != nil {
		return x.Tick
	}
	return 0
}

func (x *GetAssetOwnershipResponse) GetOwners() []*AssetOwnership {
	if x != nil {
		return x.Owners
	}
	return nil
}

type AssetOwnership struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Identity       string `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity,omitempty"`
	NumberOfShares int64  `protobuf:"varint,2,opt,name=number_of_shares,json=numberOfShares,proto3" json:"number_of_shares,omitempty"`
}

func (x *AssetOwnership) Reset() {
	*x = AssetOwnership{}
	if protoimpl.UnsafeEnabled {
		mi := &file_stats_api_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssetOwnership) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssetOwnership) ProtoMessage() {}

func (x *AssetOwnership) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssetOwnership.ProtoReflect.Descriptor instead.
func (*AssetOwnership) Descriptor() ([]byte, []int) {
	return file_stats_api_proto_rawDescGZIP(), []int{8}
}

func (x *AssetOwnership) GetIdentity() string {
	if x != nil {
		return x.Identity
	}
	return ""
}

func (x *AssetOwnership) GetNumberOfShares() int64 {
	if x != nil {
		return x.NumberOfShares
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
		mi := &file_stats_api_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetLatestDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLatestDataResponse) ProtoMessage() {}

func (x *GetLatestDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_stats_api_proto_msgTypes[9]
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
	return file_stats_api_proto_rawDescGZIP(), []int{9}
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
	0x6e, 0x65, 0x64, 0x51, 0x75, 0x73, 0x22, 0x92, 0x01, 0x0a, 0x0a, 0x50, 0x61, 0x67, 0x69, 0x6e,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x72,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x74, 0x6f,
	0x74, 0x61, 0x6c, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x75,
	0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0b, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x67, 0x65, 0x12, 0x1f, 0x0a,
	0x0b, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0a, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x50, 0x61, 0x67, 0x65, 0x73, 0x12, 0x1b,
	0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x46, 0x0a, 0x0e, 0x52,
	0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x1a, 0x0a,
	0x08, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x62, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x22, 0x4a, 0x0a, 0x08, 0x52, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x12,
	0x3e, 0x0a, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x22, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x45,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x08, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x22,
	0x4a, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x52, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x6c,
	0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1b,
	0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0xab, 0x01, 0x0a, 0x18,
	0x47, 0x65, 0x74, 0x52, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x6c, 0x69, 0x63, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3e, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x69,
	0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x71,
	0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70,
	0x62, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x70, 0x61,
	0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x39,
	0x0a, 0x09, 0x72, 0x69, 0x63, 0x68, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1c, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x52,
	0x08, 0x72, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x22, 0x93, 0x01, 0x0a, 0x18, 0x47, 0x65,
	0x74, 0x41, 0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12,
	0x1d, 0x0a, 0x0a, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x73, 0x73, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x61,
	0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22,
	0xab, 0x01, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x41, 0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65,
	0x72, 0x73, 0x68, 0x69, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3e, 0x0a,
	0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1e, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x69, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x69, 0x63,
	0x6b, 0x12, 0x3a, 0x0a, 0x06, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x22, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65,
	0x72, 0x73, 0x68, 0x69, 0x70, 0x52, 0x06, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x22, 0x56, 0x0a,
	0x0e, 0x41, 0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x12,
	0x1a, 0x0a, 0x08, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x28, 0x0a, 0x10, 0x6e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x6f, 0x66, 0x5f, 0x73, 0x68, 0x61, 0x72, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0e, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x66, 0x53,
	0x68, 0x61, 0x72, 0x65, 0x73, 0x22, 0x4a, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65,
	0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x71,
	0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70,
	0x62, 0x2e, 0x51, 0x75, 0x62, 0x69, 0x63, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x32, 0xb5, 0x03, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x6c, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x29, 0x2e, 0x71, 0x75,
	0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62,
	0x2e, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10,
	0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x2d, 0x73, 0x74, 0x61, 0x74, 0x73,
	0x12, 0x84, 0x01, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x52, 0x69, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74,
	0x53, 0x6c, 0x69, 0x63, 0x65, 0x12, 0x2b, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74,
	0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x69,
	0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x6c, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x69, 0x63, 0x68, 0x4c,
	0x69, 0x73, 0x74, 0x53, 0x6c, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x15, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0f, 0x12, 0x0d, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x69,
	0x63, 0x68, 0x2d, 0x6c, 0x69, 0x73, 0x74, 0x12, 0xaf, 0x01, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x41,
	0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x12, 0x2c, 0x2e, 0x71, 0x75, 0x62,
	0x69, 0x63, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e,
	0x47, 0x65, 0x74, 0x41, 0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x68, 0x69,
	0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x71, 0x75, 0x62, 0x69, 0x63,
	0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x65,
	0x74, 0x41, 0x73, 0x73, 0x65, 0x74, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x40, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x3a, 0x12,
	0x38, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72, 0x73, 0x2f, 0x7b, 0x69, 0x73,
	0x73, 0x75, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x7d, 0x2f, 0x61,
	0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x7b, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x7d, 0x2f, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x71, 0x75, 0x62, 0x69, 0x63, 0x2f, 0x71, 0x75,
	0x62, 0x69, 0x63, 0x2d, 0x73, 0x74, 0x61, 0x74, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x66, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_stats_api_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_stats_api_proto_goTypes = []any{
	(*QubicData)(nil),                 // 0: qubic.stats.api.pb.QubicData
	(*Pagination)(nil),                // 1: qubic.stats.api.pb.Pagination
	(*RichListEntity)(nil),            // 2: qubic.stats.api.pb.RichListEntity
	(*RichList)(nil),                  // 3: qubic.stats.api.pb.RichList
	(*GetRichListSliceRequest)(nil),   // 4: qubic.stats.api.pb.GetRichListSliceRequest
	(*GetRichListSliceResponse)(nil),  // 5: qubic.stats.api.pb.GetRichListSliceResponse
	(*GetAssetOwnershipRequest)(nil),  // 6: qubic.stats.api.pb.GetAssetOwnershipRequest
	(*GetAssetOwnershipResponse)(nil), // 7: qubic.stats.api.pb.GetAssetOwnershipResponse
	(*AssetOwnership)(nil),            // 8: qubic.stats.api.pb.AssetOwnership
	(*GetLatestDataResponse)(nil),     // 9: qubic.stats.api.pb.GetLatestDataResponse
	(*emptypb.Empty)(nil),             // 10: google.protobuf.Empty
}
var file_stats_api_proto_depIdxs = []int32{
	2,  // 0: qubic.stats.api.pb.RichList.entities:type_name -> qubic.stats.api.pb.RichListEntity
	1,  // 1: qubic.stats.api.pb.GetRichListSliceResponse.pagination:type_name -> qubic.stats.api.pb.Pagination
	3,  // 2: qubic.stats.api.pb.GetRichListSliceResponse.rich_list:type_name -> qubic.stats.api.pb.RichList
	1,  // 3: qubic.stats.api.pb.GetAssetOwnershipResponse.pagination:type_name -> qubic.stats.api.pb.Pagination
	8,  // 4: qubic.stats.api.pb.GetAssetOwnershipResponse.owners:type_name -> qubic.stats.api.pb.AssetOwnership
	0,  // 5: qubic.stats.api.pb.GetLatestDataResponse.data:type_name -> qubic.stats.api.pb.QubicData
	10, // 6: qubic.stats.api.pb.StatsService.GetLatestData:input_type -> google.protobuf.Empty
	4,  // 7: qubic.stats.api.pb.StatsService.GetRichListSlice:input_type -> qubic.stats.api.pb.GetRichListSliceRequest
	6,  // 8: qubic.stats.api.pb.StatsService.GetAssetOwners:input_type -> qubic.stats.api.pb.GetAssetOwnershipRequest
	9,  // 9: qubic.stats.api.pb.StatsService.GetLatestData:output_type -> qubic.stats.api.pb.GetLatestDataResponse
	5,  // 10: qubic.stats.api.pb.StatsService.GetRichListSlice:output_type -> qubic.stats.api.pb.GetRichListSliceResponse
	7,  // 11: qubic.stats.api.pb.StatsService.GetAssetOwners:output_type -> qubic.stats.api.pb.GetAssetOwnershipResponse
	9,  // [9:12] is the sub-list for method output_type
	6,  // [6:9] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_stats_api_proto_init() }
func file_stats_api_proto_init() {
	if File_stats_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_stats_api_proto_msgTypes[0].Exporter = func(v any, i int) any {
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
		file_stats_api_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Pagination); i {
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
		file_stats_api_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*RichListEntity); i {
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
		file_stats_api_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*RichList); i {
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
		file_stats_api_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*GetRichListSliceRequest); i {
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
		file_stats_api_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*GetRichListSliceResponse); i {
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
		file_stats_api_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*GetAssetOwnershipRequest); i {
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
		file_stats_api_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*GetAssetOwnershipResponse); i {
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
		file_stats_api_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*AssetOwnership); i {
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
		file_stats_api_proto_msgTypes[9].Exporter = func(v any, i int) any {
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
			NumMessages:   10,
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
