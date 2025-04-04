// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.29.2
// source: crypto.proto

package v1

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

type CoinType int32

const (
	CoinType_XMR CoinType = 0
	CoinType_BTC CoinType = 1
	CoinType_LTC CoinType = 2
	CoinType_ETH CoinType = 3
	CoinType_TON CoinType = 4
	// ERC20
	CoinType_USDT_ERC20  CoinType = 5
	CoinType_USDC_ERC20  CoinType = 6
	CoinType_DAI_ERC20   CoinType = 7
	CoinType_WBTC_ERC20  CoinType = 8
	CoinType_UNI_ERC20   CoinType = 9
	CoinType_LINK_ERC20  CoinType = 10
	CoinType_AAVE_ERC20  CoinType = 11
	CoinType_CRV_ERC20   CoinType = 12
	CoinType_MATIC_ERC20 CoinType = 13
	CoinType_SHIB_ERC20  CoinType = 14
	CoinType_BNB_ERC20   CoinType = 15
	CoinType_ATOM_ERC20  CoinType = 16
	CoinType_ARB_ERC20   CoinType = 17
	CoinType_BNB         CoinType = 18
	// BEP20
	CoinType_BSCUSD_BEP20 CoinType = 19
	CoinType_USDC_BEP20   CoinType = 20
	CoinType_DAI_BEP20    CoinType = 21
	CoinType_BUSD_BEP20   CoinType = 22
	CoinType_WBTC_BEP20   CoinType = 23
	CoinType_BTCB_BEP20   CoinType = 24
	CoinType_UNI_BEP20    CoinType = 25
	CoinType_LINK_BEP20   CoinType = 26
	CoinType_AAVE_BEP20   CoinType = 27
	CoinType_MATIC_BEP20  CoinType = 28
	CoinType_SHIB_BEP20   CoinType = 29
	CoinType_ATOM_BEP20   CoinType = 30
	CoinType_ARB_BEP20    CoinType = 31
	CoinType_ETH_BEP20    CoinType = 32
	CoinType_XRP_BEP20    CoinType = 33
	CoinType_ADA_BEP20    CoinType = 34
	CoinType_TRX_BEP20    CoinType = 35
	CoinType_DOGE_BEP20   CoinType = 36
	CoinType_LTC_BEP20    CoinType = 37
	CoinType_BCH_BEP20    CoinType = 38
	CoinType_TWT_BEP20    CoinType = 39
	CoinType_AVAX_BEP20   CoinType = 40
	CoinType_CAKE_BEP20   CoinType = 41
)

