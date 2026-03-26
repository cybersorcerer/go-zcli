package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var filesystemsCmd = &cobra.Command{
	Use:   "filesystems",
	Short: "Interact with z/OS z/Unix filesystems",
	Long: `
DESCRIPTION
-----------
Interact with z/OS z/Unix filesystems.`,
}

var filesystemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a zfs z/UNIX filesystem",
	Long: `
DESCRIPTION
-----------
The command will create a new zfs filesystem.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zfsName, _ := cmd.Flags().GetString("zfs-dataset-name")
		owner, _ := cmd.Flags().GetString("owner")
		group, _ := cmd.Flags().GetString("group")
		storageClass, _ := cmd.Flags().GetString("storage-class")
		managementClass, _ := cmd.Flags().GetString("management-class")
		dataClass, _ := cmd.Flags().GetString("data-class")
		cylsPri, _ := cmd.Flags().GetInt("primary-cylinders")
		cylsSec, _ := cmd.Flags().GetInt("secondary-cylinders")
		volumes, _ := cmd.Flags().GetStringSlice("volumes")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/zfs/%s", zfsName)

		payload := map[string]interface{}{
			"zfsDatasetName": zfsName,
			"cylsPri":        cylsPri,
			"cylsSec":        cylsSec,
		}
		if owner != "" {
			payload["owner"] = owner
		}
		if group != "" {
			payload["group"] = group
		}
		if storageClass != "" {
			payload["storageClass"] = storageClass
		}
		if managementClass != "" {
			payload["managementClass"] = managementClass
		}
		if dataClass != "" {
			payload["dataClass"] = dataClass
		}
		if len(volumes) > 0 {
			payload["volumes"] = volumes
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 201); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

var filesystemsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a zfs z/UNIX filesystem",
	Long: `
DESCRIPTION
-----------
The command will delete a zfs filesystem.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zfsName, _ := cmd.Flags().GetString("zfs-dataset-name")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/zfs/%s", zfsName)

		resp, err := client.Delete(path, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

var filesystemsMountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount a z/UNIX filesystem",
	Long: `
DESCRIPTION
-----------
You can use this command to mount a z/OS UNIX file system on a specified directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fsName, _ := cmd.Flags().GetString("fs-dataset-name")
		fsType, _ := cmd.Flags().GetString("fs-type")
		mountPoint, _ := cmd.Flags().GetString("mount-point")
		mode, _ := cmd.Flags().GetString("mode")
		setuid, _ := cmd.Flags().GetBool("setuid")

		modeStr := mode
		if setuid {
			modeStr += " setuid"
		} else {
			modeStr += " nosetuid"
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/%s", fsName)
		payload := map[string]interface{}{
			"action":      "mount",
			"mount-point": mountPoint,
			"fs-type":     fsType,
			"mode":        modeStr,
		}

		resp, err := client.Put(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

var filesystemsUnmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount a z/UNIX filesystem",
	Long: `
DESCRIPTION
-----------
You can use this command to unmount a z/OS UNIX file system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fsName, _ := cmd.Flags().GetString("fs-dataset-name")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/%s", fsName)
		payload := map[string]interface{}{
			"action": "unmount",
		}

		resp, err := client.Put(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

var filesystemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List z/UNIX filesystem(s)",
	Long: `
DESCRIPTION
-----------
You can use the list z/OS UNIX filesystems command to list all mounted filesystems,
or the specific filesystem mounted at a given path, or the filesystem with a given
Filesystem name. If no options are provided, the command will list all mounted filesystems.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fsName, _ := cmd.Flags().GetString("fs-dataset-name")
		pathName, _ := cmd.Flags().GetString("path")

		client := Profile.NewZosmfClient()
		var path string
		if fsName != "" {
			path = fmt.Sprintf("/restfiles/mfs/?fsname=%s", fsName)
		} else if pathName != "" {
			path = fmt.Sprintf("/restfiles/mfs/?path=%s", pathName)
		} else {
			path = "/restfiles/mfs"
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
	// filesystems create
	filesystemsCreateCmd.Flags().StringP("zfs-dataset-name", "z", "", "The zfs dataset name.")
	filesystemsCreateCmd.MarkFlagRequired("zfs-dataset-name")
	filesystemsCreateCmd.Flags().StringP("owner", "o", "", "Owner User ID of the new zfs.")
	filesystemsCreateCmd.Flags().StringP("group", "g", "", "The group Owner of the new zfs.")
	filesystemsCreateCmd.Flags().String("storage-class", "", "z/OS DFSMS storage class.")
	filesystemsCreateCmd.Flags().String("management-class", "", "z/OS DFSMS management class.")
	filesystemsCreateCmd.Flags().String("data-class", "", "z/OS DFSMS data class.")
	filesystemsCreateCmd.Flags().Int("primary-cylinders", 0, "Primary space request in cylinders.")
	filesystemsCreateCmd.MarkFlagRequired("primary-cylinders")
	filesystemsCreateCmd.Flags().Int("secondary-cylinders", 0, "Secondary space request in cylinders.")
	filesystemsCreateCmd.MarkFlagRequired("secondary-cylinders")
	filesystemsCreateCmd.Flags().StringSlice("volumes", nil, "Volume serial numbers.")

	// filesystems delete
	filesystemsDeleteCmd.Flags().StringP("zfs-dataset-name", "z", "", "The zfs dataset name.")
	filesystemsDeleteCmd.MarkFlagRequired("zfs-dataset-name")

	// filesystems mount
	filesystemsMountCmd.Flags().StringP("fs-dataset-name", "f", "", "The filesystem dataset name.")
	filesystemsMountCmd.MarkFlagRequired("fs-dataset-name")
	filesystemsMountCmd.Flags().String("fs-type", "zfs", "The file system type.")
	filesystemsMountCmd.Flags().StringP("mount-point", "t", "", "The mount point.")
	filesystemsMountCmd.MarkFlagRequired("mount-point")
	filesystemsMountCmd.Flags().StringP("mode", "m", "rdonly", "Mode of mount operation.")
	filesystemsMountCmd.Flags().Bool("setuid", false, "If true uses the setuid option in the mount mode.")

	// filesystems unmount
	filesystemsUnmountCmd.Flags().StringP("fs-dataset-name", "f", "", "The filesystem dataset name.")
	filesystemsUnmountCmd.MarkFlagRequired("fs-dataset-name")

	// filesystems list
	filesystemsListCmd.Flags().StringP("fs-dataset-name", "f", "", "The filesystem dataset name.")
	filesystemsListCmd.Flags().StringP("path", "p", "", "The mount path.")

	// Wire up
	filesystemsCmd.AddCommand(filesystemsCreateCmd)
	filesystemsCmd.AddCommand(filesystemsDeleteCmd)
	filesystemsCmd.AddCommand(filesystemsMountCmd)
	filesystemsCmd.AddCommand(filesystemsUnmountCmd)
	filesystemsCmd.AddCommand(filesystemsListCmd)

	rootCmd.AddCommand(filesystemsCmd)
}
