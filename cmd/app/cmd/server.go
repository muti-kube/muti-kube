package cmd

import (
	"muti-kube/cmd/app/config"
	"muti-kube/pkg/periodic"
	"muti-kube/pkg/util/logger"
	"muti-kube/router"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

func serverPreRun() {
	config.LoadConfigFile(configFile)
}

func newCmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Start muti-kube API Server",
		Example: "muti-kube config/config.yml",
		PreRun: func(cmd *cobra.Command, args []string) {
			serverPreRun()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config/config.yml", "Start muti-kube with provided configuration file")
	return cmd
}

func run() error {
	r := router.InitRouter()
	gin.SetMode(gin.DebugMode)
	go func() {
		if err := r.Run(":9000"); err != nil {
			logger.Error()
		}
	}()
	go func() {
		runPeriodic()
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	return nil
}

func runPeriodic() {
	ticketPeriodic, err := periodic.NewTicketPeriodic()
	if err != nil {
		logger.Warn(err)
	}
	ticketPeriodic.Start()
}
