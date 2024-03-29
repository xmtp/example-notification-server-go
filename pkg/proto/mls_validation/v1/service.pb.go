// Message API

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: mls_validation/v1/service.proto

package mls_validationv1

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

// Contains a batch of serialized Key Packages
type ValidateKeyPackagesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeyPackages []*ValidateKeyPackagesRequest_KeyPackage `protobuf:"bytes,1,rep,name=key_packages,json=keyPackages,proto3" json:"key_packages,omitempty"`
}

func (x *ValidateKeyPackagesRequest) Reset() {
	*x = ValidateKeyPackagesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateKeyPackagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesRequest) ProtoMessage() {}

func (x *ValidateKeyPackagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesRequest.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesRequest) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{0}
}

func (x *ValidateKeyPackagesRequest) GetKeyPackages() []*ValidateKeyPackagesRequest_KeyPackage {
	if x != nil {
		return x.KeyPackages
	}
	return nil
}

// Response to ValidateKeyPackagesRequest
type ValidateKeyPackagesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Responses []*ValidateKeyPackagesResponse_ValidationResponse `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
}

func (x *ValidateKeyPackagesResponse) Reset() {
	*x = ValidateKeyPackagesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateKeyPackagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesResponse) ProtoMessage() {}

func (x *ValidateKeyPackagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesResponse.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{1}
}

func (x *ValidateKeyPackagesResponse) GetResponses() []*ValidateKeyPackagesResponse_ValidationResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Contains a batch of serialized Group Messages
type ValidateGroupMessagesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupMessages []*ValidateGroupMessagesRequest_GroupMessage `protobuf:"bytes,1,rep,name=group_messages,json=groupMessages,proto3" json:"group_messages,omitempty"`
}

func (x *ValidateGroupMessagesRequest) Reset() {
	*x = ValidateGroupMessagesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateGroupMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesRequest) ProtoMessage() {}

func (x *ValidateGroupMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesRequest.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesRequest) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{2}
}

func (x *ValidateGroupMessagesRequest) GetGroupMessages() []*ValidateGroupMessagesRequest_GroupMessage {
	if x != nil {
		return x.GroupMessages
	}
	return nil
}

// Response to ValidateGroupMessagesRequest
type ValidateGroupMessagesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Responses []*ValidateGroupMessagesResponse_ValidationResponse `protobuf:"bytes,1,rep,name=responses,proto3" json:"responses,omitempty"`
}

func (x *ValidateGroupMessagesResponse) Reset() {
	*x = ValidateGroupMessagesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateGroupMessagesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesResponse) ProtoMessage() {}

func (x *ValidateGroupMessagesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesResponse.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{3}
}

func (x *ValidateGroupMessagesResponse) GetResponses() []*ValidateGroupMessagesResponse_ValidationResponse {
	if x != nil {
		return x.Responses
	}
	return nil
}

// Wrapper for each key package
type ValidateKeyPackagesRequest_KeyPackage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeyPackageBytesTlsSerialized []byte `protobuf:"bytes,1,opt,name=key_package_bytes_tls_serialized,json=keyPackageBytesTlsSerialized,proto3" json:"key_package_bytes_tls_serialized,omitempty"`
}

func (x *ValidateKeyPackagesRequest_KeyPackage) Reset() {
	*x = ValidateKeyPackagesRequest_KeyPackage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateKeyPackagesRequest_KeyPackage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesRequest_KeyPackage) ProtoMessage() {}

func (x *ValidateKeyPackagesRequest_KeyPackage) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesRequest_KeyPackage.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesRequest_KeyPackage) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ValidateKeyPackagesRequest_KeyPackage) GetKeyPackageBytesTlsSerialized() []byte {
	if x != nil {
		return x.KeyPackageBytesTlsSerialized
	}
	return nil
}

// An individual response to one key package
type ValidateKeyPackagesResponse_ValidationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsOk                    bool   `protobuf:"varint,1,opt,name=is_ok,json=isOk,proto3" json:"is_ok,omitempty"`
	ErrorMessage            string `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	InstallationId          []byte `protobuf:"bytes,3,opt,name=installation_id,json=installationId,proto3" json:"installation_id,omitempty"`
	AccountAddress          string `protobuf:"bytes,4,opt,name=account_address,json=accountAddress,proto3" json:"account_address,omitempty"`
	CredentialIdentityBytes []byte `protobuf:"bytes,5,opt,name=credential_identity_bytes,json=credentialIdentityBytes,proto3" json:"credential_identity_bytes,omitempty"`
	Expiration              uint64 `protobuf:"varint,6,opt,name=expiration,proto3" json:"expiration,omitempty"`
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) Reset() {
	*x = ValidateKeyPackagesResponse_ValidationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateKeyPackagesResponse_ValidationResponse) ProtoMessage() {}

