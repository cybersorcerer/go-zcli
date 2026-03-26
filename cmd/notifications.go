package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var notificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "Work with z/OSMF notification services",
	Long: `
DESCRIPTION
-----------
These commands are used to send a notification in the form of a notification
record or email, to a single or multiple recipients.`,
}

var notificationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get all notifications received by the current user",
	Long: `
DESCRIPTION
-----------
You can use the list command to get all of the notifications that were
received by the current user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()

		resp, err := client.Get("/notifications/inbox", nil)
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

var notificationsSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a notification and mail from a z/OSMF user or task",
	Long: `
DESCRIPTION
-----------
You can use the send subcommand to send a notification AND mail. The
notification details are in the file specified as input option to this
command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, _ := cmd.Flags().GetString("file-name")

		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("error reading input file %s: %w", fileName, err)
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post("/notifications/new", string(data), nil)
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
	notificationsSendCmd.Flags().StringP("file-name", "f", "", "The filename with the notification input data.")
	notificationsSendCmd.MarkFlagRequired("file-name")

	notificationsCmd.AddCommand(notificationsListCmd)
	notificationsCmd.AddCommand(notificationsSendCmd)

	rootCmd.AddCommand(notificationsCmd)
}
