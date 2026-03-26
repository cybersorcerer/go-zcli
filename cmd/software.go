package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"zcli/internal/config"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

// pollAsyncOperation polls a z/OSMF async statusurl until completion.
func pollAsyncOperation(client *zosmf.Client, statusURL string) (*zosmf.Response, error) {
	// statusURL from z/OSMF is a full URL; extract the path after /zosmf
	path := statusURL
	if idx := strings.Index(statusURL, "/zosmf"); idx >= 0 {
		path = statusURL[idx+len("/zosmf"):]
	}

	for {
		resp, err := client.Get(path, nil)
		if err != nil {
			return nil, err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			return resp, apiErr
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			return resp, nil
		}

		status, _ := result["status"].(string)
		if status != "running" {
			return resp, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// resolveSwInstancePath builds the URL path for software instance operations.
func resolveSwInstancePath(basePath, nickName, swiName, uuid string) (string, error) {
	if uuid != "" {
		return basePath + "/" + uuid, nil
	}
	if nickName != "" && swiName != "" {
		return basePath + "/" + nickName + "/" + swiName, nil
	}
	return "", fmt.Errorf("either --uuid or both --nick-name and --swi-name must be specified")
}

var softwareCmd = &cobra.Command{
	Use:   "software",
	Short: "Interact with the z/OSMF Software Management task",
	Long: `
DESCRIPTION
-----------
Interact with the z/OSMF Software Management task.`,
}

// query subgroup
var softwareQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query z/OS SMP/E",
	Long: `
DESCRIPTION
-----------
The SMP/E CSI Query service allows you to query entries defined
in SMP/E CSI data sets.`,
}

var softwareQueryCsidsCmd = &cobra.Command{
	Use:   "csids",
	Short: "Query SMP/E CSI data sets",
	Long: `
DESCRIPTION
-----------
SMP/E CSI Query action is to be performed on the identified
global zone CSI data set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCSI, _ := cmd.Flags().GetString("global-csi")
		zones, _ := cmd.Flags().GetString("zones")
		entry, _ := cmd.Flags().GetString("entry")
		subentries, _ := cmd.Flags().GetString("subentries")
		filter, _ := cmd.Flags().GetString("filter")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/swmgmt/csi/csiquery/%s", globalCSI)

		payload := map[string]interface{}{
			"zones":   strings.Split(zones, ","),
			"entries": strings.Split(entry, ","),
		}
		if subentries != "" {
			payload["subentries"] = strings.Split(subentries, ",")
		}
		if filter != "" {
			payload["filter"] = filter
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

var softwareQueryCritUpdatesCmd = &cobra.Command{
	Use:   "critupdates",
	Short: "Get information about Missing Critical Updates",
	Long: `
DESCRIPTION
-----------
The Missing Critical Updates command helps you determine if your software
instances are missing software updates to resolve PE PTFs, HIPER fixes, or
other exception SYSMODs identified by ERROR HOLDDATA.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post(instancePath+"/missingcriticalupdates", nil, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

var softwareQuerySoftUpdatesCmd = &cobra.Command{
	Use:   "softupdates",
	Short: "Search Software updates",
	Long: `
DESCRIPTION
-----------
The Software Update Search command allows you to search a software instance
for one or more software updates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		payload := map[string]interface{}{
			"updates": args,
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post(instancePath+"/softwareupdatesearch", payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

var softwareQueryFixcatUpdatesCmd = &cobra.Command{
	Use:   "fixcatupdates",
	Short: "Get information about missing FIXCAT Updates",
	Long: `
DESCRIPTION
-----------
The Missing FIXCAT Updates command helps you identify missing updates for fix categories
that might be applicable to the software instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post(instancePath+"/missingfixcatupdates", nil, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

// instances subgroup
var softwareInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Work with software instances",
	Long: `
DESCRIPTION
-----------
Allows you to list software instances, retrieve their properties,
list the datasets included in a software instance, add a new software
instance, export or delete a defined software instance, and retrieve
the z/OS system UUID.`,
}

var softwareInstancesUUIDCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Retrieve the z/OS system UUID",
	Long: `
DESCRIPTION
-----------
You can use this command to retrieve the UUID for the software instance
that represents the installed software for the specified z/OSMF host system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/swmgmt/system/uuid/%s", nickName)

		resp, err := client.Post(path, nil, nil)
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

var softwareInstancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Obtain a list of software or portable software instances",
	Long: `
DESCRIPTION
-----------
You can use the list command to get a list of all software or
portable software instances defined to z/OSMF.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pswi, _ := cmd.Flags().GetBool("pswi")

		client := Profile.NewZosmfClient()
		path := "/swmgmt/swi"
		if pswi {
			path = "/swmgmt/pswi"
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

var softwareInstancesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a software instance to z/OSMF",
	Long: `
DESCRIPTION
-----------
You can use the add command to add a new software instance to z/OSMF.
The request content must be in a json formatted text file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, _ := cmd.Flags().GetString("file-name")

		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", fileName, err)
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post("/swmgmt/swi", string(data), nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

var softwareInstancesExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a defined software instance",
	Long: `
DESCRIPTION
-----------
You can use this command to perform an Export action on a software instance
that is defined to z/OSMF, which generates a portable software instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, _ := cmd.Flags().GetString("file-name")
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", fileName, err)
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post(instancePath+"/export", string(data), nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

var softwareInstancesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a defined software instance",
	Long: `
DESCRIPTION
-----------
You can use this command to remove a software instance definition from z/OSMF.
The delete operation removes only the definition of the software instance from z/OSMF.
The physical data sets that compose the software instance are not affected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Delete(instancePath, nil)
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

var softwareInstancesSipropsCmd = &cobra.Command{
	Use:   "siprops",
	Short: "Obtain the properties of a defined software instance",
	Long: `
DESCRIPTION
-----------
You can use the siprops command to get the properties of a software instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Get(instancePath, nil)
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

var softwareInstancesSilistdsCmd = &cobra.Command{
	Use:   "silistds",
	Short: "Obtain a list of the data sets that compose a software instance",
	Long: `
DESCRIPTION
-----------
You can use the silistds command to get the datasets of a software instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nickName, _ := cmd.Flags().GetString("nick-name")
		swiName, _ := cmd.Flags().GetString("swi-name")
		uuid, _ := cmd.Flags().GetString("uuid")

		instancePath, err := resolveSwInstancePath("/swmgmt/swi", nickName, swiName, uuid)
		if err != nil {
			return err
		}

		client := Profile.NewZosmfClient()
		resp, err := client.Post(instancePath+"/datasets", nil, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var startResult map[string]interface{}
		if err := json.Unmarshal(resp.Body, &startResult); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		statusURL, _ := startResult["statusurl"].(string)

		finalResp, err := pollAsyncOperation(client, statusURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(8)
		}
		fmt.Println(finalResp.BodyString())
		return nil
	},
}

func init() {
	defaultCSI := config.GetSoftwareProperty("globalCSI")
	if defaultCSI == "" {
		defaultCSI = "SMPE.GLOBAL.CSI"
	}

	// query csids
	softwareQueryCsidsCmd.Flags().String("global-csi", defaultCSI, "The name of the SMP/E Global CSI.")
	softwareQueryCsidsCmd.Flags().StringP("zones", "z", "GLOBAL", "One or more SMP/E Zone names, separated by comma.")
	softwareQueryCsidsCmd.Flags().StringP("entry", "e", "", "SMP/E entry type (sysmod, dddef etc).")
	softwareQueryCsidsCmd.Flags().String("subentries", "", "Blank separated string of subentries.")
	softwareQueryCsidsCmd.Flags().String("filter", "", "Filter criteria for the query.")

	// query critupdates
	softwareQueryCritUpdatesCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareQueryCritUpdatesCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareQueryCritUpdatesCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// query softupdates
	softwareQuerySoftUpdatesCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareQuerySoftUpdatesCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareQuerySoftUpdatesCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// query fixcatupdates
	softwareQueryFixcatUpdatesCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareQueryFixcatUpdatesCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareQueryFixcatUpdatesCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// instances uuid
	softwareInstancesUUIDCmd.Flags().String("nick-name", "", "The system nick name.")
	softwareInstancesUUIDCmd.MarkFlagRequired("nick-name")

	// instances list
	softwareInstancesListCmd.Flags().Bool("pswi", false, "List portable software instances.")

	// instances add
	softwareInstancesAddCmd.Flags().String("file-name", "", "File name with the instance input data.")
	softwareInstancesAddCmd.MarkFlagRequired("file-name")

	// instances export
	softwareInstancesExportCmd.Flags().String("file-name", "", "File Name of the export input file.")
	softwareInstancesExportCmd.MarkFlagRequired("file-name")
	softwareInstancesExportCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareInstancesExportCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareInstancesExportCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// instances delete
	softwareInstancesDeleteCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareInstancesDeleteCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareInstancesDeleteCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// instances siprops
	softwareInstancesSipropsCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareInstancesSipropsCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareInstancesSipropsCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// instances silistds
	softwareInstancesSilistdsCmd.Flags().String("nick-name", "", "The system nick name of the software instance.")
	softwareInstancesSilistdsCmd.Flags().String("swi-name", "", "The name of the software instance.")
	softwareInstancesSilistdsCmd.Flags().StringP("uuid", "u", "", "The uuid representing the software instance.")

	// Wire up query subgroup
	softwareQueryCmd.AddCommand(softwareQueryCsidsCmd)
	softwareQueryCmd.AddCommand(softwareQueryCritUpdatesCmd)
	softwareQueryCmd.AddCommand(softwareQuerySoftUpdatesCmd)
	softwareQueryCmd.AddCommand(softwareQueryFixcatUpdatesCmd)

	// Wire up instances subgroup
	softwareInstancesCmd.AddCommand(softwareInstancesUUIDCmd)
	softwareInstancesCmd.AddCommand(softwareInstancesListCmd)
	softwareInstancesCmd.AddCommand(softwareInstancesAddCmd)
	softwareInstancesCmd.AddCommand(softwareInstancesExportCmd)
	softwareInstancesCmd.AddCommand(softwareInstancesDeleteCmd)
	softwareInstancesCmd.AddCommand(softwareInstancesSipropsCmd)
	softwareInstancesCmd.AddCommand(softwareInstancesSilistdsCmd)

	// Wire up software
	softwareCmd.AddCommand(softwareQueryCmd)
	softwareCmd.AddCommand(softwareInstancesCmd)

	rootCmd.AddCommand(softwareCmd)
}
