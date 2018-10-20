package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

//const dockerExe = "/usr/local/bin/docker"
const dockerExe = "docker"
const dockerdExe = "/usr/local/bin/dockerd"

func commandLogin(w io.Writer, d DockerRegistry) error {
	cmd := exec.Command(
		"docker", "login",
		"-u", d.Username,
		"-p", d.Password,
		d.Registry,
	)

	cmd.Stdout = w
	cmd.Stderr = w

	return cmd.Run()
}

func commandVersion() *exec.Cmd {
	return exec.Command(dockerExe, "version")
}

func commandInfo() *exec.Cmd {
	return exec.Command(dockerExe, "info")
}

func commandDockerDaemon(c Config) *exec.Cmd {
	cmd := exec.Command(dockerdExe)
	if c.Debug {

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	} else {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	}

	return cmd
}

func commandPullSource(config Config) *exec.Cmd {
	repository := fmt.Sprintf("%s:%s", config.SourceRepository, config.SourceTag)

	args := []string{"pull", repository}
	return exec.Command(dockerExe, args...)
}

func commandAddTag(config Config, tag string) *exec.Cmd {
	sourceRepository := fmt.Sprintf("%s:%s", config.SourceRepository, config.SourceTag)
	destinationRepository := fmt.Sprintf("%s:%s", config.DestinationRepository, tag)

	args := []string{"tag", sourceRepository, destinationRepository}
	return exec.Command(dockerExe, args...)

}

func commandPushTag(config Config, tag string) *exec.Cmd {
	destinationRepository := fmt.Sprintf("%s:%s", config.DestinationRepository, tag)

	args := []string{"push", destinationRepository}
	return exec.Command(dockerExe, args...)
}
