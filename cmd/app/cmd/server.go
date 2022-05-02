package cmd

import (
	"fmt"
	"muti-kube/router"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func newCmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "muti-kube server",
		Short:   "Start API muti-kube",
		Example: "muti-kube config/settings.yml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output format; available options are 'yaml', 'json' and 'short'")
	return cmd
}

func run() error {
	r := router.InitRouter()
	gin.SetMode(gin.DebugMode)
	go func() {
		if err := r.Run(":9000"); err != nil {
			fmt.Println(err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	return nil
}
