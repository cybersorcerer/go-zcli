package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var sysvarCmd = &cobra.Command{
	Use:   "sysvar",
	Short: "Interact with z/OSMF and System variables",
	Long: `
DESCRIPTION
-----------
Interact with the z/OSMF and System variables.`,
}

var sysvarGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve z/OSMF and system variables",
	Long: `
DESCRIPTION
-----------
Use this command to retrieve z/OSMF and system variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plexName, _ := cmd.Flags().GetString("plex-name")
		systemName, _ := cmd.Flags().GetString("system-name")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems?sysplex-name=%s", plexName)
		if systemName != "" {
			path += fmt.Sprintf("&system-name=%s", systemName)
		}

		resp, err := client.Get(path, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

func init() {
	sysvarGetCmd.Flags().StringP("plex-name", "x", "", "The name of a z/OS sysplex.")
	sysvarGetCmd.MarkFlagRequired("plex-name")
	sysvarGetCmd.Flags().StringP("system-name", "y", "", "The name of a z/OS system.")

	sysvarCmd.AddCommand(sysvarGetCmd)
	rootCmd.AddCommand(sysvarCmd)
}
