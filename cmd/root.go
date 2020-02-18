package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	cmdconfig "gerrit.instructure.com/muss/cmd/config"
	config "gerrit.instructure.com/muss/config"
)

// CommandBuilder is a function that takes the project config as an argument
// and returns a cobra command.
type CommandBuilder func(*config.ProjectConfig) *cobra.Command

var cmdBuilders = make([]CommandBuilder, 0)

// AddCommandBuilder takes the provided function and adds it to the list of
// commands that will be added to the root command when it is built.
func AddCommandBuilder(f CommandBuilder) {
	cmdBuilders = append(cmdBuilders, f)
}

// NewRootCommand takes a config value and returns a new root command.
func NewRootCommand(cfg *config.ProjectConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "muss",
		Short: "Configure and run project services",
		// Root command just shows help (which shows subcommands).
		// SilenceUsage and Errors so that we don't print excessively when dc exits non-zero.
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(cmdconfig.NewCommand(cfg))
	for _, f := range cmdBuilders {
		cmd.AddCommand(f(cfg))
	}
	return cmd
}

// Execute loads the config and runs the root command with the provided arguments.
func Execute(args []string) int {
	cfg, _ := config.All()
	cmd := NewRootCommand(cfg)
	return ExecuteRoot(cmd, args)
}

// ExecuteRoot executes the passed root command with the provided args.
// This simplifies testing.
func ExecuteRoot(rootCmd *cobra.Command, args []string) int {
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		// Propagate errors from command delegation.
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}

		stderr := rootCmd.ErrOrStderr()

		// Print information about other errors.
		fmt.Fprintln(stderr, "Error: ", err.Error())

		// Print usage if it's a flag error
		// TODO: Could this be any other type of error that we don't want to print usage for?
		cmd, _, findErr := rootCmd.Find(args)
		// If subcmd isn't found, print root command usage
		if findErr != nil {
			cmd = rootCmd
		}
		fmt.Fprintln(stderr, "") // Print blank line between "Error:" and "Usage:".
		fmt.Fprintln(stderr, cmd.UsageString())

		return 1
	}

	return 0
}
