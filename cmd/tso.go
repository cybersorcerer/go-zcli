package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"zcli/internal/config"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		command, _ := cmd.Flags().GetString("command")
		text, _ := cmd.Flags().GetBool("text")

		client := Profile.NewZosmfClient()

		// Get TSO properties from config
		account := config.GetTSOProperty("account")
		if account == "" {
			account = "ACCT#"
		}

		// Start TSO address space
		startPayload := map[string]interface{}{
			"startParameters": map[string]interface{}{
				"account":        account,
				"proc":           config.GetTSOProperty("logonProcedure"),
				"characterSet":   config.GetTSOProperty("characterSet"),
				"codePage":       config.GetTSOProperty("codePage"),
				"columns":        80,
				"regionSize":     8192,
				"rows":           24,
			},
		}

		startResp, err := client.Post("/tsoApp/tso", startPayload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(startResp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(startResp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse TSO start response: %w", err)
		}

		servletKey, ok := startResult["servletKey"].(string)
		if !ok {
			return fmt.Errorf("no servletKey in TSO start response")
		}

		// Issue TSO command
		cmdPayload := map[string]interface{}{
			"TSO COMMAND": command,
		}

		cmdResp, err := client.Put(fmt.Sprintf("/tsoApp/tso/%s", servletKey), cmdPayload, nil)
		if err != nil {
			return err
		}

		// Stop TSO address space (best effort)
		defer client.Delete(fmt.Sprintf("/tsoApp/tso/%s", servletKey), nil)

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
