package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Retrieve information about z/OSMF",
	Long: `
DESCRIPTION
-----------
Use this command to retrieve information about z/OSMF.
This service allows the caller to query the version and other details
about the instance of z/OSMF running on a particular system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()
		resp, err := client.Get("/info", nil)
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
	rootCmd.AddCommand(infoCmd)
}
