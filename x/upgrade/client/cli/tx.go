package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	FlagUpgradeHeight    = "upgrade-height"
	FlagUpgradeInfo      = "upgrade-info"
	FlagNoValidate       = "no-validate"
	FlagNoValidateInfoCS = "no-validate-info-checksum"
	FlagNoValidateBinCS  = "no validate-binary-checksums"
	FlagNoValidateCS     = "no-validate-checksums"
	FlagDaemonName       = "daemon-name"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Upgrade transaction subcommands",
	}

	return cmd
}

// NewCmdSubmitUpgradeProposal implements a command handler for submitting a software upgrade proposal transaction.
func NewCmdSubmitUpgradeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "software-upgrade [name] (--upgrade-height [height]) (--upgrade-info [info]) [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a software upgrade proposal",
		Long: "Submit a software upgrade along with an initial deposit.\n" +
			"Please specify a unique name and height for the upgrade to take effect.\n" +
			"You may include info to reference a binary download link, in a format compatible with: https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			name := args[0]
			content, err := parseArgsToContent(cmd, name)
			if err != nil {
				return err
			}
			noValidate, err := cmd.Flags().GetBool(FlagNoValidate)
			if err != nil {
				return err
			}
			if !noValidate {
				prop := content.(*types.SoftwareUpgradeProposal)
				var daemonName string
				var noValCS, noValInfoCS, noValBinCS bool
				if daemonName, err = cmd.Flags().GetString(FlagDaemonName); err != nil {
					return err
				}
				if noValCS, err = cmd.Flags().GetBool(FlagNoValidateCS); err != nil {
					return err
				}
				if noValInfoCS, err = cmd.Flags().GetBool(FlagNoValidateInfoCS); err != nil {
					return err
				}
				if noValBinCS, err = cmd.Flags().GetBool(FlagNoValidateBinCS); err != nil {
					return err
				}
				var planInfo *plan.Info
				if planInfo, err = plan.ParseInfo(prop.Plan.Info, !noValCS && !noValInfoCS); err != nil {
					return err
				}
				if err = planInfo.ValidateFull(daemonName, !noValCS && !noValBinCS); err != nil {
					return err
				}
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen")
	cmd.Flags().String(FlagUpgradeInfo, "", "Info for the upgrade plan such as new version download urls, etc.")
	cmd.Flags().Bool(FlagNoValidate, false, "Skip validation of the upgrade info")
	cmd.Flags().Bool(FlagNoValidateInfoCS, false, "Do not require a url in the info string to have a checksum")
	cmd.Flags().Bool(FlagNoValidateBinCS, false, "Do not require urls to binaries to have a checksum")
	cmd.Flags().Bool(FlagNoValidateCS, false, fmt.Sprintf("Do not require any urls to have a checksum. same as --%s --%s", FlagNoValidateInfoCS, FlagNoValidateBinCS))
	cmd.Flags().String(FlagDaemonName, getDefaultDaemonName(), "The name of the executable being upgraded (for upgrade-info validation). Default is the DAEMON_NAME env var if set, or else this executable")

	return cmd
}

// NewCmdSubmitCancelUpgradeProposal implements a command handler for submitting a software upgrade cancel proposal transaction.
func NewCmdSubmitCancelUpgradeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-software-upgrade [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Cancel the current software upgrade proposal",
		Long:  "Cancel a software upgrade along with an initial deposit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			content := types.NewCancelSoftwareUpgradeProposal(title, description)

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}

func parseArgsToContent(cmd *cobra.Command, name string) (gov.Content, error) {
	title, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(cli.FlagDescription)
	if err != nil {
		return nil, err
	}

	height, err := cmd.Flags().GetInt64(FlagUpgradeHeight)
	if err != nil {
		return nil, err
	}

	info, err := cmd.Flags().GetString(FlagUpgradeInfo)
	if err != nil {
		return nil, err
	}

	plan := types.Plan{Name: name, Height: height, Info: info}
	content := types.NewSoftwareUpgradeProposal(title, description, plan)
	return content, nil
}

// getDefaultDaemonName gets the default name to use for the daemon.
// If a DAEMON_NAME env var is set, that is used.
// Otherwise, the last part of the currently running executable is used.
func getDefaultDaemonName() string {
	// DAEMON_NAME is specifically used here to correspond with the Comsovisor setup env vars.
	name := os.Getenv("DAEMON_NAME")
	if len(name) == 0 {
		_, name = filepath.Split(os.Args[0])
	}
	return name
}
