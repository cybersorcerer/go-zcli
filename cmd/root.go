package cmd

import (
	"fmt"
	"os"
	"zcli/internal/config"
	logger "zcli/internal/logging"

	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
)

// Global flags accessible to all subcommands.
var (
	profileName string
	verify      bool
	debug       bool
	format      string
)

// ProfileData resolved from the config, shared with subcommands.
var Profile *config.ProfileData

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "zcli",
	Short: "Command Line Interface to z/OS",
	Long: `
DESCRIPTION
-----------
zCli is a command line interface (CLI) to z/OS Rest Services allowing you
to interact with IBM z/OS from your local machines.

Program Name.: z/OS CLI (zcli)
Author.......: Ronny Funk
Function.....: z/OS z/OSMF REST API CLI

Environment: *ix Terminal CLI / Batch Job`,
	Version: "dev",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			logger.Lvl.Set(logger.MyDebug)
			logger.Log.Debug("Debug mode enabled")
		}

		logger.Log.Debug("zcli started with",
			"profile-name", profileName,
			"verify", verify,
			"debug", debug,
			"format", format,
		)

		// Skip profile resolution for commands that don't need z/OSMF (e.g. profile get)
		if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "version" {
			return nil
		}
		// Also skip for the profile command group itself
		if cmd.Parent() != nil && cmd.Parent().Name() == "profile" {
			return nil
		}
		if cmd.Name() == "profile" && !cmd.HasSubCommands() {
			return nil
		}

		var err error
		Profile, err = config.ResolveProfile(profileName, verify)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		logger.Log.Debug("Profile resolved",
			"profile", Profile.ProfileName,
			"host", Profile.HostName,
			"port", Profile.Port,
			"user", Profile.User,
		)

		return nil
	},
}

// SetVersionInfo sets the version string displayed by --version.
func SetVersionInfo(version, commit string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s)\nCopyright (c) 2025, 2026 Sir Tobi aka Cybersorcerer", version, commit)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	cc.Init(&cc.Config{
		RootCmd:         rootCmd,
		Headings:        cc.HiRed + cc.Bold + cc.Underline,
		Commands:        cc.HiYellow + cc.Bold,
		CmdShortDescr:   cc.Blue,
		Example:         0,
		ExecName:        cc.Bold + cc.HiYellow,
		Flags:           cc.Bold + cc.HiYellow,
		FlagsDataType:   0,
		FlagsDescr:      cc.Blue,
		NoExtraNewlines: true,
		NoBottomNewline: true,
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profileName, "profile-name", "", "", "z/OSMF Profile to use.")
	rootCmd.PersistentFlags().BoolVar(&verify, "verify", false, "Turn certificate verification on (--verify=true) or off (--verify=false).")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Turn debugging on (--debug) or off (--debug=false).")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format: json or text.")
}
