// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: wso2/discovery/api/endpoint_security.proto

package api

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

// Endpoint config model
// todo(amali) fix endpoint security
type EndpointSecurity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SandBoxSecurityInfo    *SecurityInfo `protobuf:"bytes,1,opt,name=SandBoxSecurityInfo,proto3" json:"SandBoxSecurityInfo,omitempty"`
	ProductionSecurityInfo *SecurityInfo `protobuf:"bytes,2,opt,name=ProductionSecurityInfo,proto3" json:"ProductionSecurityInfo,omitempty"`
}

func (x *EndpointSecurity) Reset() {
	*x = EndpointSecurity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_wso2_discovery_api_endpoint_security_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EndpointSecurity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EndpointSecurity) ProtoMessage() {}

func (x *EndpointSecurity) ProtoReflect() protoreflect.Message {
	mi := &file_wso2_discovery_api_endpoint_security_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EndpointSecurity.ProtoReflect.Descriptor instead.
func (*EndpointSecurity) Descriptor() ([]byte, []int) {
	return file_wso2_discovery_api_endpoint_security_proto_rawDescGZIP(), []int{0}
}

func (x *EndpointSecurity) GetSandBoxSecurityInfo() *SecurityInfo {
	if x != nil {
		return x.SandBoxSecurityInfo
	}
	return nil
}

func (x *EndpointSecurity) GetProductionSecurityInfo() *SecurityInfo {
	if x != nil {
		return x.ProductionSecurityInfo
	}
	return nil
}

var File_wso2_discovery_api_endpoint_security_proto protoreflect.FileDescriptor

var file_wso2_discovery_api_endpoint_security_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x77, 0x73, 0x6f, 0x32, 0x2f, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x73, 0x65,
	0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x77, 0x73,
	0x6f, 0x32, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x69,
	0x1a, 0x26, 0x77, 0x73, 0x6f, 0x32, 0x2f, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x5f, 0x69, 0x6e,
	0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc0, 0x01, 0x0a, 0x10, 0x45, 0x6e, 0x64,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x12, 0x52, 0x0a,
	0x13, 0x53, 0x61, 0x6e, 0x64, 0x42, 0x6f, 0x78, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79,
	0x49, 0x6e, 0x66, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x77, 0x73, 0x6f,
	0x32, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x13, 0x53, 0x61,
	0x6e, 0x64, 0x42, 0x6f, 0x78, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x58, 0x0a, 0x16, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x77, 0x73, 0x6f, 0x32, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x79, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x16, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x42, 0x7f, 0x0a, 0x25, 0x6f,
	0x72, 0x67, 0x2e, 0x77, 0x73, 0x6f, 0x32, 0x2e, 0x63, 0x68, 0x6f, 0x72, 0x65, 0x6f, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x2e, 0x61, 0x70, 0x69, 0x42, 0x15, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x65,
	0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3d, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70,
	0x72, 0x6f, 0x78, 0x79, 0x2f, 0x67, 0x6f, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2d,
	0x70, 0x6c, 0x61, 0x6e, 0x65, 0x2f, 0x77, 0x73, 0x6f, 0x32, 0x2f, 0x64, 0x69, 0x73, 0x63, 0x6f,
	0x76, 0x65, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x3b, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_wso2_discovery_api_endpoint_security_proto_rawDescOnce sync.Once
	file_wso2_discovery_api_endpoint_security_proto_rawDescData = file_wso2_discovery_api_endpoint_security_proto_rawDesc
)

func file_wso2_discovery_api_endpoint_security_proto_rawDescGZIP() []byte {
	file_wso2_discovery_api_endpoint_security_proto_rawDescOnce.Do(func() {
		file_wso2_discovery_api_endpoint_security_proto_rawDescData = protoimpl.X.CompressGZIP(file_wso2_discovery_api_endpoint_security_proto_rawDescData)
	})
	return file_wso2_discovery_api_endpoint_security_proto_rawDescData
}

var file_wso2_discovery_api_endpoint_security_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_wso2_discovery_api_endpoint_security_proto_goTypes = []interface{}{
	(*EndpointSecurity)(nil), // 0: wso2.discovery.api.EndpointSecurity
	(*SecurityInfo)(nil),     // 1: wso2.discovery.api.SecurityInfo
}
var file_wso2_discovery_api_endpoint_security_proto_depIdxs = []int32{
	1, // 0: wso2.discovery.api.EndpointSecurity.SandBoxSecurityInfo:type_name -> wso2.discovery.api.SecurityInfo
	1, // 1: wso2.discovery.api.EndpointSecurity.ProductionSecurityInfo:type_name -> wso2.discovery.api.SecurityInfo
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_wso2_discovery_api_endpoint_security_proto_init() }
func file_wso2_discovery_api_endpoint_security_proto_init() {
	if File_wso2_discovery_api_endpoint_security_proto != nil {
		return
	}
	file_wso2_discovery_api_security_info_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_wso2_discovery_api_endpoint_security_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EndpointSecurity); i {
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
			RawDescriptor: file_wso2_discovery_api_endpoint_security_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_wso2_discovery_api_endpoint_security_proto_goTypes,
		DependencyIndexes: file_wso2_discovery_api_endpoint_security_proto_depIdxs,
		MessageInfos:      file_wso2_discovery_api_endpoint_security_proto_msgTypes,
	}.Build()
	File_wso2_discovery_api_endpoint_security_proto = out.File
	file_wso2_discovery_api_endpoint_security_proto_rawDesc = nil
	file_wso2_discovery_api_endpoint_security_proto_goTypes = nil
	file_wso2_discovery_api_endpoint_security_proto_depIdxs = nil
}
