package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// subcommandCmd represents the subcommand command
var subcommandCmd = &cobra.Command{
	Use:   "set-param",
	Short: "Prompt the user to set one ore parameters for their runbook",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		fields := ParamFields(ctx, params)
		var fs []huh.Field
		for _, param := range params {
			fs = append(fs, fields[param])
		}

		if len(fs) == 0 {
			return
		}

		param := huh.NewGroup(fs...).Title(title)

		if err := huh.NewForm(param).Run(); err != nil {
			// TODO: handle error
			panic(err)
		}

		for _, f := range fs {
			i, ok := f.(*huh.Input)
			if !ok {
				continue
			}
			fmt.Println(i.GetKey(), i.GetValue())
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
