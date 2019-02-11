package client

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/coinbase/step/utils/to"
)

func Package(releaseFile *string) error {
	release, err := releaseFromFile(releaseFile, to.Strp("region"), to.Strp("account"))
	if err != nil {
		return err
	}

	// fmt.Println("Executing pwd")
	// cmd := exec.Command("pwd")
	// fmt.Println(cmd.Args)
	// if err := cmd.Run(); err != nil {
	// 	return err
	// }
	// fmt.Println(cmd.Dir)
	// fmt.Println(cmd.Stdout)
	// fmt.Println("pwd")

	// docker build . -t release.ProjectName
	buildTag := strings.ToLower(to.RandomString(8))
	err = execute("docker", "build", "-t", buildTag, ".")
	if err != nil {
		return err
	}

	containerName := strings.ToLower(to.RandomString(8))
	err = execute("docker", "create", "-it", "--name", containerName, buildTag, "bash")
	if err != nil {
		return err
	}

	defer execute("docker", "rm", containerName)

	for name, _ := range release.Template.GetAllAWSServerlessFunctionResources() {
		err = execute(
			"docker",
			"cp",
			fmt.Sprintf("%v:%v", containerName, fmt.Sprintf("%v.zip", name)),
			fmt.Sprintf("%v.%v.zip", *releaseFile, name),
		)
		if err != nil {
			return err
		}
	}

	fmt.Println("Complete")

	return nil
}

func execute(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	fmt.Println(name, args)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println(out.String())
	return nil
}
