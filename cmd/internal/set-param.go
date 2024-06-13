package internal

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/param"
	"github.com/getsavvyinc/savvy-cli/server/run"
	"github.com/getsavvyinc/savvy-cli/slice"
	"github.com/spf13/cobra"
)

// subcommandCmd represents the subcommand command
var subcommandCmd = &cobra.Command{
	Use:   "set-param",
	Short: "Prompt the user to set one ore parameters for their runbook",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cl, err := run.NewDefaultClient(ctx)
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		state, err := cl.CurrentState()
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		command := state.Command
		params := param.Extract(command)
		// Exit early
		if len(params) == 0 {
			return
		}

		unsetParams := slice.Filter(params, func(p string) bool {
			_, ok := state.Params[p]
			return ok
		})

		if len(unsetParams) == 0 {
			return
		}

		fields := ParamFields(ctx, unsetParams)

		var fs []huh.Field
		for _, param := range params {
			fs = append(fs, fields[param])
		}

		if len(fs) == 0 {
			return
		}

		param := huh.NewGroup(fs...).Title(title)

		if err := huh.NewForm(param).Run(); err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		newParams := map[string]string{}
		for _, f := range fs {
			i, ok := f.(*huh.Input)
			if !ok {
				continue
			}
			strVal, ok := i.GetValue().(string)
			if !ok {
				continue
			}
			newParams[i.GetKey()] = strVal
		}

		if err := cl.SetParams(newParams); err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}
	},
}

var title string
var params []string

func init() {
	InternalCmd.AddCommand(subcommandCmd)

	// title flag
	subcommandCmd.Flags().StringVarP(&title, "title", "t", "Set Params", "form title")
	// params flag. This is a slice of strings
	subcommandCmd.Flags().StringSliceVarP(&params, "params", "p", []string{}, "form params")
}

func ParamFields(ctx context.Context, params []string) map[string]huh.Field {
	fields := map[string]huh.Field{}

	for _, param := range params {
		if _, ok := fields[param]; ok {
			continue
		}
		title, desc := parseParam(param)
		fields[param] = huh.NewInput().Title(title).Description(desc).Key(param)
	}
	return fields
}

const DefaultTitle = "Set Params"
const DefaultDescription = ""

func parseParam(param string) (string, string) {
	if !strings.HasPrefix(param, "<") || !strings.HasSuffix(param, ">") {
		return DefaultTitle, DefaultDescription
	}

	title := "Set " + param

	return title, ""
}
