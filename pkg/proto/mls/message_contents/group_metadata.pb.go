// Group metadata

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: mls/message_contents/group_metadata.proto

package message_contents

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

// Defines the type of conversation
type ConversationType int32

const (
	ConversationType_CONVERSATION_TYPE_UNSPECIFIED ConversationType = 0
	ConversationType_CONVERSATION_TYPE_GROUP       ConversationType = 1
	ConversationType_CONVERSATION_TYPE_DM          ConversationType = 2
)

// Enum value maps for ConversationType.
var (
	ConversationType_name = map[int32]string{
		0: "CONVERSATION_TYPE_UNSPECIFIED",
		1: "CONVERSATION_TYPE_GROUP",
		2: "CONVERSATION_TYPE_DM",
	}
	ConversationType_value = map[string]int32{
		"CONVERSATION_TYPE_UNSPECIFIED": 0,
		"CONVERSATION_TYPE_GROUP":       1,
		"CONVERSATION_TYPE_DM":          2,
	}
)

func (x ConversationType) Enum() *ConversationType {
	p := new(ConversationType)
	*p = x
	return p
}

func (x ConversationType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConversationType) Descriptor() protoreflect.EnumDescriptor {
	return file_mls_message_contents_group_metadata_proto_enumTypes[0].Descriptor()
}

func (ConversationType) Type() protoreflect.EnumType {
	return &file_mls_message_contents_group_metadata_proto_enumTypes[0]
}

func (x ConversationType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConversationType.Descriptor instead.
func (ConversationType) EnumDescriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{0}
}

// Base policy
type MembershipPolicy_BasePolicy int32

const (
	MembershipPolicy_BASE_POLICY_UNSPECIFIED            MembershipPolicy_BasePolicy = 0
	MembershipPolicy_BASE_POLICY_ALLOW                  MembershipPolicy_BasePolicy = 1
	MembershipPolicy_BASE_POLICY_DENY                   MembershipPolicy_BasePolicy = 2
	MembershipPolicy_BASE_POLICY_ALLOW_IF_ACTOR_CREATOR MembershipPolicy_BasePolicy = 3
)

// Enum value maps for MembershipPolicy_BasePolicy.
var (
	MembershipPolicy_BasePolicy_name = map[int32]string{
		0: "BASE_POLICY_UNSPECIFIED",
		1: "BASE_POLICY_ALLOW",
		2: "BASE_POLICY_DENY",
		3: "BASE_POLICY_ALLOW_IF_ACTOR_CREATOR",
	}
	MembershipPolicy_BasePolicy_value = map[string]int32{
		"BASE_POLICY_UNSPECIFIED":            0,
		"BASE_POLICY_ALLOW":                  1,
		"BASE_POLICY_DENY":                   2,
		"BASE_POLICY_ALLOW_IF_ACTOR_CREATOR": 3,
	}
)

func (x MembershipPolicy_BasePolicy) Enum() *MembershipPolicy_BasePolicy {
	p := new(MembershipPolicy_BasePolicy)
	*p = x
	return p
}

func (x MembershipPolicy_BasePolicy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MembershipPolicy_BasePolicy) Descriptor() protoreflect.EnumDescriptor {
	return file_mls_message_contents_group_metadata_proto_enumTypes[1].Descriptor()
}

func (MembershipPolicy_BasePolicy) Type() protoreflect.EnumType {
	return &file_mls_message_contents_group_metadata_proto_enumTypes[1]
}

func (x MembershipPolicy_BasePolicy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MembershipPolicy_BasePolicy.Descriptor instead.
func (MembershipPolicy_BasePolicy) EnumDescriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{2, 0}
}

// Parent message for group metadata
type GroupMetadataV1 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ConversationType      ConversationType `protobuf:"varint,1,opt,name=conversation_type,json=conversationType,proto3,enum=xmtp.mls.message_contents.ConversationType" json:"conversation_type,omitempty"`
	CreatorAccountAddress string           `protobuf:"bytes,2,opt,name=creator_account_address,json=creatorAccountAddress,proto3" json:"creator_account_address,omitempty"`
	Policies              *PolicySet       `protobuf:"bytes,3,opt,name=policies,proto3" json:"policies,omitempty"`
}

