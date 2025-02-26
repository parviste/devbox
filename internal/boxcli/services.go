// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package boxcli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/devbox"
)

type servicesCmdFlags struct {
	config configFlags
}

type serviceManagerCmdFlag struct {
	configFlags
	processComposeFile string
}

func (flags *serviceManagerCmdFlag) register(cmd *cobra.Command) {
	flags.configFlags.register(cmd)
	cmd.Flags().StringVar(
		&flags.processComposeFile,
		"process-compose-file",
		"",
		"path to process compose file or directory containing process "+
			"compose-file.yaml|yml. Default is directory containing devbox.json",
	)
}

func ServicesCmd() *cobra.Command {
	flags := servicesCmdFlags{}
	managerFlags := serviceManagerCmdFlag{}
	servicesCommand := &cobra.Command{
		Use:   "services",
		Short: "Interact with devbox services",
	}

	lsCommand := &cobra.Command{
		Use:   "ls",
		Short: "List available services",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listServices(cmd, flags)
		},
	}

	startCommand := &cobra.Command{
		Use:   "start [service]...",
		Short: "Starts service. If no service is specified, starts all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServices(cmd, args, flags)
		},
	}

	stopCommand := &cobra.Command{
		Use:   "stop [service]...",
		Short: "Stops service. If no service is specified, stops all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stopServices(cmd, args, flags)
		},
	}

	restartCommand := &cobra.Command{
		Use:   "restart [service]...",
		Short: "Restarts service. If no service is specified, restarts all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return restartServices(cmd, args, flags)
		},
	}

	processManagerCommand := &cobra.Command{
		Use:   "manager",
		Short: "Starts process manager with all supported services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startProcessManager(cmd, managerFlags)
		},
	}

	flags.config.register(servicesCommand)
	managerFlags.register(processManagerCommand)
	servicesCommand.AddCommand(lsCommand)
	servicesCommand.AddCommand(processManagerCommand)
	servicesCommand.AddCommand(restartCommand)
	servicesCommand.AddCommand(startCommand)
	servicesCommand.AddCommand(stopCommand)
	return servicesCommand
}

func listServices(cmd *cobra.Command, flags servicesCmdFlags) error {
	box, err := devbox.Open(flags.config.path, cmd.ErrOrStderr())
	if err != nil {
		return errors.WithStack(err)
	}
	services, err := box.Services()
	if err != nil {
		return err
	}
	for _, service := range services {
		cmd.Println(service.Name)
	}
	return nil
}

func startServices(cmd *cobra.Command, services []string, flags servicesCmdFlags) error {
	box, err := devbox.Open(flags.config.path, cmd.ErrOrStderr())
	if err != nil {
		return errors.WithStack(err)
	}
	if len(services) == 0 {
		services, err = serviceNames(box)
		if err != nil {
			return err
		}
		if len(services) == 0 {
			cmd.Println("No services to start")
			return nil
		}
	}
	return box.StartServices(cmd.Context(), services...)
}

func stopServices(cmd *cobra.Command, services []string, flags servicesCmdFlags) error {
	box, err := devbox.Open(flags.config.path, cmd.ErrOrStderr())
	if err != nil {
		return errors.WithStack(err)
	}
	if len(services) == 0 {
		services, err = serviceNames(box)
		if err != nil {
			return err
		}
		if len(services) == 0 {
			cmd.Println("No services to stop")
			return nil
		}
	}
	return box.StopServices(cmd.Context(), services...)
}

func serviceNames(box devbox.Devbox) ([]string, error) {
	services, err := box.Services()
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, service := range services {
		names = append(names, service.Name)
	}
	return names, nil
}

func restartServices(
	cmd *cobra.Command,
	services []string,
	flags servicesCmdFlags,
) error {
	_ = stopServices(cmd, services, flags)
	return startServices(cmd, services, flags)
}

func startProcessManager(cmd *cobra.Command, flags serviceManagerCmdFlag) error {
	box, err := devbox.Open(flags.path, cmd.ErrOrStderr())
	if err != nil {
		return errors.WithStack(err)
	}
	return box.StartProcessManager(cmd.Context(), flags.processComposeFile)
}
