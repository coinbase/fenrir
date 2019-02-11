package main

import (
	"fmt"
	"os"

	"github.com/coinbase/fenrir/client"
	"github.com/coinbase/fenrir/deployer"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/run"
	"github.com/coinbase/step/utils/to"
)

func main() {
	var arg, command string
	switch len(os.Args) {
	case 1:
		fmt.Println("Starting Lambda")
		run.LambdaTasks(deployer.TaskHandlers())
	case 2:
		command = os.Args[1]
		arg = ""
	case 3:
		command = os.Args[1]
		arg = os.Args[2]
	default:
		printUsage() // Print how to use and exit
	}

	step_fn := to.Strp("coinbase-fenrir")

	switch command {
	case "json":
		// This is required to use the step to deploy
		run.JSON(deployer.StateMachine())
	case "deploy":
		releaseFile := &arg
		if is.EmptyStr(releaseFile) {
			releaseFile = to.Strp("./template.yml")
		}

		err := client.Deploy(step_fn, releaseFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	case "package":
		releaseFile := &arg
		if is.EmptyStr(releaseFile) {
			releaseFile = to.Strp("./template.yml")
		}

		err := client.Package(releaseFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	default:
		printUsage() // Print how to use and exit
	}
}

func printUsage() {
	fmt.Println("Usage: fenrir json|deploy|package <release_file> (No args starts Lambda)")
	os.Exit(0)
}
