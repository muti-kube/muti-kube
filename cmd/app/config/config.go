package config

import (
	"fmt"
	"io/ioutil"
	"muti-kube/pkg/util/logger"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfigFile(path string) {
	viper.SetConfigFile(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Read config file fail: %s", err.Error()))
	}
	//Replace environment variables
	err = viper.ReadConfig(strings.NewReader(os.ExpandEnv(string(content))))
	if err != nil {
		logger.Fatal(fmt.Sprintf("Parse config file fail: %s", err.Error()))
	}

	// init log plugin
	logger.Init()
}
