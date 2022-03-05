package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.UpdateTallyOfVPEndProposals(ctx)
}