func (x *GroupMetadataV1) Reset() {
	*x = GroupMetadataV1{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_group_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GroupMetadataV1) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupMetadataV1) ProtoMessage() {}

func (x *GroupMetadataV1) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_group_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupMetadataV1.ProtoReflect.Descriptor instead.
func (*GroupMetadataV1) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *GroupMetadataV1) GetConversationType() ConversationType {
	if x != nil {
		return x.ConversationType
	}
	return ConversationType_CONVERSATION_TYPE_UNSPECIFIED
}

func (x *GroupMetadataV1) GetCreatorAccountAddress() string {
	if x != nil {
		return x.CreatorAccountAddress
	}
	return ""
}

func (x *GroupMetadataV1) GetPolicies() *PolicySet {
	if x != nil {
		return x.Policies
	}
	return nil
}

// The set of policies that govern the group
type PolicySet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AddMemberPolicy    *MembershipPolicy `protobuf:"bytes,1,opt,name=add_member_policy,json=addMemberPolicy,proto3" json:"add_member_policy,omitempty"`
	RemoveMemberPolicy *MembershipPolicy `protobuf:"bytes,2,opt,name=remove_member_policy,json=removeMemberPolicy,proto3" json:"remove_member_policy,omitempty"`
}

func (x *PolicySet) Reset() {
	*x = PolicySet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_group_metadata_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PolicySet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolicySet) ProtoMessage() {}

func (x *PolicySet) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_group_metadata_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolicySet.ProtoReflect.Descriptor instead.
func (*PolicySet) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{1}
}

func (x *PolicySet) GetAddMemberPolicy() *MembershipPolicy {
	if x != nil {
		return x.AddMemberPolicy
	}
	return nil
}

func (x *PolicySet) GetRemoveMemberPolicy() *MembershipPolicy {
	if x != nil {
		return x.RemoveMemberPolicy
	}
	return nil
}

// A policy that governs adding/removing members or installations
type MembershipPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Kind:
	//
	//	*MembershipPolicy_Base
	//	*MembershipPolicy_AndCondition_
	//	*MembershipPolicy_AnyCondition_
	Kind isMembershipPolicy_Kind `protobuf_oneof:"kind"`
}

func (x *MembershipPolicy) Reset() {
	*x = MembershipPolicy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_group_metadata_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MembershipPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MembershipPolicy) ProtoMessage() {}

func (x *MembershipPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_group_metadata_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MembershipPolicy.ProtoReflect.Descriptor instead.
func (*MembershipPolicy) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{2}
}

func (m *MembershipPolicy) GetKind() isMembershipPolicy_Kind {
	if m != nil {
		return m.Kind
	}
	return nil
}

func (x *MembershipPolicy) GetBase() MembershipPolicy_BasePolicy {
	if x, ok := x.GetKind().(*MembershipPolicy_Base); ok {
		return x.Base
	}
	return MembershipPolicy_BASE_POLICY_UNSPECIFIED
}

func (x *MembershipPolicy) GetAndCondition() *MembershipPolicy_AndCondition {
	if x, ok := x.GetKind().(*MembershipPolicy_AndCondition_); ok {
		return x.AndCondition
	}
	return nil
}

func (x *MembershipPolicy) GetAnyCondition() *MembershipPolicy_AnyCondition {
	if x, ok := x.GetKind().(*MembershipPolicy_AnyCondition_); ok {
		return x.AnyCondition
	}
	return nil
}

type isMembershipPolicy_Kind interface {
	isMembershipPolicy_Kind()
}

type MembershipPolicy_Base struct {
	Base MembershipPolicy_BasePolicy `protobuf:"varint,1,opt,name=base,proto3,enum=xmtp.mls.message_contents.MembershipPolicy_BasePolicy,oneof"`
}

type MembershipPolicy_AndCondition_ struct {
	AndCondition *MembershipPolicy_AndCondition `protobuf:"bytes,2,opt,name=and_condition,json=andCondition,proto3,oneof"`
}