func (x *ValidateKeyPackagesResponse_ValidationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateKeyPackagesResponse_ValidationResponse.ProtoReflect.Descriptor instead.
func (*ValidateKeyPackagesResponse_ValidationResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{1, 0}
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetIsOk() bool {
	if x != nil {
		return x.IsOk
	}
	return false
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetInstallationId() []byte {
	if x != nil {
		return x.InstallationId
	}
	return nil
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetAccountAddress() string {
	if x != nil {
		return x.AccountAddress
	}
	return ""
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetCredentialIdentityBytes() []byte {
	if x != nil {
		return x.CredentialIdentityBytes
	}
	return nil
}

func (x *ValidateKeyPackagesResponse_ValidationResponse) GetExpiration() uint64 {
	if x != nil {
		return x.Expiration
	}
	return 0
}

// Wrapper for each message
type ValidateGroupMessagesRequest_GroupMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GroupMessageBytesTlsSerialized []byte `protobuf:"bytes,1,opt,name=group_message_bytes_tls_serialized,json=groupMessageBytesTlsSerialized,proto3" json:"group_message_bytes_tls_serialized,omitempty"`
}

func (x *ValidateGroupMessagesRequest_GroupMessage) Reset() {
	*x = ValidateGroupMessagesRequest_GroupMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateGroupMessagesRequest_GroupMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesRequest_GroupMessage) ProtoMessage() {}

func (x *ValidateGroupMessagesRequest_GroupMessage) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesRequest_GroupMessage.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesRequest_GroupMessage) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{2, 0}
}

func (x *ValidateGroupMessagesRequest_GroupMessage) GetGroupMessageBytesTlsSerialized() []byte {
	if x != nil {
		return x.GroupMessageBytesTlsSerialized
	}
	return nil
}

// An individual response to one message
type ValidateGroupMessagesResponse_ValidationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsOk         bool   `protobuf:"varint,1,opt,name=is_ok,json=isOk,proto3" json:"is_ok,omitempty"`
	ErrorMessage string `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	GroupId      string `protobuf:"bytes,3,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) Reset() {
	*x = ValidateGroupMessagesResponse_ValidationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_validation_v1_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidateGroupMessagesResponse_ValidationResponse) ProtoMessage() {}

func (x *ValidateGroupMessagesResponse_ValidationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mls_validation_v1_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidateGroupMessagesResponse_ValidationResponse.ProtoReflect.Descriptor instead.
func (*ValidateGroupMessagesResponse_ValidationResponse) Descriptor() ([]byte, []int) {
	return file_mls_validation_v1_service_proto_rawDescGZIP(), []int{3, 0}
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) GetIsOk() bool {
	if x != nil {
		return x.IsOk
	}
	return false
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *ValidateGroupMessagesResponse_ValidationResponse) GetGroupId() string {
	if x != nil {
		return x.GroupId
	}
	return ""
}

var File_mls_validation_v1_service_proto protoreflect.FileDescriptor

var file_mls_validation_v1_service_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x16, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x22, 0xd4, 0x01, 0x0a, 0x1a, 0x56, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x60, 0x0a, 0x0c, 0x6b, 0x65, 0x79, 0x5f,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3d,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x4b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2e, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52, 0x0b, 0x6b,
	0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x1a, 0x54, 0x0a, 0x0a, 0x4b, 0x65,
	0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x46, 0x0a, 0x20, 0x6b, 0x65, 0x79, 0x5f,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x74, 0x6c,
	0x73, 0x5f, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x1c, 0x6b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x54, 0x6c, 0x73, 0x53, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64,
	0x22, 0x82, 0x03, 0x0a, 0x1b, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79,
	0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x64, 0x0a, 0x09, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x46, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x09, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x1a, 0xfc, 0x01, 0x0a, 0x12, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x13, 0x0a,
	0x05, 0x69, 0x73, 0x5f, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x69, 0x73,
	0x4f, 0x6b, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0e, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x12, 0x27, 0x0a, 0x0f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x3a, 0x0a, 0x19, 0x63, 0x72, 0x65,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x17, 0x63, 0x72,
	0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xe4, 0x01, 0x0a, 0x1c, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x68, 0x0a, 0x0e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x41,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x52, 0x0d, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x1a, 0x5a, 0x0a, 0x0c, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x4a, 0x0a, 0x22, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x74, 0x6c, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x69,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x1e, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x54,
	0x6c, 0x73, 0x53, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x22, 0xf2, 0x01, 0x0a,
	0x1d, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x66,
	0x0a, 0x09, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x48, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x09, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x1a, 0x69, 0x0a, 0x12, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x13, 0x0a, 0x05,
	0x69, 0x73, 0x5f, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x69, 0x73, 0x4f,
	0x6b, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49,
	0x64, 0x32, 0x9b, 0x02, 0x0a, 0x0d, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x41, 0x70, 0x69, 0x12, 0x80, 0x01, 0x0a, 0x13, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x4b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x12, 0x32, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79,
	0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x33, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x86, 0x01, 0x0a, 0x15, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x12, 0x34, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x35, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c,
	0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42,
	0x97, 0x02, 0x0a, 0x34, 0x6f, 0x72, 0x67, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x61, 0x6e, 0x64,
	0x72, 0x6f, 0x69, 0x64, 0x2e, 0x6c, 0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x2e, 0x70, 0x75, 0x73,
	0x68, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x42, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x5b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2d, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2d, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x6c, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x58, 0xaa, 0x02, 0x15, 0x58, 0x6d,
	0x74, 0x70, 0x2e, 0x4d, 0x6c, 0x73, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x56, 0x31, 0xca, 0x02, 0x15, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x56, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x21, 0x58, 0x6d,
	0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x17, 0x58, 0x6d, 0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x6c, 0x73, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_mls_validation_v1_service_proto_rawDescOnce sync.Once
	file_mls_validation_v1_service_proto_rawDescData = file_mls_validation_v1_service_proto_rawDesc
)

func file_mls_validation_v1_service_proto_rawDescGZIP() []byte {
	file_mls_validation_v1_service_proto_rawDescOnce.Do(func() {
		file_mls_validation_v1_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_mls_validation_v1_service_proto_rawDescData)
	})
	return file_mls_validation_v1_service_proto_rawDescData
}

