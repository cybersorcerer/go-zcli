package cmd

import (
	"fmt"
	"os"
	"strings"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var datasetsCmd = &cobra.Command{
	Use:   "datasets",
	Short: "Interact with z/OS datasets",
	Long: `
DESCRIPTION
-----------
Interact with z/OS datasets.`,
}

var datasetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List z/OS datasets",
	Long: `
DESCRIPTION
-----------
You can use this command to obtain a list of z/OS datasets.
You can search the z/OS catalog or a z/OS volume serial for
matching datasets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsnLevel, _ := cmd.Flags().GetString("dsn-level")
		volser, _ := cmd.Flags().GetString("volser")
		start, _ := cmd.Flags().GetString("start")

		if dsnLevel == "" {
			dsnLevel = fmt.Sprintf("%s.**", strings.ToUpper(Profile.User))
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds?dslevel=%s", dsnLevel)
		if volser != "" {
			path += fmt.Sprintf("&volser=%s", volser)
		}
		if start != "" {
			path += fmt.Sprintf("&start=%s", start)
		}

		headers := map[string]string{
			"X-IBM-Attributes": "base,total",
			"X-IBM-Max-Items":  "0",
		}

		resp, err := client.Get(path, headers)
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

var datasetsMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "List members of z/OS partitioned datasets",
	Long: `
DESCRIPTION
-----------
You can use this command to obtain a list of members of z/OS
partitioned datasets (PDS and PDS/E).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		pattern, _ := cmd.Flags().GetString("pattern")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s/member", dsName)
		if pattern != "" {
			path += fmt.Sprintf("&pattern=%s", pattern)
		}

		headers := map[string]string{
			"X-IBM-Attributes": "base,total",
			"X-IBM-Max-Items":  "0",
		}

		resp, err := client.Get(path, headers)
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

var datasetsReadCmd = &cobra.Command{
	Use:   "read",
	Short: "Read a member of a PDS/PDS(E) or a sequential dataset",
	Long: `
DESCRIPTION
-----------
You can use this command to read a member of a dataset or a
sequential dataset. If the dataset is not cataloged specify a
volume serial number to read the member or the dataset directly
from the volume.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		memberName, _ := cmd.Flags().GetString("member-name")
		volser, _ := cmd.Flags().GetString("volser")
		encoding, _ := cmd.Flags().GetString("encoding")
		enqExclusive, _ := cmd.Flags().GetBool("enq-exclusive")

		client := Profile.NewZosmfClient()
		path := "/restfiles/ds"
		if volser != "" {
			path += fmt.Sprintf("/-(%s)", volser)
		}
		if dsName != "" {
			path += fmt.Sprintf("/%s", dsName)
		}
		if memberName != "" {
			path += fmt.Sprintf("(%s)", memberName)
		}

		headers := map[string]string{
			"X-IBM-Data-Type":   "text",
			"X-IBM-Obtain-ENQ":  "SHRW",
			"X-IBM-Return-Etag": "true",
		}
		if enqExclusive {
			headers["X-IBM-Obtain-ENQ"] = "EXCLU"
		}
		if encoding != "" {
			headers["X-IBM-Dsname-Encoding"] = encoding
		}

		resp, err := client.Get(path, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		fmt.Println(resp.Headers)
		return nil
	},
}

var datasetsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create sequential and partitioned datasets",
	Long: `
DESCRIPTION
-----------
You can use this command to create a sequential or partitioned
dataset on a z/OS system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		volser, _ := cmd.Flags().GetString("volser")
		unit, _ := cmd.Flags().GetString("unit")
		dsorg, _ := cmd.Flags().GetString("dsorg")
		alcunit, _ := cmd.Flags().GetString("alcunit")
		primary, _ := cmd.Flags().GetInt("primary")
		secondary, _ := cmd.Flags().GetInt("secondary")
		dirblk, _ := cmd.Flags().GetInt("dirblk")
		avgblk, _ := cmd.Flags().GetInt("avgblk")
		recfm, _ := cmd.Flags().GetString("recfm")
		blksize, _ := cmd.Flags().GetInt("blksize")
		lrecl, _ := cmd.Flags().GetInt("lrecl")
		storclass, _ := cmd.Flags().GetString("storclass")
		mgntclass, _ := cmd.Flags().GetString("mgntclass")
		dataclass, _ := cmd.Flags().GetString("dataclass")
		dsnType, _ := cmd.Flags().GetString("dsn-type")
		like, _ := cmd.Flags().GetString("like")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := map[string]interface{}{
			"volser":    volser,
			"unit":      unit,
			"dsorg":     dsorg,
			"alcunit":   alcunit,
			"primary":   primary,
			"secondary": secondary,
			"dirblk":    dirblk,
			"avgblk":    avgblk,
			"recfm":     recfm,
			"blksize":   blksize,
			"lrecl":     lrecl,
			"storclass": storclass,
			"mgntclass": mgntclass,
			"dataclass": dataclass,
			"dsntype":   dsnType,
			"like":      like,
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

var datasetsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete sequential/partitioned z/OS datasets or members",
	Long: `
DESCRIPTION
-----------
You can use this command to delete a sequential or partitioned
dataset or delete members on a partitioned dataset. If you would
like to delete non cataloged datasets or members, you can specify
a z/OS volume serial.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		memberName, _ := cmd.Flags().GetString("member-name")
		volser, _ := cmd.Flags().GetString("volser")

		client := Profile.NewZosmfClient()
		path := "/restfiles/ds/"
		if volser != "" {
			path += fmt.Sprintf("-(%s)/", volser)
		}
		path += dsName
		if memberName != "" {
			path += fmt.Sprintf("/%s", memberName)
		}

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

// utilities subgroup
var datasetsUtilitiesCmd = &cobra.Command{
	Use:   "utilities",
	Short: "Copy, rename, hmigrate, hrecall or hdelete z/OS datasets",
	Long: `
DESCRIPTION
-----------
You can use the z/OS data set and member utilities to work with
data sets and members. The available commands allow you to rename
members or datasets, migrate data sets, recall migrated data sets,
and delete backup versions of data sets.`,
}

var datasetsHrecallCmd = &cobra.Command{
	Use:   "hrecall",
	Short: "Recall a migrated z/OS dataset",
	Long: `
DESCRIPTION
-----------
You can use this command to recall a previously migrated sequential
or partitioned dataset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		wait, _ := cmd.Flags().GetBool("wait")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := fmt.Sprintf(`{"request": "hrecall", "wait": "%t"}`, wait)

		resp, err := client.Put(path, payload, nil)
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

var datasetsHmigrateCmd = &cobra.Command{
	Use:   "hmigrate",
	Short: "Migrate a z/OS dataset",
	Long: `
DESCRIPTION
-----------
Migrates a data set to a DFSMShsm level 1 or level 2 volume.
Performed in the foreground by DFSMShsm.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		wait, _ := cmd.Flags().GetBool("wait")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := fmt.Sprintf(`{"request": "hmigrate", "wait": "%t"}`, wait)

		resp, err := client.Put(path, payload, nil)
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

var datasetsHdeleteCmd = &cobra.Command{
	Use:   "hdelete",
	Short: "Delete a migrated z/OS dataset",
	Long: `
DESCRIPTION
-----------
Scratches a data set on a DASD migration volume without recalling the data set
or marks the data set not valid in the TTOC for any tape ML2 volumes that contain
the data set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		wait, _ := cmd.Flags().GetBool("wait")
		purge, _ := cmd.Flags().GetBool("purge")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := fmt.Sprintf(`{"request": "hdelete", "wait": "%t", "purge": %t}`, wait, purge)

		resp, err := client.Put(path, payload, nil)
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

var datasetsRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename z/OS datasets or members in a z/OS PDS(E)",
	Long: `
DESCRIPTION
-----------
You can use this command to rename a member of a dataset or a
sequential dataset.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		toDsName, _ := cmd.Flags().GetString("to-ds-name")
		memberName, _ := cmd.Flags().GetString("member-name")
		toMemberName, _ := cmd.Flags().GetString("to-member-name")
		enqExclusive, _ := cmd.Flags().GetBool("enq-exclusive")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)

		payload := map[string]interface{}{
			"request": "rename",
		}
		if toDsName != "" {
			payload["to-dsname"] = toDsName
		}
		if memberName != "" {
			payload["from-member"] = memberName
		}
		if toMemberName != "" {
			payload["to-member"] = toMemberName
		}
		if enqExclusive {
			payload["enq"] = "exclu"
		}

		resp, err := client.Put(path, payload, nil)
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
	// datasets list
	datasetsListCmd.Flags().StringP("dsn-level", "d", "", "A dataset level to list. DEFAULT <USERID>.**")
	datasetsListCmd.Flags().StringP("volser", "v", "", "A volume serial number.")
	datasetsListCmd.Flags().StringP("start", "s", "", "A dataset level to list.")

	// datasets members
	datasetsMembersCmd.Flags().StringP("ds-name", "n", "", "The dataset name of a z/OS PDS or PDS/E.")
	datasetsMembersCmd.MarkFlagRequired("ds-name")
	datasetsMembersCmd.Flags().StringP("pattern", "p", "", "A search pattern following the ISPF LMMLIST.")

	// datasets read
	datasetsReadCmd.Flags().StringP("ds-name", "n", "", "The dataset name of a z/OS PDS or PDS/E or sequential dataset.")
	datasetsReadCmd.MarkFlagRequired("ds-name")
	datasetsReadCmd.Flags().StringP("member-name", "", "", "A member name if --ds-name is a PDS or PDS/E.")
	datasetsReadCmd.Flags().StringP("volser", "v", "", "A volume serial number if --ds-name is not cataloged.")
	datasetsReadCmd.Flags().StringP("encoding", "e", "", "Encoding of the data.")
	datasetsReadCmd.Flags().Bool("enq-exclusive", false, "If true an exclusive enq will be set.")

	// datasets create
	datasetsCreateCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsCreateCmd.MarkFlagRequired("ds-name")
	datasetsCreateCmd.Flags().StringP("volser", "v", "", "A volume serial number.")
	datasetsCreateCmd.Flags().StringP("unit", "u", "", "z/OS Device type.")
	datasetsCreateCmd.Flags().String("dsorg", "PS", "Data set organization (PS/PO).")
	datasetsCreateCmd.Flags().String("alcunit", "TRK", "Unit of space allocation (TRK/CYL/M/G).")
	datasetsCreateCmd.Flags().IntP("primary", "p", 10, "Primary space allocation.")
	datasetsCreateCmd.Flags().IntP("secondary", "s", 5, "Secondary space allocation.")
	datasetsCreateCmd.Flags().Int("dirblk", 5, "Number of directory blocks.")
	datasetsCreateCmd.Flags().IntP("avgblk", "a", 0, "Average block size.")
	datasetsCreateCmd.Flags().StringP("recfm", "r", "FB", "Record format.")
	datasetsCreateCmd.Flags().Int("blksize", 0, "Block size.")
	datasetsCreateCmd.Flags().IntP("lrecl", "l", 80, "Record length.")
	datasetsCreateCmd.Flags().String("storclass", "", "Storage class.")
	datasetsCreateCmd.Flags().String("mgntclass", "", "Management class.")
	datasetsCreateCmd.Flags().String("dataclass", "", "Data class.")
	datasetsCreateCmd.Flags().String("dsn-type", "", "Data set type.")
	datasetsCreateCmd.Flags().String("like", "", "Model data set name.")

	// datasets delete
	datasetsDeleteCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsDeleteCmd.MarkFlagRequired("ds-name")
	datasetsDeleteCmd.Flags().StringP("member-name", "", "", "Dataset member to delete.")
	datasetsDeleteCmd.Flags().StringP("volser", "v", "", "A volume serial number.")

	// utilities: hrecall
	datasetsHrecallCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsHrecallCmd.MarkFlagRequired("ds-name")
	datasetsHrecallCmd.Flags().Bool("wait", false, "If true, wait for completion of the request.")

	// utilities: hmigrate
	datasetsHmigrateCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsHmigrateCmd.MarkFlagRequired("ds-name")
	datasetsHmigrateCmd.Flags().Bool("wait", false, "If true, wait for completion of the request.")

	// utilities: hdelete
	datasetsHdeleteCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsHdeleteCmd.MarkFlagRequired("ds-name")
	datasetsHdeleteCmd.Flags().Bool("wait", false, "If true, wait for completion of the request.")
	datasetsHdeleteCmd.Flags().Bool("purge", false, "If true uses PURGE=YES on ARCHDEL request.")

	// utilities: rename
	datasetsRenameCmd.Flags().StringP("ds-name", "n", "", "Dataset name.")
	datasetsRenameCmd.MarkFlagRequired("ds-name")
	datasetsRenameCmd.Flags().String("to-ds-name", "", "New dataset name for a Dataset rename.")
	datasetsRenameCmd.Flags().String("member-name", "", "Member name on --ds-name.")
	datasetsRenameCmd.Flags().String("to-member-name", "", "New member name on --ds-name.")
	datasetsRenameCmd.Flags().Bool("enq-exclusive", false, "If true a shared enq will be set.")

	// Wire up command tree
	datasetsUtilitiesCmd.AddCommand(datasetsHrecallCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsHmigrateCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsHdeleteCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsRenameCmd)

	datasetsCmd.AddCommand(datasetsListCmd)
	datasetsCmd.AddCommand(datasetsMembersCmd)
	datasetsCmd.AddCommand(datasetsReadCmd)
	datasetsCmd.AddCommand(datasetsCreateCmd)
	datasetsCmd.AddCommand(datasetsDeleteCmd)
	datasetsCmd.AddCommand(datasetsUtilitiesCmd)

	rootCmd.AddCommand(datasetsCmd)
}
