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
get a corresponding response. The command can be issued
synchronously or asynchronously. You can detect keywords
in solicited and unsolicited messages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		command, _ := cmd.Flags().GetString("command")
		text, _ := cmd.Flags().GetBool("text")
		consoleName, _ := cmd.Flags().GetString("console-name")
		async, _ := cmd.Flags().GetBool("async")
		system, _ := cmd.Flags().GetString("system")
		solKey, _ := cmd.Flags().GetString("sol-key")
		solKeyReg, _ := cmd.Flags().GetBool("sol-key-regex")
		unsolKey, _ := cmd.Flags().GetString("unsol-key")
		unsolKeyReg, _ := cmd.Flags().GetBool("unsol-key-regex")
		detectTime, _ := cmd.Flags().GetInt("detect-time")
		unsolDetectSync, _ := cmd.Flags().GetBool("unsol-detect-sync")
		unsolDetectTimeout, _ := cmd.Flags().GetInt("unsol-detect-timeout")
		auth, _ := cmd.Flags().GetString("auth")
		routcode, _ := cmd.Flags().GetString("routcode")
		mscope, _ := cmd.Flags().GetString("mscope")
		storage, _ := cmd.Flags().GetInt("storage")
		autoFlag, _ := cmd.Flags().GetString("auto")

		client := Profile.NewZosmfClient()

		payload := map[string]interface{}{
			"cmd": command,
		}

		if async {
			payload["async"] = "Y"
		}
		if system != "" {
			payload["system"] = system
		}
		if solKey != "" {
			payload["sol-key"] = solKey
			if solKeyReg {
				payload["solKeyReg"] = "Y"
			}
		}
		if unsolKey != "" {
			payload["unsol-key"] = unsolKey
			if unsolKeyReg {
				payload["unsolKeyReg"] = "Y"
			}
			if detectTime > 0 {
				payload["detect-time"] = detectTime
			}
			if unsolDetectSync {
				payload["unsol-detect-sync"] = "Y"
				if unsolDetectTimeout > 0 {
					payload["unsol-detect-timeout"] = unsolDetectTimeout
				}
			}
		}
		if auth != "" {
			payload["auth"] = auth
		}
		if routcode != "" {
			payload["routcode"] = routcode
		}
		if mscope != "" {
			payload["mscope"] = mscope
		}
		if cmd.Flags().Changed("storage") {
			payload["storage"] = storage
		}
		if autoFlag != "" {
			payload["auto"] = autoFlag
		}

		path := fmt.Sprintf("/restconsoles/consoles/%s", consoleName)
		resp, err := client.Put(path, payload, nil)
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
				if cmdResp, ok := respMap["cmd-response"].(string); ok && cmdResp != "" {
					for _, line := range strings.Split(cmdResp, "\r") {
						fmt.Println(strings.TrimSpace(line))
					}
				} else if cmdResp == "" {
					fmt.Println("No immediate response. Use --sol-key or retrieve via cmd-response-key.")
				}
				if detected, ok := respMap["sol-key-detected"].(bool); ok {
					fmt.Printf("\nsol-key-detected: %v\n", detected)
				}
				if status, ok := respMap["status"].(string); ok {
					fmt.Printf("\nstatus: %s\n", status)
				}
				if msg, ok := respMap["msg"].(string); ok && msg != "" {
					fmt.Printf("msg: %s\n", msg)
				}
				if key, ok := respMap["cmd-response-key"].(string); ok {
					fmt.Printf("\ncmd-response-key: %s\n", key)
				}
				if dkey, ok := respMap["detection-key"].(string); ok {
					fmt.Printf("detection-key: %s\n", dkey)
				}
				return nil
			}
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