// Enum value maps for CoinType.
var (
	CoinType_name = map[int32]string{
		0:  "XMR",
		1:  "BTC",
		2:  "LTC",
		3:  "ETH",
		4:  "TON",
		5:  "USDT_ERC20",
		6:  "USDC_ERC20",
		7:  "DAI_ERC20",
		8:  "WBTC_ERC20",
		9:  "UNI_ERC20",
		10: "LINK_ERC20",
		11: "AAVE_ERC20",
		12: "CRV_ERC20",
		13: "MATIC_ERC20",
		14: "SHIB_ERC20",
		15: "BNB_ERC20",
		16: "ATOM_ERC20",
		17: "ARB_ERC20",
		18: "BNB",
		19: "BSCUSD_BEP20",
		20: "USDC_BEP20",
		21: "DAI_BEP20",
		22: "BUSD_BEP20",
		23: "WBTC_BEP20",
		24: "BTCB_BEP20",
		25: "UNI_BEP20",
		26: "LINK_BEP20",
		27: "AAVE_BEP20",
		28: "MATIC_BEP20",
		29: "SHIB_BEP20",
		30: "ATOM_BEP20",
		31: "ARB_BEP20",
		32: "ETH_BEP20",
		33: "XRP_BEP20",
		34: "ADA_BEP20",
		35: "TRX_BEP20",
		36: "DOGE_BEP20",
		37: "LTC_BEP20",
		38: "BCH_BEP20",
		39: "TWT_BEP20",
		40: "AVAX_BEP20",
		41: "CAKE_BEP20",
	}
	CoinType_value = map[string]int32{
		"XMR":          0,
		"BTC":          1,
		"LTC":          2,
		"ETH":          3,
		"TON":          4,
		"USDT_ERC20":   5,
		"USDC_ERC20":   6,
		"DAI_ERC20":    7,
		"WBTC_ERC20":   8,
		"UNI_ERC20":    9,
		"LINK_ERC20":   10,
		"AAVE_ERC20":   11,
		"CRV_ERC20":    12,
		"MATIC_ERC20":  13,
		"SHIB_ERC20":   14,
		"BNB_ERC20":    15,
		"ATOM_ERC20":   16,
		"ARB_ERC20":    17,
		"BNB":          18,
		"BSCUSD_BEP20": 19,
		"USDC_BEP20":   20,
		"DAI_BEP20":    21,
		"BUSD_BEP20":   22,
		"WBTC_BEP20":   23,
		"BTCB_BEP20":   24,
		"UNI_BEP20":    25,
		"LINK_BEP20":   26,
		"AAVE_BEP20":   27,
		"MATIC_BEP20":  28,
		"SHIB_BEP20":   29,
		"ATOM_BEP20":   30,
		"ARB_BEP20":    31,
		"ETH_BEP20":    32,
		"XRP_BEP20":    33,
		"ADA_BEP20":    34,
		"TRX_BEP20":    35,
		"DOGE_BEP20":   36,
		"LTC_BEP20":    37,
		"BCH_BEP20":    38,
		"TWT_BEP20":    39,
		"AVAX_BEP20":   40,
		"CAKE_BEP20":   41,
	}
)

func (x CoinType) Enum() *CoinType {
	p := new(CoinType)
	*p = x
	return p
}

func (x CoinType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CoinType) Descriptor() protoreflect.EnumDescriptor {
	return file_crypto_proto_enumTypes[0].Descriptor()
}

func (CoinType) Type() protoreflect.EnumType {
	return &file_crypto_proto_enumTypes[0]
}

func (x CoinType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CoinType.Descriptor instead.
func (CoinType) EnumDescriptor() ([]byte, []int) {
	return file_crypto_proto_rawDescGZIP(), []int{0}
}

type XmrKeysUpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrivViewKey string `protobuf:"bytes,1,opt,name=privViewKey,proto3" json:"privViewKey,omitempty"`
	PubSpendKey string `protobuf:"bytes,2,opt,name=pubSpendKey,proto3" json:"pubSpendKey,omitempty"`
}

func (x *XmrKeysUpdateRequest) Reset() {
	*x = XmrKeysUpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_crypto_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *XmrKeysUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*XmrKeysUpdateRequest) ProtoMessage() {}

func (x *XmrKeysUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_crypto_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use XmrKeysUpdateRequest.ProtoReflect.Descriptor instead.
func (*XmrKeysUpdateRequest) Descriptor() ([]byte, []int) {
	return file_crypto_proto_rawDescGZIP(), []int{0}
}

func (x *XmrKeysUpdateRequest) GetPrivViewKey() string {
	if x != nil {
		return x.PrivViewKey
	}
	return ""
}

func (x *XmrKeysUpdateRequest) GetPubSpendKey() string {
	if x != nil {
		return x.PubSpendKey
	}
	return ""
}

type BtcKeysUpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MasterPubKey string `protobuf:"bytes,1,opt,name=masterPubKey,proto3" json:"masterPubKey,omitempty"`
}

func (x *BtcKeysUpdateRequest) Reset() {
	*x = BtcKeysUpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_crypto_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BtcKeysUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BtcKeysUpdateRequest) ProtoMessage() {}

