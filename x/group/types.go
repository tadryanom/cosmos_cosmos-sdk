package group

import (
	"fmt"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MaxMetadataLength defines the max length of the metadata bytes field
// for various entities within the group module
// TODO: This could be used as params once x/params is upgraded to use protobuf
const MaxMetadataLength = 255

type DecisionPolicyResult struct {
	Allow bool
	Final bool
}

// DecisionPolicy is the persistent set of rules to determine the result of election on a proposal.
type DecisionPolicy interface {
	codec.ProtoMarshaler

	ValidateBasic() error
	GetTimeout() types.Duration
	Allow(tally Tally, totalPower string, votingDuration time.Duration) (DecisionPolicyResult, error)
	Validate(g GroupInfo) error
}

// NewGroupAccountInfo creates a new GroupAccountInfo instance
func NewGroupAccountInfo(address sdk.AccAddress, group uint64, admin sdk.AccAddress, metadata []byte,
	version uint64, decisionPolicy DecisionPolicy, derivationKey []byte) (GroupAccountInfo, error) {
	p := GroupAccountInfo{
		Address:       address.String(),
		GroupId:       group,
		Admin:         admin.String(),
		Metadata:      metadata,
		Version:       version,
		DerivationKey: derivationKey,
	}

	err := p.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return GroupAccountInfo{}, err
	}

	return p, nil
}

func (g *GroupAccountInfo) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	g.DecisionPolicy = any
	return nil
}

func (g GroupAccountInfo) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := g.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g GroupAccountInfo) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(g.DecisionPolicy, &decisionPolicy)
}

func (g GroupAccountInfo) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(g.Address)
	if err != nil {
		panic(err)
	}
	return []interface{}{addr.Bytes()}
}

func (g GroupMember) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(g.Member.Address)
	if err != nil {
		panic(err)
	}
	return []interface{}{g.GroupId, addr.Bytes()}
}

// func (g GroupMember) ValidateBasic() error {
// 	if g.GroupId == 0 {
// 		return sdkerrors.Wrap(ErrEmpty, "group")
// 	}

// 	err := g.Member.ValidateBasic()
// 	if err != nil {
// 		return sdkerrors.Wrap(err, "member")
// 	}
// 	return nil
// }

func (v Vote) PrimaryKeyFields() []interface{} {
	addr, err := sdk.AccAddressFromBech32(v.Voter)
	if err != nil {
		panic(err)
	}
	return []interface{}{v.ProposalId, addr.Bytes()}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (q QueryGroupAccountsByGroupResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpackGroupAccounts(unpacker, q.GroupAccounts)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (q QueryGroupAccountsByAdminResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpackGroupAccounts(unpacker, q.GroupAccounts)
}

func unpackGroupAccounts(unpacker codectypes.AnyUnpacker, accs []*GroupAccountInfo) error {
	for _, g := range accs {
		err := g.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}

	return nil
}
