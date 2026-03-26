package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var rtdCmd = &cobra.Command{
	Use:   "rtd",
	Short: "Retrieve Runtime Diagnostic Data from z/OS",
	Long: `
DESCRIPTION
-----------
Use this command to retrieve Runtime Diagnostic Data from z/OS.
Asname is the name of the address space name or name prefix.
DEFAULT: All address spaces are analyzed by Runtime Diagnostics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		asName, _ := cmd.Flags().GetString("address-space-name")

		client := Profile.NewZosmfClient()
		path := "/diagnostic/rest/1.0/diag"
		if asName != "" {
			path += fmt.Sprintf("?asname=%s", asName)
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
	rtdCmd.Flags().StringP("address-space-name", "a", "", "Name of a z/OS address space.")
	rootCmd.AddCommand(rtdCmd)
}