func (x *BtcKeysUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_crypto_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BtcKeysUpdateRequest.ProtoReflect.Descriptor instead.
func (*BtcKeysUpdateRequest) Descriptor() ([]byte, []int) {
	return file_crypto_proto_rawDescGZIP(), []int{1}
}

func (x *BtcKeysUpdateRequest) GetMasterPubKey() string {
	if x != nil {
		return x.MasterPubKey
	}
	return ""
}

type LtcKeysUpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MasterPubKey string `protobuf:"bytes,1,opt,name=masterPubKey,proto3" json:"masterPubKey,omitempty"`
}

func (x *LtcKeysUpdateRequest) Reset() {
	*x = LtcKeysUpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_crypto_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LtcKeysUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LtcKeysUpdateRequest) ProtoMessage() {}

func (x *LtcKeysUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_crypto_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LtcKeysUpdateRequest.ProtoReflect.Descriptor instead.
func (*LtcKeysUpdateRequest) Descriptor() ([]byte, []int) {
	return file_crypto_proto_rawDescGZIP(), []int{2}
}

func (x *LtcKeysUpdateRequest) GetMasterPubKey() string {
	if x != nil {
		return x.MasterPubKey
	}
	return ""
}

type EthKeysUpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MasterPubKey string `protobuf:"bytes,1,opt,name=masterPubKey,proto3" json:"masterPubKey,omitempty"`
}

func (x *EthKeysUpdateRequest) Reset() {
	*x = EthKeysUpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_crypto_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EthKeysUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EthKeysUpdateRequest) ProtoMessage() {}

func (x *EthKeysUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_crypto_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EthKeysUpdateRequest.ProtoReflect.Descriptor instead.
func (*EthKeysUpdateRequest) Descriptor() ([]byte, []int) {
	return file_crypto_proto_rawDescGZIP(), []int{3}
}

func (x *EthKeysUpdateRequest) GetMasterPubKey() string {
	if x != nil {
		return x.MasterPubKey
	}
	return ""
}

type BnbKeysUpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MasterPubKey string `protobuf:"bytes,1,opt,name=masterPubKey,proto3" json:"masterPubKey,omitempty"`
}

func (x *BnbKeysUpdateRequest) Reset() {
	*x = BnbKeysUpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_crypto_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BnbKeysUpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BnbKeysUpdateRequest) ProtoMessage() {}