type MembershipPolicy_AnyCondition_ struct {
	AnyCondition *MembershipPolicy_AnyCondition `protobuf:"bytes,3,opt,name=any_condition,json=anyCondition,proto3,oneof"`
}

func (*MembershipPolicy_Base) isMembershipPolicy_Kind() {}

func (*MembershipPolicy_AndCondition_) isMembershipPolicy_Kind() {}

func (*MembershipPolicy_AnyCondition_) isMembershipPolicy_Kind() {}

// Combine multiple policies. All must evaluate to true
type MembershipPolicy_AndCondition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Policies []*MembershipPolicy `protobuf:"bytes,1,rep,name=policies,proto3" json:"policies,omitempty"`
}

func (x *MembershipPolicy_AndCondition) Reset() {
	*x = MembershipPolicy_AndCondition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_group_metadata_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MembershipPolicy_AndCondition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MembershipPolicy_AndCondition) ProtoMessage() {}

func (x *MembershipPolicy_AndCondition) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_group_metadata_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MembershipPolicy_AndCondition.ProtoReflect.Descriptor instead.
func (*MembershipPolicy_AndCondition) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{2, 0}
}

func (x *MembershipPolicy_AndCondition) GetPolicies() []*MembershipPolicy {
	if x != nil {
		return x.Policies
	}
	return nil
}

// Combine multiple policies. Any must evaluate to true
type MembershipPolicy_AnyCondition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Policies []*MembershipPolicy `protobuf:"bytes,1,rep,name=policies,proto3" json:"policies,omitempty"`
}

func (x *MembershipPolicy_AnyCondition) Reset() {
	*x = MembershipPolicy_AnyCondition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mls_message_contents_group_metadata_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MembershipPolicy_AnyCondition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MembershipPolicy_AnyCondition) ProtoMessage() {}

