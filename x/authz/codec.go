package authz

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/authz interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	codec.RegisterAminoMsg(cdc, &MsgGrant{}, "cosmos-sdk/MsgGrant")
	codec.RegisterAminoMsg(cdc, &MsgRevoke{}, "cosmos-sdk/MsgRevoke")
	codec.RegisterAminoMsg(cdc, &MsgExec{}, "cosmos-sdk/MsgExec")

	cdc.RegisterInterface((*Authorization)(nil), nil)
	cdc.RegisterConcrete(&GenericAuthorization{}, "cosmos-sdk/GenericAuthorization", nil)
}

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgGrant{},
		&MsgRevoke{},
		&MsgExec{},
	)

	registry.RegisterInterface(
		"cosmos.v1beta1.Authorization",
		(*Authorization)(nil),
		&GenericAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, MsgServiceDesc())
}
func init() {
	// Register all Amino interfaces and concrete types on the global Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(legacy.Cdc)
}