func (x *BnbKeysUpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_crypto_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BnbKeysUpdateRequest.ProtoReflect.Descriptor instead.
func (*BnbKeysUpdateRequest) Descriptor() ([]byte, []int) {
	return file_crypto_proto_rawDescGZIP(), []int{4}
}

func (x *BnbKeysUpdateRequest) GetMasterPubKey() string {
	if x != nil {
		return x.MasterPubKey
	}
	return ""
}

var File_crypto_proto protoreflect.FileDescriptor

var file_crypto_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e, 0x76, 0x31, 0x22, 0x5a, 0x0a, 0x14, 0x58, 0x6d, 0x72,
	0x4b, 0x65, 0x79, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x72, 0x69, 0x76, 0x56, 0x69, 0x65, 0x77, 0x4b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x69, 0x76, 0x56, 0x69, 0x65, 0x77,
	0x4b, 0x65, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x75, 0x62, 0x53, 0x70, 0x65, 0x6e, 0x64, 0x4b,
	0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x75, 0x62, 0x53, 0x70, 0x65,
	0x6e, 0x64, 0x4b, 0x65, 0x79, 0x22, 0x3a, 0x0a, 0x14, 0x42, 0x74, 0x63, 0x4b, 0x65, 0x79, 0x73,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a,
	0x0c, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65,
	0x79, 0x22, 0x3a, 0x0a, 0x14, 0x4c, 0x74, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x6d, 0x61, 0x73,
	0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x22, 0x3a, 0x0a,
	0x14, 0x45, 0x74, 0x68, 0x4b, 0x65, 0x79, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50,
	0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6d, 0x61, 0x73,
	0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x22, 0x3a, 0x0a, 0x14, 0x42, 0x6e, 0x62,
	0x4b, 0x65, 0x79, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x22, 0x0a, 0x0c, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50,
	0x75, 0x62, 0x4b, 0x65, 0x79, 0x2a, 0xf5, 0x04, 0x0a, 0x08, 0x43, 0x6f, 0x69, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x58, 0x4d, 0x52, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x42,
	0x54, 0x43, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x4c, 0x54, 0x43, 0x10, 0x02, 0x12, 0x07, 0x0a,
	0x03, 0x45, 0x54, 0x48, 0x10, 0x03, 0x12, 0x07, 0x0a, 0x03, 0x54, 0x4f, 0x4e, 0x10, 0x04, 0x12,
	0x0e, 0x0a, 0x0a, 0x55, 0x53, 0x44, 0x54, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x05, 0x12,
	0x0e, 0x0a, 0x0a, 0x55, 0x53, 0x44, 0x43, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x06, 0x12,
	0x0d, 0x0a, 0x09, 0x44, 0x41, 0x49, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x07, 0x12, 0x0e,
	0x0a, 0x0a, 0x57, 0x42, 0x54, 0x43, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x08, 0x12, 0x0d,
	0x0a, 0x09, 0x55, 0x4e, 0x49, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x09, 0x12, 0x0e, 0x0a,
	0x0a, 0x4c, 0x49, 0x4e, 0x4b, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x0a, 0x12, 0x0e, 0x0a,
	0x0a, 0x41, 0x41, 0x56, 0x45, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x0b, 0x12, 0x0d, 0x0a,
	0x09, 0x43, 0x52, 0x56, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x0c, 0x12, 0x0f, 0x0a, 0x0b,
	0x4d, 0x41, 0x54, 0x49, 0x43, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x0d, 0x12, 0x0e, 0x0a,
	0x0a, 0x53, 0x48, 0x49, 0x42, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x0e, 0x12, 0x0d, 0x0a,
	0x09, 0x42, 0x4e, 0x42, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x0f, 0x12, 0x0e, 0x0a, 0x0a,
	0x41, 0x54, 0x4f, 0x4d, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x10, 0x12, 0x0d, 0x0a, 0x09,
	0x41, 0x52, 0x42, 0x5f, 0x45, 0x52, 0x43, 0x32, 0x30, 0x10, 0x11, 0x12, 0x07, 0x0a, 0x03, 0x42,
	0x4e, 0x42, 0x10, 0x12, 0x12, 0x10, 0x0a, 0x0c, 0x42, 0x53, 0x43, 0x55, 0x53, 0x44, 0x5f, 0x42,
	0x45, 0x50, 0x32, 0x30, 0x10, 0x13, 0x12, 0x0e, 0x0a, 0x0a, 0x55, 0x53, 0x44, 0x43, 0x5f, 0x42,
	0x45, 0x50, 0x32, 0x30, 0x10, 0x14, 0x12, 0x0d, 0x0a, 0x09, 0x44, 0x41, 0x49, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x15, 0x12, 0x0e, 0x0a, 0x0a, 0x42, 0x55, 0x53, 0x44, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x16, 0x12, 0x0e, 0x0a, 0x0a, 0x57, 0x42, 0x54, 0x43, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x17, 0x12, 0x0e, 0x0a, 0x0a, 0x42, 0x54, 0x43, 0x42, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x18, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x49, 0x5f, 0x42, 0x45, 0x50,
	0x32, 0x30, 0x10, 0x19, 0x12, 0x0e, 0x0a, 0x0a, 0x4c, 0x49, 0x4e, 0x4b, 0x5f, 0x42, 0x45, 0x50,
	0x32, 0x30, 0x10, 0x1a, 0x12, 0x0e, 0x0a, 0x0a, 0x41, 0x41, 0x56, 0x45, 0x5f, 0x42, 0x45, 0x50,
	0x32, 0x30, 0x10, 0x1b, 0x12, 0x0f, 0x0a, 0x0b, 0x4d, 0x41, 0x54, 0x49, 0x43, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x1c, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x48, 0x49, 0x42, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x1d, 0x12, 0x0e, 0x0a, 0x0a, 0x41, 0x54, 0x4f, 0x4d, 0x5f, 0x42, 0x45,
	0x50, 0x32, 0x30, 0x10, 0x1e, 0x12, 0x0d, 0x0a, 0x09, 0x41, 0x52, 0x42, 0x5f, 0x42, 0x45, 0x50,
	0x32, 0x30, 0x10, 0x1f, 0x12, 0x0d, 0x0a, 0x09, 0x45, 0x54, 0x48, 0x5f, 0x42, 0x45, 0x50, 0x32,
	0x30, 0x10, 0x20, 0x12, 0x0d, 0x0a, 0x09, 0x58, 0x52, 0x50, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30,
	0x10, 0x21, 0x12, 0x0d, 0x0a, 0x09, 0x41, 0x44, 0x41, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10,
	0x22, 0x12, 0x0d, 0x0a, 0x09, 0x54, 0x52, 0x58, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x23,
	0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x4f, 0x47, 0x45, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x24,
	0x12, 0x0d, 0x0a, 0x09, 0x4c, 0x54, 0x43, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x25, 0x12,
	0x0d, 0x0a, 0x09, 0x42, 0x43, 0x48, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x26, 0x12, 0x0d,
	0x0a, 0x09, 0x54, 0x57, 0x54, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x27, 0x12, 0x0e, 0x0a,
	0x0a, 0x41, 0x56, 0x41, 0x58, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x28, 0x12, 0x0e, 0x0a,
	0x0a, 0x43, 0x41, 0x4b, 0x45, 0x5f, 0x42, 0x45, 0x50, 0x32, 0x30, 0x10, 0x29, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_crypto_proto_rawDescOnce sync.Once
	file_crypto_proto_rawDescData = file_crypto_proto_rawDesc
)

