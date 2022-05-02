package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Version struct {
	ClientVersion string `json:"clientVersion"`
}

func newCmdVersion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunVersion(out, cmd)
		},
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringP("output", "o", "", "Output format; available options are 'yaml', 'json' and 'short'")
	return cmd
}

// RunVersion provides the version information of kubeadm in format depending on arguments
// specified in cobra.Command.
func RunVersion(out io.Writer, cmd *cobra.Command) error {
	v := Version{
		ClientVersion: "1.5.0",
	}
	const flag = "output"
	of, err := cmd.Flags().GetString(flag)
	if err != nil {
		return errors.Wrapf(err, "error accessing flag %s for command %s", flag, cmd.Name())
	}
	switch of {
	case "":
		fmt.Fprintf(out, "cloud version: %#v\n", v.ClientVersion)
	case "short":
		fmt.Fprintf(out, "%s\n", v.ClientVersion)
	case "yaml":
		y, err := yaml.Marshal(&v)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(y))
	case "json":
		y, err := json.MarshalIndent(&v, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(y))
	default:
		return errors.Errorf("invalid output format: %s", of)
	}
	return nil
}
