package args

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
)

func init() {
	cobra.MousetrapHelpText = ""
}

var rootCommand = cobra.Command{
	Use:   "cvdsmdu",
	Short: "CVDS MDU Application",
	Long:  "CVDS Media Distribute Unit Application",
	Run: func(cmd *cobra.Command, args []string) {
		argsConfig.Help = false
		if len(args) > 0 {
			argsConfig.Command = args[0]
		}
	},
}

type ArgsConfig struct {
	Help       bool
	ConfigFile string
	Command    string
}

func (c *ArgsConfig) PreHandle() {
	c.Help = true
}

var argsConfig *ArgsConfig
var argsConfigInit sync.Once

func initArgsConfig() {
	argsConfig = new(ArgsConfig)
	argsConfig.PreHandle()
	rootCommand.Flags().StringVarP(&argsConfig.ConfigFile, "config", "f", "", "specify config file")
	err := rootCommand.Execute()
	if err != nil {
		log.Fatalln(err)
	}
	if argsConfig.Help {
		os.Exit(0)
	}
}

func GetArgsConfig() *ArgsConfig {
	if argsConfig != nil {
		return argsConfig
	}
	argsConfigInit.Do(initArgsConfig)
	return argsConfig
}
