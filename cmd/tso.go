package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var tsoCmd = &cobra.Command{
	Use:   "tso",
	Short: "Work with TSO/E address space services",
	Long: `
DESCRIPTION
-----------
Work with TSO/E address space services on a z/OS system.`,
}

var tsoCommandCmd = &cobra.Command{
	Use:   "command",
	Short: "Issue TSO/E command",
	Long: `
DESCRIPTION
-----------
You can use this operation to issue a TSO/E command and
get a corresponding response.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		command, _ := cmd.Flags().GetString("command")
		text, _ := cmd.Flags().GetBool("text")

		client := Profile.NewZosmfClient()

		// Issue TSO command via PUT /zosmf/tsoApp/v1/tso
		cmdPayload := map[string]interface{}{
			"tsoCmd": command,
		}

		cmdResp, err := client.Put("/tsoApp/v1/tso", cmdPayload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(cmdResp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		if text {
			var respMap map[string]interface{}
			if err := json.Unmarshal(cmdResp.Body, &respMap); err == nil {
				if cmdResponse, ok := respMap["cmdResponse"].([]interface{}); ok {
					for _, line := range cmdResponse {
						if lineMap, ok := line.(map[string]interface{}); ok {
							if msg, ok := lineMap["message"].(string); ok {
								fmt.Println(msg)
							}
						}
					}
					return nil
				}
			}
		}
		fmt.Println(cmdResp.BodyString())
		return nil
	},
}

func init() {
	tsoCommandCmd.Flags().StringP("command", "c", "", "The TSO/E command to issue.")
	tsoCommandCmd.MarkFlagRequired("command")
	tsoCommandCmd.Flags().Bool("text", false, "Display the response as text.")

	tsoCmd.AddCommand(tsoCommandCmd)
	rootCmd.AddCommand(tsoCmd)
}
