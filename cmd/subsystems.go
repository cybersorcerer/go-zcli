package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var subsystemsCmd = &cobra.Command{
	Use:   "subsystems",
	Short: "List z/OS subsystems",
	Long: `
DESCRIPTION
-----------
This service lists the subsystems on a z/OS system.`,
}

var subsystemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get information about z/OS subsystems",
	Long: `
DESCRIPTION
-----------
You can use the list subcommand to get information about the
subsystems on a z/OS system. You can filter the returned list
of subsystems by specifying a subsystem id or wild-card.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		client := Profile.NewZosmfClient()
		path := "/restfiles/mfs?type=subsys"
		if name != "" {
			path = fmt.Sprintf("/restfiles/mfs?type=subsys&subsys=%s", name)
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
	subsystemsListCmd.Flags().StringP("name", "n", "", "Name of a subsystem, if empty all subsystems are returned.")

	subsystemsCmd.AddCommand(subsystemsListCmd)
	rootCmd.AddCommand(subsystemsCmd)
}
