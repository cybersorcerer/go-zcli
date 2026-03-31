package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"zcli/internal/config"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Work with zcli profiles",
	Long: `
DESCRIPTION
-----------
With the profile commands, you can work with zcli profile definitions.`,
}

var profileGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get zcli profile",
	Long: `
DESCRIPTION
-----------
You can use this command to get a list of a single or all zcli profile definitions.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		conf := config.Conf
		if conf == nil {
			config.LoadConfig()
			conf = config.Conf
		}

		defaultProfile, _ := config.GetDefaultProfileName("zosmf")

		getProfileValues := func(profileName string) map[string]interface{} {
			isDefault := profileName == defaultProfile
			return map[string]interface{}{
				"name":               profileName,
				"description":        config.GetProfileProperty(profileName, "zosmf", "description"),
				"user":               config.GetProfileProperty(profileName, "zosmf", "user"),
				"protocol":           config.GetProfileProperty(profileName, "zosmf", "protocol"),
				"host":               config.GetProfileProperty(profileName, "zosmf", "host"),
				"port":               config.GetProfileProperty(profileName, "zosmf", "port"),
				"rejectUnauthorized": config.GetProfileProperty(profileName, "zosmf", "rejectUnauthorized"),
				"default":            isDefault,
				"home":               config.GetProfileProperty(profileName, "zosmf", "home"),
			}
		}

		if name != "" {
			if _, ok := conf.Profiles[name]; !ok {
				fmt.Fprintf(os.Stderr, "ZCLI-PROFILE-001E Profile %s not found, terminating rc = 12\n", name)
				os.Exit(12)
			}
			values := getProfileValues(name)
			data, _ := json.Marshal(values)
			fmt.Println(string(data))
		} else {
			for profileName := range conf.Profiles {
				values := getProfileValues(profileName)
				data, _ := json.Marshal(values)
				fmt.Println(string(data))
			}
		}
		return nil
	},
}

func init() {
	profileGetCmd.Flags().StringP("name", "n", "", "Profile name to retrieve, all profiles if empty.")

	profileCmd.AddCommand(profileGetCmd)
	rootCmd.AddCommand(profileCmd)
}