var consoleGetResponseCmd = &cobra.Command{
	Use:   "get-response",
	Short: "Retrieve console command response",
	Long: `
DESCRIPTION
-----------
Retrieve the command response for a previously issued command
using the cmd-response-key returned by the command request.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		consoleName, _ := cmd.Flags().GetString("console-name")
		responseKey, _ := cmd.Flags().GetString("response-key")
		text, _ := cmd.Flags().GetBool("text")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restconsoles/consoles/%s/solmsgs/%s", consoleName, responseKey)

		resp, err := client.Get(path, nil)
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

var consoleGetDetectionCmd = &cobra.Command{
	Use:   "get-detection",
	Short: "Retrieve unsolicited keyword detection result",
	Long: `
DESCRIPTION
-----------
Retrieve the result of an unsolicited keyword detection
using the detection-key returned by the command request.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		consoleName, _ := cmd.Flags().GetString("console-name")
		detectionKey, _ := cmd.Flags().GetString("detection-key")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restconsoles/consoles/%s/detections/%s", consoleName, detectionKey)

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

var consoleLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Get messages from a hardcopy log",
	Long: `
DESCRIPTION
-----------
Retrieve messages from the hardcopy log (OPERLOG or SYSLOG).
You can specify a time, time range, and direction to filter
the messages returned. Maximum 10000 log entries per request.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		time, _ := cmd.Flags().GetString("time")
		timestamp, _ := cmd.Flags().GetString("timestamp")
		timeRange, _ := cmd.Flags().GetString("time-range")
		direction, _ := cmd.Flags().GetString("direction")
		hardcopy, _ := cmd.Flags().GetString("hardcopy")
		sysName, _ := cmd.Flags().GetString("sys-name")
		text, _ := cmd.Flags().GetBool("text")

		client := Profile.NewZosmfClient()
		path := "/restconsoles/v1/log"

		var params []string
		if timestamp != "" {
			params = append(params, "timestamp="+timestamp)
		} else if time != "" {
			params = append(params, "time="+time)
		}
		if timeRange != "" {
			params = append(params, "timeRange="+timeRange)
		}
		if direction != "" {
			params = append(params, "direction="+direction)
		}
		if hardcopy != "" {
			params = append(params, "hardcopy="+hardcopy)
		}
		if sysName != "" {
			params = append(params, "sysName="+sysName)
		}
		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		resp, err := client.Get(path, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		if text {
			var result struct {
				Source     string `json:"source"`
				TotalItems int   `json:"totalitems"`
				Items      []struct {
					Time      string `json:"time"`
					JobName   string `json:"jobName"`
					System    string `json:"system"`
					MessageID string `json:"messageId"`
					Message   string `json:"message"`
					SubType   string `json:"subType"`
				} `json:"items"`
			}
			if err := json.Unmarshal(resp.Body, &result); err == nil {
				fmt.Printf("Source: %s | Total: %d\n\n", result.Source, result.TotalItems)
				for _, item := range result.Items {
					msg := strings.TrimSpace(item.Message)
					fmt.Printf("%s %-8s %s\n", item.Time, strings.TrimSpace(item.JobName), msg)
				}
				return nil
			}
		}

		fmt.Println(resp.BodyString())
		return nil
	},
}

func init() {
	// command subcommand
	consoleCommandCmd.Flags().StringP("command", "c", "", "The z/OS command to issue.")
	consoleCommandCmd.MarkFlagRequired("command")
	consoleCommandCmd.Flags().Bool("text", false, "Display the response as formatted text.")
	consoleCommandCmd.Flags().StringP("console-name", "n", "defcn", "EMCS console name (2-8 chars) or defcn for default.")
	consoleCommandCmd.Flags().BoolP("async", "a", false, "Issue the command asynchronously.")
	consoleCommandCmd.Flags().StringP("system", "s", "", "Target system name in the sysplex.")
	consoleCommandCmd.Flags().String("sol-key", "", "Keyword to detect in solicited messages (command response).")
	consoleCommandCmd.Flags().Bool("sol-key-regex", false, "Treat sol-key as a regular expression.")
	consoleCommandCmd.Flags().String("unsol-key", "", "Keyword to detect in unsolicited messages.")
	consoleCommandCmd.Flags().Bool("unsol-key-regex", false, "Treat unsol-key as a regular expression.")
	consoleCommandCmd.Flags().Int("detect-time", 0, "Seconds to detect unsol-key (default 30 on server).")
	consoleCommandCmd.Flags().Bool("unsol-detect-sync", false, "Detect unsol-key synchronously (block until result or timeout).")
	consoleCommandCmd.Flags().Int("unsol-detect-timeout", 0, "Timeout in seconds for synchronous unsol-key detection (default 20 on server).")
	consoleCommandCmd.Flags().String("auth", "", "Console command authority: MASTER, ALL, INFO, CONS, IO, SYS.")
	consoleCommandCmd.Flags().String("routcode", "", "Console routing codes: ALL, NONE, or list (e.g. 1,2,3 or 1:10).")
	consoleCommandCmd.Flags().String("mscope", "", "Console message scope: ALL, LOCAL, or system names.")
	consoleCommandCmd.Flags().Int("storage", 0, "Console storage in KB for message queuing (1-2000).")
	consoleCommandCmd.Flags().String("auto", "", "Console automation: YES or NO.")

	// get-response subcommand
	consoleGetResponseCmd.Flags().StringP("console-name", "n", "defcn", "EMCS console name.")
	consoleGetResponseCmd.Flags().StringP("response-key", "k", "", "The cmd-response-key from a previous command.")
	consoleGetResponseCmd.MarkFlagRequired("response-key")
	consoleGetResponseCmd.Flags().Bool("text", false, "Display the response as formatted text.")

	// get-detection subcommand
	consoleGetDetectionCmd.Flags().StringP("console-name", "n", "defcn", "EMCS console name.")
	consoleGetDetectionCmd.Flags().StringP("detection-key", "k", "", "The detection-key from a previous command.")
	consoleGetDetectionCmd.MarkFlagRequired("detection-key")

	// log subcommand
	consoleLogCmd.Flags().String("time", "", "Start time in ISO 8601 format (e.g. 2021-01-26T03:33:18.065Z).")
	consoleLogCmd.Flags().String("timestamp", "", "Start time as UNIX timestamp in milliseconds (overrides --time).")
	consoleLogCmd.Flags().String("time-range", "", "Time range: nnnu where u is s/m/h (e.g. 10m, 1h, 30s). Default: 10m.")
	consoleLogCmd.Flags().String("direction", "", "Direction from start time: forward or backward (default: backward).")
	consoleLogCmd.Flags().String("hardcopy", "", "Log source: operlog or syslog (default: operlog, fallback syslog).")
	consoleLogCmd.Flags().String("sys-name", "", "System name for SYSLOG (only valid with --hardcopy syslog).")
	consoleLogCmd.Flags().Bool("text", false, "Display messages as formatted text.")

	consoleCmd.AddCommand(consoleCommandCmd)
	consoleCmd.AddCommand(consoleGetResponseCmd)
	consoleCmd.AddCommand(consoleGetDetectionCmd)
	consoleCmd.AddCommand(consoleLogCmd)
	rootCmd.AddCommand(consoleCmd)
}
