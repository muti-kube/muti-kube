package cmd

import (
	"io"
	cloudutil "muti-kube/cmd/app/cmd/util"
	"muti-kube/pkg/util/logger"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// NewCloudCommand returns cobra.Command to run muti-kube command
func NewCloudCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	var rootfsPath string
	cmds := &cobra.Command{
		Use:   "muti-kube",
		Short: "muti-kube is a multi-functional cloud native management system",
		Long: Dedent(`
			    ┌──────────────────────────────────────────────────────────┐
			    │           This is kube cloud description                 │
				│1. Provides the cluster management function		       │
				│2. Provides the cluster node monitoring function	       │
			    │                                                          │
			    └──────────────────────────────────────────────────────────┘
		`),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if rootfsPath != "" {
				if err := cloudutil.Chroot(rootfsPath); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmds.AddCommand(newCmdVersion(out))
	cmds.AddCommand(newCmdServer())
	// init log plugin
	logger.Init()
	return cmds
}

var (
	whitespaceOnly    = regexp.MustCompile("(?m)^[ \t]+$")
	leadingWhitespace = regexp.MustCompile("(?m)(^[ \t]*)(?:[^ \t\n])")
)

func Dedent(text string) string {
	var margin string
	text = whitespaceOnly.ReplaceAllString(text, "")
	indents := leadingWhitespace.FindAllStringSubmatch(text, -1)
	for i, indent := range indents {
		if i == 0 {
			margin = indent[1]
		} else if strings.HasPrefix(indent[1], margin) {
			continue
		} else if strings.HasPrefix(margin, indent[1]) {
			margin = indent[1]
		} else {
			margin = ""
			break
		}
	}
	if margin != "" {
		text = regexp.MustCompile("(?m)^"+margin).ReplaceAllString(text, "")
	}
	return text
}
