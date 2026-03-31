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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsnLevel, _ := cmd.Flags().GetString("dsn-level")
		volser, _ := cmd.Flags().GetString("volser")
		start, _ := cmd.Flags().GetString("start")
		attributes, _ := cmd.Flags().GetString("attributes")
		maxItems, _ := cmd.Flags().GetString("max-items")

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
			"X-IBM-Attributes": attributes,
			"X-IBM-Max-Items":  maxItems,
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		pattern, _ := cmd.Flags().GetString("pattern")
		start, _ := cmd.Flags().GetString("start")
		attributes, _ := cmd.Flags().GetString("attributes")
		maxItems, _ := cmd.Flags().GetString("max-items")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s/member", dsName)

		var params []string
		if start != "" {
			params = append(params, "start="+start)
		}
		if pattern != "" {
			params = append(params, "pattern="+pattern)
		}
		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		headers := map[string]string{
			"X-IBM-Attributes": attributes,
			"X-IBM-Max-Items":  maxItems,
		}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		memberName, _ := cmd.Flags().GetString("member-name")
		volser, _ := cmd.Flags().GetString("volser")
		encoding, _ := cmd.Flags().GetString("encoding")
		dataType, _ := cmd.Flags().GetString("data-type")
		fileEncoding, _ := cmd.Flags().GetString("file-encoding")
		search, _ := cmd.Flags().GetString("search")
		research, _ := cmd.Flags().GetString("research")
		insensitive, _ := cmd.Flags().GetString("insensitive")
		maxReturnSize, _ := cmd.Flags().GetString("max-return-size")
		recordRange, _ := cmd.Flags().GetString("record-range")
		obtainEnq, _ := cmd.Flags().GetString("obtain-enq")
		sessionRef, _ := cmd.Flags().GetString("session-ref")
		releaseEnq, _ := cmd.Flags().GetBool("release-enq")
		returnEtag, _ := cmd.Flags().GetBool("return-etag")
		ifNoneMatch, _ := cmd.Flags().GetString("if-none-match")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")

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

		// Query parameters
		var params []string
		if search != "" {
			params = append(params, "search="+search)
		}
		if research != "" {
			params = append(params, "research="+research)
		}
		if insensitive != "" {
			params = append(params, "insensitive="+insensitive)
		}
		if maxReturnSize != "" {
			params = append(params, "maxreturnsize="+maxReturnSize)
		}
		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		// Headers
		dataTypeHeader := dataType
		if dataType == "text" && fileEncoding != "" {
			dataTypeHeader = fmt.Sprintf("text;fileEncoding=%s", fileEncoding)
		}
		headers := map[string]string{
			"X-IBM-Data-Type": dataTypeHeader,
		}
		if returnEtag {
			headers["X-IBM-Return-Etag"] = "true"
		}
		if obtainEnq != "" {
			headers["X-IBM-Obtain-ENQ"] = obtainEnq
		}
		if sessionRef != "" {
			headers["X-IBM-Session-Ref"] = sessionRef
		}
		if releaseEnq {
			headers["X-IBM-Release-ENQ"] = "true"
		}
		if encoding != "" {
			headers["X-IBM-Dsname-Encoding"] = encoding
		}
		if recordRange != "" {
			headers["X-IBM-Record-Range"] = recordRange
		}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}
		if ifNoneMatch != "" {
			headers["If-None-Match"] = ifNoneMatch
		}

		resp, err := client.Get(path, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 304); apiErr != nil {
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
	Args: cobra.NoArgs,
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
		dataType, _ := cmd.Flags().GetString("data-type")
		fileEncoding, _ := cmd.Flags().GetString("file-encoding")
		crlf, _ := cmd.Flags().GetBool("crlf")
		wrap, _ := cmd.Flags().GetBool("wrap")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")
		obtainEnq, _ := cmd.Flags().GetString("obtain-enq")
		sessionRef, _ := cmd.Flags().GetString("session-ref")
		releaseEnq, _ := cmd.Flags().GetBool("release-enq")
		encoding, _ := cmd.Flags().GetString("encoding")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := map[string]interface{}{
			"dsorg":     dsorg,
			"alcunit":   alcunit,
			"primary":   primary,
			"secondary": secondary,
			"recfm":     recfm,
			"lrecl":     lrecl,
		}
		if volser != "" {
			payload["volser"] = volser
		}
		if unit != "" {
			payload["unit"] = unit
		}
		if dirblk != 0 {
			payload["dirblk"] = dirblk
		}
		if avgblk != 0 {
			payload["avgblk"] = avgblk
		}
		if blksize != 0 {
			payload["blksize"] = blksize
		}
		if storclass != "" {
			payload["storclass"] = storclass
		}
		if mgntclass != "" {
			payload["mgntclass"] = mgntclass
		}
		if dataclass != "" {
			payload["dataclass"] = dataclass
		}
		if dsnType != "" {
			payload["dsntype"] = dsnType
		}
		if like != "" {
			payload["like"] = like
		}

		// Build X-IBM-Data-Type header
		dataTypeHeader := dataType
		if dataType == "text" {
			var opts []string
			if fileEncoding != "" {
				opts = append(opts, "fileEncoding="+fileEncoding)
			}
			if crlf {
				opts = append(opts, "crlf=true")
			}
			if wrap {
				opts = append(opts, "wrap=true")
			}
			if len(opts) > 0 {
				dataTypeHeader = "text;" + strings.Join(opts, ";")
			}
		}

		headers := map[string]string{
			"X-IBM-Data-Type": dataTypeHeader,
		}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}
		if obtainEnq != "" {
			headers["X-IBM-Obtain-ENQ"] = obtainEnq
		}
		if sessionRef != "" {
			headers["X-IBM-Session-Ref"] = sessionRef
		}
		if releaseEnq {
			headers["X-IBM-Release-ENQ"] = "true"
		}
		if encoding != "" {
			headers["X-IBM-Dsname-Encoding"] = encoding
		}

		resp, err := client.Post(path, payload, headers)
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
	Args: cobra.NoArgs,
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
	Short: "Copy, rename, hmigrate, hrecall, hdelete or ams z/OS datasets",
	Long: `
DESCRIPTION
-----------
You can use the z/OS data set and member utilities to work with
data sets and members. The available commands allow you to rename
members or datasets, copy datasets or members, migrate data sets,
recall migrated data sets, delete backup versions of data sets,
and invoke IDCAMS Access Method Services (AMS).`,
}