func file_crypto_proto_rawDescGZIP() []byte {
	file_crypto_proto_rawDescOnce.Do(func() {
		file_crypto_proto_rawDescData = protoimpl.X.CompressGZIP(file_crypto_proto_rawDescData)
	})
	return file_crypto_proto_rawDescData
}

var file_crypto_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_crypto_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_crypto_proto_goTypes = []any{
	(CoinType)(0),                // 0: crypto.v1.CoinType
	(*XmrKeysUpdateRequest)(nil), // 1: crypto.v1.XmrKeysUpdateRequest
	(*BtcKeysUpdateRequest)(nil), // 2: crypto.v1.BtcKeysUpdateRequest
	(*LtcKeysUpdateRequest)(nil), // 3: crypto.v1.LtcKeysUpdateRequest
	(*EthKeysUpdateRequest)(nil), // 4: crypto.v1.EthKeysUpdateRequest
	(*BnbKeysUpdateRequest)(nil), // 5: crypto.v1.BnbKeysUpdateRequest
}
var file_crypto_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_crypto_proto_init() }
func file_crypto_proto_init() {
	if File_crypto_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_crypto_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*XmrKeysUpdateRequest); i {
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
		file_crypto_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*BtcKeysUpdateRequest); i {
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
		file_crypto_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*LtcKeysUpdateRequest); i {
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
		file_crypto_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*EthKeysUpdateRequest); i {
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
		file_crypto_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*BnbKeysUpdateRequest); i {
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
			RawDescriptor: file_crypto_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_crypto_proto_goTypes,
		DependencyIndexes: file_crypto_proto_depIdxs,
		EnumInfos:         file_crypto_proto_enumTypes,
		MessageInfos:      file_crypto_proto_msgTypes,
	}.Build()
	File_crypto_proto = out.File
	file_crypto_proto_rawDesc = nil
	file_crypto_proto_goTypes = nil
	file_crypto_proto_depIdxs = nil
}
