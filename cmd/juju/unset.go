// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"errors"
	"fmt"
	"strings"

	"launchpad.net/gnuflag"

	"launchpad.net/juju-core/charm"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/juju"
)

// UnsetCommand sets configuration values of a service back
// to their default.
type UnsetCommand struct {
	cmd.EnvCommandBase
	ServiceName string
	Options     []string
}

const unsetDoc = `
Set one or more configuration options for the specified service
to their default. See also the set commmand to set one or more 
configuration options for a specified service.
`

func (c *UnsetCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "unset",
		Args:    "<service> name ...",
		Purpose: "set service config options back to their default",
		Doc:     unsetDoc,
	}
}

func (c *UnsetCommand) SetFlags(f *gnuflag.FlagSet) {
	c.EnvCommandBase.SetFlags(f)
}

func (c *UnsetCommand) Init(args []string) error {
	if len(args) == 0 {
		return errors.New("no service name specified")
	}
	c.ServiceName = args[0]
	c.Options = args[1:]
	return nil
}

// Run resets the configuration of a service.
func (c *UnsetCommand) Run(ctx *cmd.Context) error {
	conn, err := juju.NewConnFromName(c.EnvName)
	if err != nil {
		return err
	}
	defer conn.Close()
	service, err := conn.State.Service(c.ServiceName)
	if err != nil {
		return err
	}
	ch, _, err := service.Charm()
	if err != nil {
		return err
	}
	if len(c.Options) > 0 {
		settings := make(charm.Settings)
		defaultSettings := ch.Config().DefaultSettings()
		for _, option := range c.Options {
			defaultSetting, ok := defaultSettings[option]
			if !ok {
				if strings.Contains(option, "=") {
					return fmt.Errorf("invalid setting during unset: %q", option)
				}
				return fmt.Errorf("invalid option: %q", option)
			}
			settings[option] = defaultSetting
		}
		return service.UpdateConfigSettings(settings)
	} else {
		return nil
	}
}
