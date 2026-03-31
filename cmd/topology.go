package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var topologyCmd = &cobra.Command{
	Use:   "topology",
	Short: "Work with z/OSMF topology services",
	Long: `
DESCRIPTION
-----------
Provides commands for working with the groups, sysplexes,
central processor complexes (CPCs), and systems that are
defined to z/OSMF.`,
}

var topologyGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "List z/OSMF defined groups",
	Long: `
DESCRIPTION
-----------
You can use this operation to obtain a list
of the groups that are defined to a z/OSMF
instance.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()

		resp, err := client.Get("/resttopology/groups", nil)
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

var topologySysplexCmd = &cobra.Command{
	Use:   "sysplex",
	Short: "List z/OSMF defined sysplexes",
	Long: `
DESCRIPTION
-----------
You can use this operation to obtain a list
of the sysplexes that are defined to a z/OSMF
instance.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()

		resp, err := client.Get("/resttopology/sysplexes", nil)
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

// systems subgroup
var topologySystemsCmd = &cobra.Command{
	Use:   "systems",
	Short: "Work with z/OSMF defined systems",
	Long: `
DESCRIPTION
-----------
You can use these operations to obtain a list of the systems
that are defined to a z/OSMF instance, systems by z/OSMF group
or by z/OSMF defined sysplexes.`,
}

var topologySystemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all z/OSMF defined systems",
	Long: `
DESCRIPTION
-----------
You can use this operation to obtain a list
of the systems that are defined to a z/OSMF
instance.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()

		resp, err := client.Get("/resttopology/systems", nil)
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

var topologySystemsInGroupCmd = &cobra.Command{
	Use:   "in-group",
	Short: "List systems defined in a z/OSMF group",
	Long: `
DESCRIPTION
-----------
You can use this operation to obtain a list of the systems that
are included in a group.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		client := Profile.NewZosmfClient()

		path := fmt.Sprintf("/resttopology/systems/groupName/%s", name)
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

var topologySystemsInSysplexCmd = &cobra.Command{
	Use:   "in-sysplex",
	Short: "List systems defined in a z/OS parallel sysplex",
	Long: `
DESCRIPTION
-----------
You can use this operation to obtain a list of the systems that
are included in a z/OS parallel sysplex.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		client := Profile.NewZosmfClient()

		path := fmt.Sprintf("/resttopology/systems/sysplexName/%s", name)
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

// validate subgroup
var topologyValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate connection status of z/OS system(s)/plexes",
	Long: `
DESCRIPTION
-----------
Validate Connection status of z/OS system(s)/plexes.`,
}

var topologyValidateSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Check z/OS system connection status",
	Long: `
DESCRIPTION
-----------
You can use this operation to check the connection status of a
specified system which is managed through the z/OSMF Systems task.
If no system is provided, then validate LocalSystemDefinition.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		client := Profile.NewZosmfClient()

		path := "/services/systems/v1/validation/system"
		if name != "" {
			path += fmt.Sprintf("?system=%s", name)
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

var topologyValidatePlexCmd = &cobra.Command{
	Use:   "plex",
	Short: "Check z/OS plex systems connection status",
	Long: `
DESCRIPTION
-----------
You can use this operation to obtain a list of the systems
that are defined to a z/OSMF instance and validate them.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()

		resp, err := client.Get("/services/systems/v1/validation/plex", nil)
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
	// systems in-group
	topologySystemsInGroupCmd.Flags().StringP("name", "n", "", "The z/OSMF group name.")
	topologySystemsInGroupCmd.MarkFlagRequired("name")

	// systems in-sysplex
	topologySystemsInSysplexCmd.Flags().StringP("name", "n", "", "The z/OS sysplex name.")
	topologySystemsInSysplexCmd.MarkFlagRequired("name")

	// validate system
	topologyValidateSystemCmd.Flags().StringP("name", "n", "", "Name of z/OS system to validate.")

	// Wire up systems subgroup
	topologySystemsCmd.AddCommand(topologySystemsListCmd)
	topologySystemsCmd.AddCommand(topologySystemsInGroupCmd)
	topologySystemsCmd.AddCommand(topologySystemsInSysplexCmd)

	// Wire up validate subgroup
	topologyValidateCmd.AddCommand(topologyValidateSystemCmd)
	topologyValidateCmd.AddCommand(topologyValidatePlexCmd)

	// Wire up topology
	topologyCmd.AddCommand(topologyGroupsCmd)
	topologyCmd.AddCommand(topologySysplexCmd)
	topologyCmd.AddCommand(topologySystemsCmd)
	topologyCmd.AddCommand(topologyValidateCmd)

	rootCmd.AddCommand(topologyCmd)
}
