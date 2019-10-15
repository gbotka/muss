package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newBuildCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build",
		Short: "Build or rebuild services",
		Long: `Build or rebuild services.

Services are built once and then tagged as "project_service",
e.g. "myapp_db". If you change a service's "Dockerfile" or the
contents of its build directory, you can run "build" to rebuild it.

Usage: build [options] [--build-arg key=val...] [SERVICE...]

Options:
  --compress              Compress the build context using gzip.
  --force-rm              Always remove intermediate containers.
  --no-cache              Do not use cache when building the image.
  --pull                  Always attempt to pull a newer version of the image.
  -m, --memory MEM        Sets memory limit for the build container.
  --build-arg key=val     Set build-time variables for services.
  --parallel              Build images in parallel.
`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Save()
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	// Just show "Long" by providing a zero-width (but not zero-value) string.
	cmd.SetUsageTemplate(`{{ "" }}`)

	return cmd
}

func init() {
	rootCmd.AddCommand(newBuildCommand())
}
