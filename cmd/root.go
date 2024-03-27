package cmd

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"golang.org/x/term"
)

var noConfig bool

const (
	VERSION           string = "0.10.0"
	VIPER_CONFIG_NAME string = ".vos-cli"
	VIPER_CONFIG_PATH string = "$HOME"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "vos-cli",
	Version: VERSION,
	Short:   "A utility for working with Cisco IOS devices",
	Long:    "",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&noConfig, "no-config", false, "bypass the configuration file")

	// Bypass the config file if the --no-config flag is set
	if slices.Contains(os.Args, "--no-config") {
		return
	}

	viper.SetConfigName(VIPER_CONFIG_NAME)
	viper.SetConfigType("yml")
	viper.AddConfigPath(VIPER_CONFIG_PATH)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			initalizeConfig()
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}

func initalizeConfig() {
	fmt.Printf("Initializing configuration file...\n\n")

	// Initialize the default config
	var baseConfig = `
user: ""
password: ""
`
	err := viper.ReadConfig(bytes.NewBuffer([]byte(baseConfig)))
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// Get the default user from stdin
	var user string
	fmt.Fprint(os.Stderr, "Default Cisco User: ")
	fmt.Scanln(&user)
	// Add to the config
	viper.Set("user", user)

	// Get the default password from stdin
	fmt.Fprintf(os.Stderr, "Default Password (%s): ", user)
	bytepw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	// Add to the config
	viper.Set("password", string(bytepw))

	// Save the new config file
	err = viper.SafeWriteConfig()
	if err != nil {
		panic(fmt.Errorf("unable to write config file, %v", err))
	}

	// Read in the new config file
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// Set the permissions on the config file
	os.Chmod(viper.ConfigFileUsed(), 0600)

	fmt.Printf("\n\nInitialization complete.\n\n")
	fmt.Printf("Configuration file saved to %v.\n\n", viper.ConfigFileUsed())

	os.Exit(0)
}
