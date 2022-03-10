package group

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary group module concrete
// types and interfaces with the provided codec reference.
// These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*DecisionPolicy)(nil), nil)
	cdc.RegisterConcrete(&ThresholdDecisionPolicy{}, "cosmos-sdk/ThresholdDecisionPolicy", nil)
	cdc.RegisterConcrete(&PercentageDecisionPolicy{}, "cosmos-sdk/PercentageDecisionPolicy", nil)

	codec.RegisterAminoMsg(cdc, &MsgCreateGroup{}, "cosmos-sdk/MsgCreateGroup")
	codec.RegisterAminoMsg(cdc, &MsgUpdateGroupMembers{}, "cosmos-sdk/MsgUpdateGroupMembers")
	codec.RegisterAminoMsg(cdc, &MsgUpdateGroupAdmin{}, "cosmos-sdk/MsgUpdateGroupAdmin")
	codec.RegisterAminoMsg(cdc, &MsgUpdateGroupMetadata{}, "cosmos-sdk/MsgUpdateGroupMetadata")
	codec.RegisterAminoMsg(cdc, &MsgCreateGroupWithPolicy{}, "cosmos-sdk/MsgCreateGroupWithPolicy")
	codec.RegisterAminoMsg(cdc, &MsgCreateGroupPolicy{}, "cosmos-sdk/MsgCreateGroupPolicy")
	codec.RegisterAminoMsg(cdc, &MsgUpdateGroupPolicyAdmin{}, "cosmos-sdk/MsgUpdateGroupPolicyAdmin")
	codec.RegisterAminoMsg(cdc, &MsgUpdateGroupPolicyDecisionPolicy{}, "cosmos-sdk/MsgUpdateGroupPolicyDecisionPolicy")
	codec.RegisterAminoMsg(cdc, &MsgUpdateGroupPolicyMetadata{}, "cosmos-sdk/MsgUpdateGroupPolicyMetadata")
	codec.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/group/MsgSubmitProposal")
	codec.RegisterAminoMsg(cdc, &MsgWithdrawProposal{}, "cosmos-sdk/group/MsgWithdrawProposal")
	codec.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/group/MsgVote")
	codec.RegisterAminoMsg(cdc, &MsgExec{}, "cosmos-sdk/group/MsgExec")
	codec.RegisterAminoMsg(cdc, &MsgLeaveGroup{}, "cosmos-sdk/group/MsgLeaveGroup")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGroup{},
		&MsgUpdateGroupMembers{},
		&MsgUpdateGroupAdmin{},
		&MsgUpdateGroupMetadata{},
		&MsgCreateGroupWithPolicy{},
		&MsgCreateGroupPolicy{},
		&MsgUpdateGroupPolicyAdmin{},
		&MsgUpdateGroupPolicyDecisionPolicy{},
		&MsgUpdateGroupPolicyMetadata{},
		&MsgSubmitProposal{},
		&MsgWithdrawProposal{},
		&MsgVote{},
		&MsgExec{},
		&MsgLeaveGroup{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	registry.RegisterInterface(
		"cosmos.group.v1beta1.DecisionPolicy",
		(*DecisionPolicy)(nil),
		&ThresholdDecisionPolicy{},
		&PercentageDecisionPolicy{},
	)
}

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}
