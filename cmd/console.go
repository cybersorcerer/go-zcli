package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Issue z/OS Console Commands",
	Long: `
DESCRIPTION
-----------
With the z/OS console commands, you can issue system commands and work with both
solicited messages (messages that were issued in response to the command)
and unsolicited messages (other messages that might or might not have been issued
in response to the command).`,
}

var consoleCommandCmd = &cobra.Command{
	Use:   "command",
	Short: "Issue z/OS command",
	Long: `
DESCRIPTION
-----------
You can use this command to issue a z/OS command and
get a corresponding response.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		command, _ := cmd.Flags().GetString("command")
		text, _ := cmd.Flags().GetBool("text")

		client := Profile.NewZosmfClient()
		payload := map[string]interface{}{
			"cmd": command,
		}

		resp, err := client.Put("/restconsoles/consoles/defcn", payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		if text {
			var respMap map[string]interface{}
			if err := json.Unmarshal(resp.Body, &respMap); err == nil {
				if cmdResp, ok := respMap["cmd-response"].(string); ok {
					for _, line := range strings.Split(cmdResp, "\r") {
						fmt.Println(strings.TrimSpace(line))
					}
					return nil
				}
			}
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

func init() {
	consoleCommandCmd.Flags().StringP("command", "c", "", "The z/OS command to issue.")
	consoleCommandCmd.MarkFlagRequired("command")
	consoleCommandCmd.Flags().Bool("text", false, "Display the response as text.")

	consoleCmd.AddCommand(consoleCommandCmd)
	rootCmd.AddCommand(consoleCmd)
}
