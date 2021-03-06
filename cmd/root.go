package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MichaelDarr/ahab/internal"
	ahab "github.com/MichaelDarr/ahab/pkg"
)

// flags used across the CLI
var asRoot bool
var verbose bool

// rootCmd is the base command which all others are added to
var rootCmd = &cobra.Command{
	Use:     "ahab",
	Short:   "Configure, launch, and work in Dockerized environments",
	Version: internal.Version,
}

// Execute is the entrypoint to the CLI
func Execute() {
	rootCmd.Execute()
}

// Docker commands that don't take options
var diffCmd = BasicCommand("diff", "Inspect changes to files or directories on container filesystem")
var pauseCmd = BasicCommand("pause", "Pause all processes within container")
var portCmd = BasicCommand("port", "List port mappings for the container")
var topCmd = BasicCommand("top", "Display the running processes of the container")
var unpauseCmd = BasicCommand("unpause", "Unpause all processes within container")
var waitCmd = BasicCommand("wait", "Block until the container stops, then print its exit code")

// Docker commands that take options
var attachCmd = OptionCommand("attach", "Attach local standard input, output, and error streams to container")
var commitCmd = OptionCommand("commit", "Create a new image from container's changes")
var exportCmd = OptionCommand("export", "Export container’s filesystem as a tar archive")
var killCmd = OptionCommand("kill", "Kill container")
var logsCmd = OptionCommand("logs", "Fetch the container logs")
var restartCmd = OptionCommand("restart", "Restart container")
var rmCmd = OptionCommand("rm", "Remove container")
var startCmd = OptionCommand("start", "Start stopped container")
var statsCmd = OptionCommand("stats", "Display a live stream of container resource usage statistics")
var stopCmd = OptionCommand("stop", "Stop running container")
var updateCmd = OptionCommand("update", "Update configuration of the container")

// Shell attachment commands
var bashCmd = ShellCommand("bash", "bash")
var shCmd = ShellCommand("sh", "bourne")
var zshCmd = ShellCommand("zsh", "z")

// init adds all the generic subcommands
func init() {
	rootCmd.AddCommand(&attachCmd)
	rootCmd.AddCommand(&bashCmd)
	rootCmd.AddCommand(&commitCmd)
	rootCmd.AddCommand(&diffCmd)
	rootCmd.AddCommand(&exportCmd)
	rootCmd.AddCommand(&killCmd)
	rootCmd.AddCommand(&logsCmd)
	rootCmd.AddCommand(&pauseCmd)
	rootCmd.AddCommand(&portCmd)
	rootCmd.AddCommand(&restartCmd)
	rootCmd.AddCommand(&rmCmd)
	rootCmd.AddCommand(&shCmd)
	rootCmd.AddCommand(&startCmd)
	rootCmd.AddCommand(&statsCmd)
	rootCmd.AddCommand(&stopCmd)
	rootCmd.AddCommand(&topCmd)
	rootCmd.AddCommand(&unpauseCmd)
	rootCmd.AddCommand(&updateCmd)
	rootCmd.AddCommand(&waitCmd)
	rootCmd.AddCommand(&zshCmd)
}

// BasicCommand constructs and returns a Docker container command which doesn not take extra options
func BasicCommand(command string, description string) cobra.Command {
	return cobra.Command{
		Use:   command,
		Short: description,
		Long: description + `

Docker Command:
  docker ` + command + ` CONTAINER`,
		Run: func(cmd *cobra.Command, args []string) {
			container, err := internal.GetContainer()
			ahab.PrintErrFatal(err)
			ahab.PrintErrFatal(container.Cmd(command))
		},
	}
}

// OptionCommand constructs and returns a Docker container command which takes extra options
func OptionCommand(command string, description string) cobra.Command {
	return cobra.Command{
		Use:   command,
		Short: description,
		Run: func(cmd *cobra.Command, args []string) {
			helpRequested, err := internal.PrintDockerHelp(&args, command, description+`

Docker Command:
  docker `+command+` [OPTIONS] CONTAINER

Usage:
  ahab `+command+` [-h/--help] [OPTIONS]
`)
			ahab.PrintErrFatal(err)
			if helpRequested {
				return
			}
			container, err := internal.GetContainer()
			ahab.PrintErrFatal(err)

			containerOpts := append([]string{command}, args...)
			containerOpts = append(containerOpts, container.Name())
			ahab.PrintErrFatal(internal.DockerCmd(&containerOpts))
		},
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
	}
}

// ShellCommand constructs and returns a Docker container command to attach a shell to the active terminal
func ShellCommand(shellCommand string, description string) (cmd cobra.Command) {
	cmd = cobra.Command{
		Use:   shellCommand,
		Short: "Open a containerized " + description + " shell",
		Long: `Attach a containerized ` + description + ` shell to the active terminal.

*Warning!* the ` + description + ` shell must be installed in your image for this command to function!

Docker Command:
  docker exec -it [-u ` + internal.ContainerUserName + `] CONTAINER ` + shellCommand,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			container, err := internal.GetContainer()
			ahab.PrintErrFatal(err)
			ahab.PrintErrFatal(container.Up())

			execArgs := []string{"exec", "-it"}
			if asRoot {
				execArgs = append(execArgs, "-u", "root")
			} else if container.Fields.User != "" {
				execArgs = append(execArgs, "-u", container.Fields.User)
			} else if !container.Fields.Permissions.Disable {
				execArgs = append(execArgs, "-u", internal.ContainerUserName)
			}
			execArgs = append(execArgs, container.Name(), shellCommand)
			ahab.PrintErrFatal(internal.DockerCmd(&execArgs))
		},
	}

	cmd.Flags().BoolVar(&asRoot, "root", false, "Use "+description+" shell as root")
	return cmd
}
