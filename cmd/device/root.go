package device

import (
	"os"
	"sync"

	"github.com/Dapacruz/vos-cli/cmd"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	user     string
	password string
	hosts    []string
	wg       sync.WaitGroup
)

// Create objects to colorize stdout
var (
	blue   *color.Color = color.New(color.FgBlue)
	green  *color.Color = color.New(color.FgGreen)
	yellow *color.Color = color.New(color.FgHiYellow)
)

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "A set of commands for working with Cisco IOS devices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(deviceCmd)
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
