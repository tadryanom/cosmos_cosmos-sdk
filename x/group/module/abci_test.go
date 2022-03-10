package module_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestPruning(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(30000000))
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr3 := addrs[2]

	// Initial group, group policy and balance setup
	members := []group.Member{
		{Address: addr1.String(), Weight: "1"}, {Address: addr2.String(), Weight: "2"},
	}

	groupRes, err := app.GroupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addr1.String(),
		Members: members,
	})

	require.NoError(t, err)
	groupID := groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		time.Second,
		0,
	)

	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   addr1.String(),
		GroupId: groupID,
	}

	err = policyReq.SetDecisionPolicy(policy)
	require.NoError(t, err)
	policyRes, err := app.GroupKeeper.CreateGroupPolicy(ctx, policyReq)
	require.NoError(t, err)

	groupPolicyAddr, err := sdk.AccAddressFromBech32(policyRes.Address)
	require.NoError(t, err)
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	msgSend1 := &banktypes.MsgSend{
		FromAddress: groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}
	proposers := []string{addr2.String()}

	specs := map[string]struct {
		srcBlockTime      time.Time
		setupProposal     func(ctx context.Context) uint64
		expErr            bool
		expErrMsg         string
		expExecutorResult group.ProposalExecutorResult
	}{
		"proposal pruned after executor result success": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				pID, err := submitProposalAndVote(ctx, app, msgs, proposers, group.VOTE_OPTION_YES, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Signer: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expErrMsg:         "load proposal: not found",
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_SUCCESS,
		},
		"proposal with multiple messages pruned when executed with result success": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1, msgSend1}
				pID, err := submitProposalAndVote(ctx, app, msgs, proposers, group.VOTE_OPTION_YES, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Signer: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expErrMsg:         "load proposal: not found",
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_SUCCESS,
		},
		"proposal not pruned when not executed and rejected": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				pID, err := submitProposalAndVote(ctx, app, msgs, proposers, group.VOTE_OPTION_NO, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Signer: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		},
		"open proposal is not pruned which must not fail ": {
			setupProposal: func(ctx context.Context) uint64 {
				pID, err := submitProposal(ctx, app, []sdk.Msg{msgSend1}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Signer: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		},
		"proposal not pruned with group policy modified before tally": {
			setupProposal: func(ctx context.Context) uint64 {
				pID, err := submitProposal(ctx, app, []sdk.Msg{msgSend1}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.UpdateGroupPolicyMetadata(ctx, &group.MsgUpdateGroupPolicyMetadata{
					Admin:   addr1.String(),
					Address: groupPolicyAddr.String(),
				})
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Signer: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		},
		"pruned when proposal is executable when failed before": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				pID, err := submitProposalAndVote(ctx, app, msgs, proposers, group.VOTE_OPTION_YES, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Signer: addrs[2].String(), ProposalId: pID})
				require.NoError(t, err)
				return pID
			},
			expErrMsg:         "load proposal: not found",
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_SUCCESS,
		},
	}
	for msg, spec := range specs {
		spec := spec
		t.Run(msg, func(t *testing.T) {
			proposalID := spec.setupProposal(ctx)

			module.EndBlocker(ctx, app.GroupKeeper)

			if spec.expExecutorResult == group.PROPOSAL_EXECUTOR_RESULT_SUCCESS {
				// Make sure proposal is deleted from state
				_, err = app.GroupKeeper.Proposal(ctx, &group.QueryProposalRequest{ProposalId: proposalID})
				require.Contains(t, err.Error(), spec.expErrMsg)
				res, err := app.GroupKeeper.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{ProposalId: proposalID})
				require.NoError(t, err)
				require.Empty(t, res.GetVotes())
			} else {
				// Check that proposal and votes exists
				res, err := app.GroupKeeper.Proposal(ctx, &group.QueryProposalRequest{ProposalId: proposalID})
				require.NoError(t, err)
				_, err = app.GroupKeeper.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{ProposalId: res.Proposal.Id})
				require.NoError(t, err)
				require.Equal(t, "", spec.expErrMsg)

				exp := group.ProposalExecutorResult_name[int32(spec.expExecutorResult)]
				got := group.ProposalExecutorResult_name[int32(res.Proposal.ExecutorResult)]
				assert.Equal(t, exp, got)
			}
		})
	}

}

func submitProposalAndVote(
	ctx context.Context, app *simapp.SimApp, msgs []sdk.Msg,
	proposers []string, voteOption group.VoteOption, groupPolicyAddr sdk.AccAddress) (uint64, error) {
	myProposalID, err := submitProposal(ctx, app, msgs, proposers, groupPolicyAddr)
	if err != nil {
		return 0, err
	}
	_, err = app.GroupKeeper.Vote(ctx, &group.MsgVote{
		ProposalId: myProposalID,
		Voter:      proposers[0],
		Option:     voteOption,
	})
	if err != nil {
		return 0, err
	}
	return myProposalID, nil
}

func submitProposal(
	ctx context.Context, app *simapp.SimApp, msgs []sdk.Msg,
	proposers []string, groupPolicyAddr sdk.AccAddress) (uint64, error) {
	proposalReq := &group.MsgSubmitProposal{
		Address:   groupPolicyAddr.String(),
		Proposers: proposers,
	}
	err := proposalReq.SetMsgs(msgs)
	if err != nil {
		return 0, err
	}

	proposalRes, err := app.GroupKeeper.SubmitProposal(ctx, proposalReq)
	if err != nil {
		return 0, err
	}

	return proposalRes.ProposalId, nil
}