var file_mls_validation_v1_service_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_mls_validation_v1_service_proto_goTypes = []interface{}{
	(*ValidateKeyPackagesRequest)(nil),                       // 0: xmtp.mls_validation.v1.ValidateKeyPackagesRequest
	(*ValidateKeyPackagesResponse)(nil),                      // 1: xmtp.mls_validation.v1.ValidateKeyPackagesResponse
	(*ValidateGroupMessagesRequest)(nil),                     // 2: xmtp.mls_validation.v1.ValidateGroupMessagesRequest
	(*ValidateGroupMessagesResponse)(nil),                    // 3: xmtp.mls_validation.v1.ValidateGroupMessagesResponse
	(*ValidateKeyPackagesRequest_KeyPackage)(nil),            // 4: xmtp.mls_validation.v1.ValidateKeyPackagesRequest.KeyPackage
	(*ValidateKeyPackagesResponse_ValidationResponse)(nil),   // 5: xmtp.mls_validation.v1.ValidateKeyPackagesResponse.ValidationResponse
	(*ValidateGroupMessagesRequest_GroupMessage)(nil),        // 6: xmtp.mls_validation.v1.ValidateGroupMessagesRequest.GroupMessage
	(*ValidateGroupMessagesResponse_ValidationResponse)(nil), // 7: xmtp.mls_validation.v1.ValidateGroupMessagesResponse.ValidationResponse
}
var file_mls_validation_v1_service_proto_depIdxs = []int32{
	4, // 0: xmtp.mls_validation.v1.ValidateKeyPackagesRequest.key_packages:type_name -> xmtp.mls_validation.v1.ValidateKeyPackagesRequest.KeyPackage
	5, // 1: xmtp.mls_validation.v1.ValidateKeyPackagesResponse.responses:type_name -> xmtp.mls_validation.v1.ValidateKeyPackagesResponse.ValidationResponse
	6, // 2: xmtp.mls_validation.v1.ValidateGroupMessagesRequest.group_messages:type_name -> xmtp.mls_validation.v1.ValidateGroupMessagesRequest.GroupMessage
	7, // 3: xmtp.mls_validation.v1.ValidateGroupMessagesResponse.responses:type_name -> xmtp.mls_validation.v1.ValidateGroupMessagesResponse.ValidationResponse
	0, // 4: xmtp.mls_validation.v1.ValidationApi.ValidateKeyPackages:input_type -> xmtp.mls_validation.v1.ValidateKeyPackagesRequest
	2, // 5: xmtp.mls_validation.v1.ValidationApi.ValidateGroupMessages:input_type -> xmtp.mls_validation.v1.ValidateGroupMessagesRequest
	1, // 6: xmtp.mls_validation.v1.ValidationApi.ValidateKeyPackages:output_type -> xmtp.mls_validation.v1.ValidateKeyPackagesResponse
	3, // 7: xmtp.mls_validation.v1.ValidationApi.ValidateGroupMessages:output_type -> xmtp.mls_validation.v1.ValidateGroupMessagesResponse
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_mls_validation_v1_service_proto_init() }
func file_mls_validation_v1_service_proto_init() {
	if File_mls_validation_v1_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mls_validation_v1_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateKeyPackagesRequest); i {
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
		file_mls_validation_v1_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateKeyPackagesResponse); i {
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
		file_mls_validation_v1_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateGroupMessagesRequest); i {
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
		file_mls_validation_v1_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateGroupMessagesResponse); i {
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
		file_mls_validation_v1_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateKeyPackagesRequest_KeyPackage); i {
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
		file_mls_validation_v1_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateKeyPackagesResponse_ValidationResponse); i {
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
		file_mls_validation_v1_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateGroupMessagesRequest_GroupMessage); i {
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
		file_mls_validation_v1_service_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidateGroupMessagesResponse_ValidationResponse); i {
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
			RawDescriptor: file_mls_validation_v1_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mls_validation_v1_service_proto_goTypes,
		DependencyIndexes: file_mls_validation_v1_service_proto_depIdxs,
		MessageInfos:      file_mls_validation_v1_service_proto_msgTypes,
	}.Build()
	File_mls_validation_v1_service_proto = out.File
	file_mls_validation_v1_service_proto_rawDesc = nil
	file_mls_validation_v1_service_proto_goTypes = nil
	file_mls_validation_v1_service_proto_depIdxs = nil
}
