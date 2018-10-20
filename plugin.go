package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type (
	Repo struct {
		Owner   string
		Name    string
		Link    string
		Avatar  string
		Branch  string
		Private bool
		Trusted bool
	}

	Build struct {
		Number   int
		Event    string
		Status   string
		Deploy   string
		Created  int64
		Started  int64
		Finished int64
		Link     string
	}

	Commit struct {
		Remote  string
		Sha     string
		Ref     string
		Link    string
		Branch  string
		Message string
		Author  Author
	}

	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	Config struct {
		Debug                 bool
		SourceRepository      string
		SourceTag             string
		DestinationRepository string
		DestinationTags       string
		DestinationTagList    []string
		SourceRegistry        DockerRegistry
		DestinationRegistry   DockerRegistry
	}

	DockerRegistry struct {
		Username string
		Password string
		Registry string
	}

	Plugin struct {
		Repo   Repo
		Build  Build
		Commit Commit
		Config Config
	}
)

func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}

func missingField(field string) string {
	return fmt.Sprintf("'%s' needs to be provided", field)
}

func (c *Config) ValidateAndUpdate() error {
	if c.SourceRegistry.Username == "" || c.SourceRegistry.Password == "" {
		return errors.New("'docker_username' and 'docker_password' secrets needs to be provided")
	}

	if c.SourceRepository == "" {
		return errors.New(missingField("repository"))
	}

	if c.SourceTag == "" {
		return errors.New(missingField("tag"))
	}

	if c.DestinationTags == "" {
		return errors.New(missingField("destination_tags"))
	}
	c.DestinationTagList = strings.Split(c.DestinationTags, ",")

	if c.DestinationRepository == "" {
		fmt.Printf("Destination image repository was not provided, assuming: '%s'\n", c.SourceRepository)
		c.DestinationRepository = c.SourceRepository
	}

	if c.DestinationRegistry.Registry == "" && c.SourceRegistry.Registry != "" {
		fmt.Printf("Destination registry was not provided, assuming: '%s'\n", c.SourceRegistry.Registry)
		c.DestinationRegistry.Registry = c.SourceRegistry.Registry
	}

	if c.SourceRegistry.Registry != c.DestinationRegistry.Registry {
		if c.DestinationRegistry.Username == "" || c.DestinationRegistry.Password == "" {
			return errors.New("'destination_docker_username' and 'destination_docker_password' secrets needs to be" +
				" provided")
		}
	}

	return nil
}

func runCommands(cmds []*exec.Cmd) error {
	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
func (p Plugin) Exec() error {
	err := p.Config.ValidateAndUpdate()
	if err != nil {
		return err
	}

	//Start docker daemon
	cmd := commandDockerDaemon(p.Config)
	go func() {
		cmd.Run()
	}()

	for i := 0; ; i++ {
		fmt.Println("Waiting for docker daemon...")
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}

		time.Sleep(time.Second * 1)
		if i >= 15 {
			fmt.Println("Docker daemon is not running")
			os.Exit(-1)
		}
	}
	err = commandLogin(os.Stdout, p.Config.SourceRegistry)
	if err != nil {
		return fmt.Errorf("Error authenticating: %s\n", err)

	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())
	cmds = append(cmds, commandInfo())
	cmds = append(cmds, commandPullSource(p.Config))

	err = runCommands(cmds)
	if err != nil {
		return err
	}

	if p.Config.SourceRegistry.Registry != p.Config.DestinationRegistry.Registry {
		err = commandLogin(os.Stdout, p.Config.DestinationRegistry)
		if err != nil {
			return fmt.Errorf("Error authenticating: %s\n", err)

		}
	}

	cmds = []*exec.Cmd{}
	for _, tag := range p.Config.DestinationTagList {
		cmds = append(cmds, commandAddTag(p.Config, tag))
		cmds = append(cmds, commandPushTag(p.Config, tag))
	}

	err = runCommands(cmds)
	if err != nil {
		return err
	}

	return nil
}