func (x *MembershipPolicy_AnyCondition) ProtoReflect() protoreflect.Message {
	mi := &file_mls_message_contents_group_metadata_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MembershipPolicy_AnyCondition.ProtoReflect.Descriptor instead.
func (*MembershipPolicy_AnyCondition) Descriptor() ([]byte, []int) {
	return file_mls_message_contents_group_metadata_proto_rawDescGZIP(), []int{2, 1}
}

func (x *MembershipPolicy_AnyCondition) GetPolicies() []*MembershipPolicy {
	if x != nil {
		return x.Policies
	}
	return nil
}

var File_mls_message_contents_group_metadata_proto protoreflect.FileDescriptor

var file_mls_message_contents_group_metadata_proto_rawDesc = []byte{
	0x0a, 0x29, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x78, 0x6d, 0x74,
	0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xe5, 0x01, 0x0a, 0x0f, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x56, 0x31, 0x12, 0x58, 0x0a, 0x11, 0x63, 0x6f,
	0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73,
	0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x10, 0x63, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x36, 0x0a, 0x17, 0x63, 0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x5f,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x15, 0x63, 0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x40, 0x0a, 0x08,
	0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x50, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x53, 0x65, 0x74, 0x52, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x22, 0xc3,
	0x01, 0x0a, 0x09, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x53, 0x65, 0x74, 0x12, 0x57, 0x0a, 0x11,
	0x61, 0x64, 0x64, 0x5f, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d,
	0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x52, 0x0f, 0x61, 0x64, 0x64, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x5d, 0x0a, 0x14, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x5f,
	0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79,
	0x52, 0x12, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x22, 0xdc, 0x04, 0x0a, 0x10, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73,
	0x68, 0x69, 0x70, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x4c, 0x0a, 0x04, 0x62, 0x61, 0x73,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x36, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d,
	0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x2e, 0x42, 0x61, 0x73, 0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x48,
	0x00, 0x52, 0x04, 0x62, 0x61, 0x73, 0x65, 0x12, 0x5f, 0x0a, 0x0d, 0x61, 0x6e, 0x64, 0x5f, 0x63,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x38,
	0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x68, 0x69, 0x70, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x2e, 0x41, 0x6e, 0x64, 0x43,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x0c, 0x61, 0x6e, 0x64, 0x43,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x5f, 0x0a, 0x0d, 0x61, 0x6e, 0x79, 0x5f,
	0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x38, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x2e, 0x41, 0x6e, 0x79,
	0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x0c, 0x61, 0x6e, 0x79,
	0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x57, 0x0a, 0x0c, 0x41, 0x6e, 0x64,
	0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x47, 0x0a, 0x08, 0x70, 0x6f, 0x6c,
	0x69, 0x63, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x6d,
	0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68,
	0x69, 0x70, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69,
	0x65, 0x73, 0x1a, 0x57, 0x0a, 0x0c, 0x41, 0x6e, 0x79, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x47, 0x0a, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d, 0x6c, 0x73, 0x2e,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73,
	0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x50, 0x6f, 0x6c, 0x69, 0x63,
	0x79, 0x52, 0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x22, 0x7e, 0x0a, 0x0a, 0x42,
	0x61, 0x73, 0x65, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x1b, 0x0a, 0x17, 0x42, 0x41, 0x53,
	0x45, 0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49,
	0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x50,
	0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x10, 0x01, 0x12, 0x14, 0x0a,
	0x10, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x50, 0x4f, 0x4c, 0x49, 0x43, 0x59, 0x5f, 0x44, 0x45, 0x4e,
	0x59, 0x10, 0x02, 0x12, 0x26, 0x0a, 0x22, 0x42, 0x41, 0x53, 0x45, 0x5f, 0x50, 0x4f, 0x4c, 0x49,
	0x43, 0x59, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x5f, 0x49, 0x46, 0x5f, 0x41, 0x43, 0x54, 0x4f,
	0x52, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x4f, 0x52, 0x10, 0x03, 0x42, 0x06, 0x0a, 0x04, 0x6b,
	0x69, 0x6e, 0x64, 0x2a, 0x6c, 0x0a, 0x10, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x21, 0x0a, 0x1d, 0x43, 0x4f, 0x4e, 0x56, 0x45,
	0x52, 0x53, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1b, 0x0a, 0x17, 0x43, 0x4f,
	0x4e, 0x56, 0x45, 0x52, 0x53, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x47, 0x52, 0x4f, 0x55, 0x50, 0x10, 0x01, 0x12, 0x18, 0x0a, 0x14, 0x43, 0x4f, 0x4e, 0x56, 0x45,
	0x52, 0x53, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x44, 0x4d, 0x10,
	0x02, 0x42, 0x84, 0x02, 0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e, 0x78, 0x6d, 0x74, 0x70, 0x2e, 0x6d,
	0x6c, 0x73, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x6e, 0x74, 0x73, 0x42, 0x12, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4d, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x6d, 0x74, 0x70, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x2d, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2d,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2d, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6c, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xa2, 0x02, 0x03, 0x58, 0x4d, 0x4d, 0xaa, 0x02,
	0x18, 0x58, 0x6d, 0x74, 0x70, 0x2e, 0x4d, 0x6c, 0x73, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0xca, 0x02, 0x18, 0x58, 0x6d, 0x74, 0x70,
	0x5c, 0x4d, 0x6c, 0x73, 0x5c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x73, 0xe2, 0x02, 0x24, 0x58, 0x6d, 0x74, 0x70, 0x5c, 0x4d, 0x6c, 0x73, 0x5c,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1a, 0x58, 0x6d,
	0x74, 0x70, 0x3a, 0x3a, 0x4d, 0x6c, 0x73, 0x3a, 0x3a, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mls_message_contents_group_metadata_proto_rawDescOnce sync.Once
	file_mls_message_contents_group_metadata_proto_rawDescData = file_mls_message_contents_group_metadata_proto_rawDesc
)

func file_mls_message_contents_group_metadata_proto_rawDescGZIP() []byte {
	file_mls_message_contents_group_metadata_proto_rawDescOnce.Do(func() {
		file_mls_message_contents_group_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_mls_message_contents_group_metadata_proto_rawDescData)
	})
	return file_mls_message_contents_group_metadata_proto_rawDescData
}

