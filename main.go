package main

import (
	"github.com/Dapacruz/vos-cli/cmd"
	_ "github.com/Dapacruz/vos-cli/cmd/config"
	_ "github.com/Dapacruz/vos-cli/cmd/device"
)

func main() {
	cmd.Execute()
}