var datasetsHrecallCmd = &cobra.Command{
	Use:   "hrecall",
	Short: "Recall a migrated z/OS dataset",
	Long: `
DESCRIPTION
-----------
You can use this command to recall a previously migrated sequential
or partitioned dataset.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		wait, _ := cmd.Flags().GetBool("wait")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := map[string]interface{}{
			"request": "hrecall",
			"wait":    wait,
		}

		headers := map[string]string{}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}

		resp, err := client.Put(path, payload, headers)
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		wait, _ := cmd.Flags().GetBool("wait")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := map[string]interface{}{
			"request": "hmigrate",
			"wait":    wait,
		}

		headers := map[string]string{}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}

		resp, err := client.Put(path, payload, headers)
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		wait, _ := cmd.Flags().GetBool("wait")
		purge, _ := cmd.Flags().GetBool("purge")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		payload := map[string]interface{}{
			"request": "hdelete",
			"wait":    wait,
			"purge":   purge,
		}

		headers := map[string]string{}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}

		resp, err := client.Put(path, payload, headers)
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
You can use this command to rename a dataset or a member of a
partitioned dataset. The --ds-name is the NEW (target) name.
Use --from-ds-name for the source dataset.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		fromDsName, _ := cmd.Flags().GetString("from-ds-name")
		memberName, _ := cmd.Flags().GetString("member-name")
		fromMember, _ := cmd.Flags().GetString("from-member")
		enq, _ := cmd.Flags().GetString("enq")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/ds/%s", dsName)
		if memberName != "" {
			path += fmt.Sprintf("(%s)", memberName)
		}

		fromDataset := map[string]interface{}{
			"dsn": fromDsName,
		}
		if fromMember != "" {
			fromDataset["member"] = fromMember
		}

		payload := map[string]interface{}{
			"request":      "rename",
			"from-dataset": fromDataset,
		}
		if enq != "" {
			payload["enq"] = enq
		}

		headers := map[string]string{}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}

		resp, err := client.Put(path, payload, headers)
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

var datasetsCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy z/OS datasets, members, or UNIX files",
	Long: `
DESCRIPTION
-----------
You can use this command to copy a dataset, member, or z/OS UNIX
file to a target dataset or member. The --ds-name is the target
dataset. Use --from-ds-name or --from-file to specify the source.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsName, _ := cmd.Flags().GetString("ds-name")
		memberName, _ := cmd.Flags().GetString("member-name")
		volser, _ := cmd.Flags().GetString("volser")
		fromDsName, _ := cmd.Flags().GetString("from-ds-name")
		fromMember, _ := cmd.Flags().GetString("from-member")
		fromVolser, _ := cmd.Flags().GetString("from-volser")
		fromAlias, _ := cmd.Flags().GetBool("from-alias")
		fromFile, _ := cmd.Flags().GetString("from-file")
		fromFileType, _ := cmd.Flags().GetString("from-file-type")
		enq, _ := cmd.Flags().GetString("enq")
		replace, _ := cmd.Flags().GetBool("replace")
		migratedRecall, _ := cmd.Flags().GetString("migrated-recall")
		bpxkAutocvt, _ := cmd.Flags().GetString("bpxk-autocvt")

		client := Profile.NewZosmfClient()
		path := "/restfiles/ds"
		if volser != "" {
			path += fmt.Sprintf("/-(%s)", volser)
		}
		path += fmt.Sprintf("/%s", dsName)
		if memberName != "" {
			path += fmt.Sprintf("(%s)", memberName)
		}

		payload := map[string]interface{}{
			"request": "copy",
			"replace": replace,
		}

		if fromFile != "" {
			fromFileObj := map[string]interface{}{
				"filename": fromFile,
			}
			if fromFileType != "" {
				fromFileObj["type"] = fromFileType
			}
			payload["from-file"] = fromFileObj
		} else {
			fromDataset := map[string]interface{}{
				"dsn": fromDsName,
			}
			if fromMember != "" {
				fromDataset["member"] = fromMember
			}
			if fromVolser != "" {
				fromDataset["volser"] = fromVolser
			}
			if fromAlias {
				fromDataset["alias"] = true
			}
			payload["from-dataset"] = fromDataset
		}

		if enq != "" {
			payload["enq"] = enq
		}

		headers := map[string]string{}
		if migratedRecall != "" {
			headers["X-IBM-Migrated-Recall"] = migratedRecall
		}
		if bpxkAutocvt != "" {
			headers["X-IBM-BPXK-AUTOCVT"] = bpxkAutocvt
		}

		resp, err := client.Put(path, payload, headers)
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

var datasetsAmsCmd = &cobra.Command{
	Use:   "ams",
	Short: "Invoke IDCAMS Access Method Services",
	Long: `
DESCRIPTION
-----------
You can use this command to invoke IDCAMS Access Method Services.
Provide one or more IDCAMS input statements via the --input flag.
Each input line must be <= 255 characters. The total size of all
input lines plus the number of lines must be <= 8K.

Example:
  zcli datasets utilities ams \
    --input "DEFINE CLUSTER(NAME (MY.KSDS) VOLUMES(VSER05)) -" \
    --input "DATA  (KILOBYTES (50 5))"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetStringArray("input")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := "/restfiles/ams"

		payload := map[string]interface{}{
			"input":       input,
			"JSONversion": 1,
		}

		headers := map[string]string{}
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" && targetPassword != "" {
			headers["X-IBM-Target-System-User"] = targetUser
			headers["X-IBM-Target-System-Password"] = targetPassword
		}

		resp, err := client.Put(path, payload, headers)
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
	datasetsListCmd.Flags().String("dsn-level", "", "A dataset level to list. DEFAULT <USERID>.**")
	datasetsListCmd.Flags().StringP("volser", "v", "", "A volume serial number.")
	datasetsListCmd.Flags().StringP("start", "s", "", "First dataset name to return in the response.")
	datasetsListCmd.Flags().String("attributes", "base,total", "Attributes to return: dsname, base, vol (with optional ,total suffix).")
	datasetsListCmd.Flags().String("max-items", "0", "Maximum number of items to return (0 = all).")

	// datasets members
	datasetsMembersCmd.Flags().StringP("ds-name", "n", "", "The dataset name of a z/OS PDS or PDS/E.")
	datasetsMembersCmd.MarkFlagRequired("ds-name")
	datasetsMembersCmd.Flags().StringP("pattern", "p", "", "A search pattern following the ISPF LMMLIST syntax.")
	datasetsMembersCmd.Flags().StringP("start", "s", "", "First member name to return in the response (max 8 chars).")
	datasetsMembersCmd.Flags().String("attributes", "member", "Attributes to return: member or base (with optional ,total suffix).")
	datasetsMembersCmd.Flags().String("max-items", "0", "Maximum number of items to return (0 = all).")
	datasetsMembersCmd.Flags().String("migrated-recall", "", "How to handle migrated datasets: wait, nowait, or error.")

	// datasets read
	datasetsReadCmd.Flags().StringP("ds-name", "n", "", "The dataset name of a z/OS PDS or PDS/E or sequential dataset.")
	datasetsReadCmd.MarkFlagRequired("ds-name")
	datasetsReadCmd.Flags().String("member-name", "", "A member name if --ds-name is a PDS or PDS/E.")
	datasetsReadCmd.Flags().StringP("volser", "v", "", "A volume serial number if --ds-name is not cataloged.")
	datasetsReadCmd.Flags().StringP("encoding", "e", "", "Dataset/member name codepage (X-IBM-Dsname-Encoding).")
	// Query parameters
	datasetsReadCmd.Flags().String("search", "", "Search for the first record containing this string.")
	datasetsReadCmd.Flags().String("research", "", "Search using an extended regular expression.")
	datasetsReadCmd.Flags().String("insensitive", "", "Case insensitive search: true (default) or false.")
	datasetsReadCmd.Flags().String("max-return-size", "", "Max records to return for search/research (default 100).")
	// Custom headers
	datasetsReadCmd.Flags().String("data-type", "text", "Data type: text, binary, or record.")
	datasetsReadCmd.Flags().String("file-encoding", "", "Alternate EBCDIC code page for text mode (default IBM-1047).")
	datasetsReadCmd.Flags().String("record-range", "", "Record range: SSS-EEE or SSS,NNN (0-based).")
	datasetsReadCmd.Flags().String("obtain-enq", "", "Obtain ENQ: EXCLU or SHRW.")
	datasetsReadCmd.Flags().String("session-ref", "", "Session reference from a previous obtain-enq response.")
	datasetsReadCmd.Flags().Bool("release-enq", false, "Release ENQ held by the session-ref address space.")
	datasetsReadCmd.Flags().Bool("return-etag", false, "Force ETag in response regardless of data size.")
	datasetsReadCmd.Flags().String("if-none-match", "", "ETag token for conditional retrieval (HTTP 304 if unchanged).")
	datasetsReadCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")

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
	// Custom headers
	datasetsCreateCmd.Flags().String("data-type", "text", "Data type: text, binary, or record.")
	datasetsCreateCmd.Flags().String("file-encoding", "", "Alternate EBCDIC code page (default IBM-1047).")
	datasetsCreateCmd.Flags().Bool("crlf", false, "Use CRLF line terminators instead of LF (text mode only).")
	datasetsCreateCmd.Flags().Bool("wrap", false, "Wrap data for F/FB datasets to avoid truncation (text mode only).")
	datasetsCreateCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")
	datasetsCreateCmd.Flags().String("obtain-enq", "", "Obtain ENQ: EXCLU or SHRW.")
	datasetsCreateCmd.Flags().String("session-ref", "", "Session reference from a previous obtain-enq response.")
	datasetsCreateCmd.Flags().Bool("release-enq", false, "Release ENQ held by the session-ref address space.")
	datasetsCreateCmd.Flags().StringP("encoding", "e", "", "Dataset/member name codepage (X-IBM-Dsname-Encoding).")

	// datasets delete
	datasetsDeleteCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsDeleteCmd.MarkFlagRequired("ds-name")
	datasetsDeleteCmd.Flags().StringP("member-name", "", "", "Dataset member to delete.")
	datasetsDeleteCmd.Flags().StringP("volser", "v", "", "A volume serial number.")

	// utilities: hrecall
	datasetsHrecallCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsHrecallCmd.MarkFlagRequired("ds-name")
	datasetsHrecallCmd.Flags().Bool("wait", false, "Wait for completion of the request.")
	datasetsHrecallCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")

	// utilities: hmigrate
	datasetsHmigrateCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsHmigrateCmd.MarkFlagRequired("ds-name")
	datasetsHmigrateCmd.Flags().Bool("wait", false, "Wait for completion of the request.")
	datasetsHmigrateCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")

	// utilities: hdelete
	datasetsHdeleteCmd.Flags().StringP("ds-name", "n", "", "The dataset name.")
	datasetsHdeleteCmd.MarkFlagRequired("ds-name")
	datasetsHdeleteCmd.Flags().Bool("wait", false, "Wait for completion of the request.")
	datasetsHdeleteCmd.Flags().Bool("purge", false, "Use PURGE=YES on ARCHDEL request.")
	datasetsHdeleteCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")

	// utilities: rename
	datasetsRenameCmd.Flags().StringP("ds-name", "n", "", "Target (new) dataset name.")
	datasetsRenameCmd.MarkFlagRequired("ds-name")
	datasetsRenameCmd.Flags().String("from-ds-name", "", "Source dataset name to rename from.")
	datasetsRenameCmd.MarkFlagRequired("from-ds-name")
	datasetsRenameCmd.Flags().String("member-name", "", "Target member name (for member rename).")
	datasetsRenameCmd.Flags().String("from-member", "", "Source member name to rename from.")
	datasetsRenameCmd.Flags().String("enq", "", "ENQ type for the target dataset: SHRW or EXCLU.")
	datasetsRenameCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")

	// utilities: copy
	datasetsCopyCmd.Flags().StringP("ds-name", "n", "", "Target dataset name.")
	datasetsCopyCmd.MarkFlagRequired("ds-name")
	datasetsCopyCmd.Flags().String("member-name", "", "Target member name.")
	datasetsCopyCmd.Flags().StringP("volser", "v", "", "Target volume serial (for uncataloged datasets).")
	datasetsCopyCmd.Flags().String("from-ds-name", "", "Source dataset name.")
	datasetsCopyCmd.Flags().String("from-member", "", "Source member name (* for all members).")
	datasetsCopyCmd.Flags().String("from-volser", "", "Source volume serial (for uncataloged datasets).")
	datasetsCopyCmd.Flags().Bool("from-alias", false, "Copy aliases along with main member.")
	datasetsCopyCmd.Flags().String("from-file", "", "Absolute z/OS UNIX source file path (alternative to --from-ds-name).")
	datasetsCopyCmd.Flags().String("from-file-type", "text", "Source file type: binary, executable, or text.")
	datasetsCopyCmd.Flags().String("enq", "", "ENQ type for the target dataset: SHR, SHRW, or EXCLU.")
	datasetsCopyCmd.Flags().Bool("replace", false, "Replace existing members in the target dataset.")
	datasetsCopyCmd.Flags().String("migrated-recall", "", "Migrated dataset handling: wait, nowait, or error.")
	datasetsCopyCmd.Flags().String("bpxk-autocvt", "", "Auto conversion for copy to POSIX files: on, all, or off.")

	// utilities: ams
	datasetsAmsCmd.Flags().StringArray("input", nil, "IDCAMS input statement (repeatable, each <= 255 chars).")
	datasetsAmsCmd.MarkFlagRequired("input")
	datasetsAmsCmd.Flags().String("target-system", "", "Target system nickname (X-IBM-Target-System).")
	datasetsAmsCmd.Flags().String("target-user", "", "User ID for the target system.")
	datasetsAmsCmd.Flags().String("target-password", "", "Password for the target system user.")

	// Wire up command tree
	datasetsUtilitiesCmd.AddCommand(datasetsHrecallCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsHmigrateCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsHdeleteCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsRenameCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsCopyCmd)
	datasetsUtilitiesCmd.AddCommand(datasetsAmsCmd)

	datasetsCmd.AddCommand(datasetsListCmd)
	datasetsCmd.AddCommand(datasetsMembersCmd)
	datasetsCmd.AddCommand(datasetsReadCmd)
	datasetsCmd.AddCommand(datasetsCreateCmd)
	datasetsCmd.AddCommand(datasetsDeleteCmd)
	datasetsCmd.AddCommand(datasetsUtilitiesCmd)

	rootCmd.AddCommand(datasetsCmd)
}