var file_mls_message_contents_group_metadata_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_mls_message_contents_group_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_mls_message_contents_group_metadata_proto_goTypes = []interface{}{
	(ConversationType)(0),                 // 0: xmtp.mls.message_contents.ConversationType
	(MembershipPolicy_BasePolicy)(0),      // 1: xmtp.mls.message_contents.MembershipPolicy.BasePolicy
	(*GroupMetadataV1)(nil),               // 2: xmtp.mls.message_contents.GroupMetadataV1
	(*PolicySet)(nil),                     // 3: xmtp.mls.message_contents.PolicySet
	(*MembershipPolicy)(nil),              // 4: xmtp.mls.message_contents.MembershipPolicy
	(*MembershipPolicy_AndCondition)(nil), // 5: xmtp.mls.message_contents.MembershipPolicy.AndCondition
	(*MembershipPolicy_AnyCondition)(nil), // 6: xmtp.mls.message_contents.MembershipPolicy.AnyCondition
}
var file_mls_message_contents_group_metadata_proto_depIdxs = []int32{
	0, // 0: xmtp.mls.message_contents.GroupMetadataV1.conversation_type:type_name -> xmtp.mls.message_contents.ConversationType
	3, // 1: xmtp.mls.message_contents.GroupMetadataV1.policies:type_name -> xmtp.mls.message_contents.PolicySet
	4, // 2: xmtp.mls.message_contents.PolicySet.add_member_policy:type_name -> xmtp.mls.message_contents.MembershipPolicy
	4, // 3: xmtp.mls.message_contents.PolicySet.remove_member_policy:type_name -> xmtp.mls.message_contents.MembershipPolicy
	1, // 4: xmtp.mls.message_contents.MembershipPolicy.base:type_name -> xmtp.mls.message_contents.MembershipPolicy.BasePolicy
	5, // 5: xmtp.mls.message_contents.MembershipPolicy.and_condition:type_name -> xmtp.mls.message_contents.MembershipPolicy.AndCondition
	6, // 6: xmtp.mls.message_contents.MembershipPolicy.any_condition:type_name -> xmtp.mls.message_contents.MembershipPolicy.AnyCondition
	4, // 7: xmtp.mls.message_contents.MembershipPolicy.AndCondition.policies:type_name -> xmtp.mls.message_contents.MembershipPolicy
	4, // 8: xmtp.mls.message_contents.MembershipPolicy.AnyCondition.policies:type_name -> xmtp.mls.message_contents.MembershipPolicy
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_mls_message_contents_group_metadata_proto_init() }
func file_mls_message_contents_group_metadata_proto_init() {
	if File_mls_message_contents_group_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mls_message_contents_group_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GroupMetadataV1); i {
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
		file_mls_message_contents_group_metadata_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PolicySet); i {
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
		file_mls_message_contents_group_metadata_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MembershipPolicy); i {
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
		file_mls_message_contents_group_metadata_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MembershipPolicy_AndCondition); i {
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
		file_mls_message_contents_group_metadata_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MembershipPolicy_AnyCondition); i {
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
	file_mls_message_contents_group_metadata_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*MembershipPolicy_Base)(nil),
		(*MembershipPolicy_AndCondition_)(nil),
		(*MembershipPolicy_AnyCondition_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mls_message_contents_group_metadata_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mls_message_contents_group_metadata_proto_goTypes,
		DependencyIndexes: file_mls_message_contents_group_metadata_proto_depIdxs,
		EnumInfos:         file_mls_message_contents_group_metadata_proto_enumTypes,
		MessageInfos:      file_mls_message_contents_group_metadata_proto_msgTypes,
	}.Build()
	File_mls_message_contents_group_metadata_proto = out.File
	file_mls_message_contents_group_metadata_proto_rawDesc = nil
	file_mls_message_contents_group_metadata_proto_goTypes = nil
	file_mls_message_contents_group_metadata_proto_depIdxs = nil
}
