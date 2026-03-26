package cmd

import (
	"fmt"
	"os"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Interact with z/OS z/Unix files",
	Long: `
DESCRIPTION
-----------
Interact with z/OS z/Unix files.`,
}

var filesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List z/Unix files and directories",
	Long: `
DESCRIPTION
-----------
You can use this command to obtain a list
of z/Unix files and directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pathName, _ := cmd.Flags().GetString("path-name")

		if pathName == "" {
			pathName = fmt.Sprintf("/u/%s", Profile.User)
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs?path=%s", pathName)

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

var filesRetrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve a file from z/Unix",
	Long: `
DESCRIPTION
-----------
You can use this command to retrieve a file from z/Unix. To write the
retrieved data to the local file system also specify --local-file-name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFileName, _ := cmd.Flags().GetString("zunix-file-name")
		localFileName, _ := cmd.Flags().GetString("local-file-name")
		fileType, _ := cmd.Flags().GetString("file-type")
		encoding, _ := cmd.Flags().GetString("encoding")
		charset, _ := cmd.Flags().GetString("charset")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFileName)

		headers := map[string]string{
			"X-IBM-Data-Type": fmt.Sprintf("%s;fileEncoding=%s", fileType, encoding),
		}
		if fileType == "text" {
			headers["Content-Type"] = fmt.Sprintf("text/plain;charset=%s", charset)
		} else {
			headers["Content-Type"] = "text/plain"
		}

		resp, err := client.Get(path, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		if localFileName == "" {
			fmt.Println(resp.BodyString())
		} else {
			if err := os.WriteFile(localFileName, resp.Body, 0644); err != nil {
				return fmt.Errorf("error writing local file %s: %w", localFileName, err)
			}
			fmt.Printf("File written to %s\n", localFileName)
		}
		if etag := resp.Headers.Get("ETag"); etag != "" {
			fmt.Printf("ETag: %s\n", etag)
		}
		return nil
	},
}

var filesWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Write to a file in z/Unix",
	Long: `
DESCRIPTION
-----------
You can use this command to write to a file in z/Unix. The command will read
data from --local-file-name and write it to --zunix-file-name. You can specify
an etag (--etag) formerly returned by the retrieve command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFileName, _ := cmd.Flags().GetString("zunix-file-name")
		localFileName, _ := cmd.Flags().GetString("local-file-name")
		etag, _ := cmd.Flags().GetString("etag")
		fileType, _ := cmd.Flags().GetString("file-type")
		encoding, _ := cmd.Flags().GetString("encoding")
		charset, _ := cmd.Flags().GetString("charset")

		data, err := os.ReadFile(localFileName)
		if err != nil {
			return fmt.Errorf("error reading local file %s: %w", localFileName, err)
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFileName)

		headers := map[string]string{
			"X-IBM-Data-Type": fmt.Sprintf("%s;fileEncoding=%s", fileType, encoding),
		}
		if fileType == "text" {
			headers["Content-Type"] = fmt.Sprintf("text/plain;charset=%s", charset)
		} else {
			headers["Content-Type"] = "text/plain"
		}
		if etag != "" {
			headers["If-Match"] = etag
		}

		resp, err := client.PutRaw(path, data, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 201, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

var filesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a z/UNIX file or directory",
	Long: `
DESCRIPTION
-----------
The command will create a file or directory at the specified path.
You can specify the file type using --file-type (file or dir).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		fileType, _ := cmd.Flags().GetString("file-type")
		mode, _ := cmd.Flags().GetString("mode")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"type": fileType,
			"mode": mode,
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

var filesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a z/UNIX file or directory",
	Long: `
DESCRIPTION
-----------
The command will delete a file or directory at the specified path.
You decide if a non empty directory should be deleted by specifying --recursive.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		recursive, _ := cmd.Flags().GetBool("recursive")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		var headers map[string]string
		if recursive {
			headers = map[string]string{"X-IBM-Option": "recursive"}
		}

		resp, err := client.Delete(path, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

// util subgroup
var filesUtilCmd = &cobra.Command{
	Use:   "util",
	Short: "z/Unix file utilities",
	Long: `
DESCRIPTION
-----------
You can use the z/OS UNIX file utilities to operate on a UNIX System Services
file or directory. Operations include: chmod, chown, chtag, and extattr.`,
}

var filesChmodCmd = &cobra.Command{
	Use:   "chmod",
	Short: "Change mode (chmod) of z/UNIX file objects",
	Long: `
DESCRIPTION
-----------
chmod is the command used to change the access permissions and
the special mode flags of file system objects (files and directories).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		permissions, _ := cmd.Flags().GetString("permissions")
		followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")
		recursive, _ := cmd.Flags().GetBool("recursive")

		links := "follow"
		if !followSymlinks {
			links = "suppress"
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"request":   "chmod",
			"mode":      permissions,
			"links":     links,
			"recursive": recursive,
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

var filesChownCmd = &cobra.Command{
	Use:   "chown",
	Short: "Change Owner (chown) of z/UNIX file objects",
	Long: `
DESCRIPTION
-----------
Used to change the owner of file system files and directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		owner, _ := cmd.Flags().GetString("owner")
		group, _ := cmd.Flags().GetString("group")
		followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")
		recursive, _ := cmd.Flags().GetBool("recursive")

		links := "follow"
		if !followSymlinks {
			links = "suppress"
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"request":   "chown",
			"owner":     owner,
			"group":     group,
			"links":     links,
			"recursive": recursive,
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

var filesCHtagCmd = &cobra.Command{
	Use:   "chtag",
	Short: "Change File Tagging (chtag) of z/UNIX file objects",
	Long: `
DESCRIPTION
-----------
chtag allows you to set, modify, remove, or display information in a file tag.
A file tag is composed of a text flag (txtflag) and a coded character set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		action, _ := cmd.Flags().GetString("action")
		fileType, _ := cmd.Flags().GetString("file-type")
		codeset, _ := cmd.Flags().GetString("codeset")
		followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")
		recursive, _ := cmd.Flags().GetBool("recursive")

		links := "change"
		if !followSymlinks {
			links = "suppress"
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"request":   "chtag",
			"action":    action,
			"type":      fileType,
			"links":     links,
			"recursive": recursive,
		}
		if action == "set" {
			payload["codeset"] = codeset
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

var filesExtAttrCmd = &cobra.Command{
	Use:   "extattr",
	Short: "Set, reset, and display extended attributes for files",
	Long: `
DESCRIPTION
-----------
The following attributes are supported:
- a: APF-authorized
- l: shared library region
- p: program-controlled
- s: shared address space`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		action, _ := cmd.Flags().GetString("action")
		attributes, _ := cmd.Flags().GetString("attributes")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"request": "extattr",
		}
		if action != "" && attributes != "" {
			payload[action] = attributes
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
	// files list
	filesListCmd.Flags().StringP("path-name", "p", "", "A path or file name.")

	// files retrieve
	filesRetrieveCmd.Flags().StringP("zunix-file-name", "z", "", "Full path to file name to retrieve.")
	filesRetrieveCmd.MarkFlagRequired("zunix-file-name")
	filesRetrieveCmd.Flags().StringP("local-file-name", "l", "", "Full path to file on local file system to write to.")
	filesRetrieveCmd.Flags().String("file-type", "text", "File type, used for conversion (text/binary).")
	filesRetrieveCmd.Flags().StringP("encoding", "e", "IBM-1047", "Codepage on z/Unix, used for conversion.")
	filesRetrieveCmd.Flags().StringP("charset", "c", "ISO8859-1", "Codepage of the local file used for conversion.")

	// files write
	filesWriteCmd.Flags().StringP("zunix-file-name", "z", "", "Full path to file name of z/Unix file to write to.")
	filesWriteCmd.MarkFlagRequired("zunix-file-name")
	filesWriteCmd.Flags().StringP("local-file-name", "l", "", "Full path to local file name.")
	filesWriteCmd.MarkFlagRequired("local-file-name")
	filesWriteCmd.Flags().String("etag", "", "The etag returned from retrieve or empty string.")
	filesWriteCmd.Flags().String("file-type", "text", "File type, used for conversion (text/binary).")
	filesWriteCmd.Flags().StringP("encoding", "e", "IBM-1047", "Codepage on z/Unix, used for conversion.")
	filesWriteCmd.Flags().StringP("charset", "c", "ISO8859-1", "Codepage of the local file used for conversion.")

	// files create
	filesCreateCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file to create.")
	filesCreateCmd.MarkFlagRequired("zunix-file-path")
	filesCreateCmd.Flags().String("file-type", "", "File type to create (file/dir).")
	filesCreateCmd.MarkFlagRequired("file-type")
	filesCreateCmd.Flags().StringP("mode", "m", "rw-rw-rw-", "The file mode to use.")

	// files delete
	filesDeleteCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory to delete.")
	filesDeleteCmd.MarkFlagRequired("zunix-file-path")
	filesDeleteCmd.Flags().BoolP("recursive", "r", false, "If true a non empty directory will be deleted.")

	// util chmod
	filesChmodCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesChmodCmd.MarkFlagRequired("zunix-file-path")
	filesChmodCmd.Flags().StringP("permissions", "p", "644", "The mode value (POSIX symbolic form or octal).")
	filesChmodCmd.Flags().Bool("follow-symlinks", true, "If true apply mode change to file/directory pointed to by links.")
	filesChmodCmd.Flags().BoolP("recursive", "r", false, "If true apply the change recursively.")

	// util chown
	filesChownCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesChownCmd.MarkFlagRequired("zunix-file-path")
	filesChownCmd.Flags().StringP("owner", "o", "", "The file or directory owner.")
	filesChownCmd.MarkFlagRequired("owner")
	filesChownCmd.Flags().StringP("group", "g", "", "The file or directory group owner.")
	filesChownCmd.MarkFlagRequired("group")
	filesChownCmd.Flags().Bool("follow-symlinks", true, "If true apply owner change to file/directory pointed to by links.")
	filesChownCmd.Flags().BoolP("recursive", "r", false, "If true apply the change recursively.")

	// util chtag
	filesCHtagCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesCHtagCmd.MarkFlagRequired("zunix-file-path")
	filesCHtagCmd.Flags().StringP("action", "a", "", "The tag action to perform (set/remove/list).")
	filesCHtagCmd.MarkFlagRequired("action")
	filesCHtagCmd.Flags().String("file-type", "mixed", "The file type (binary/mixed/text).")
	filesCHtagCmd.Flags().StringP("codeset", "c", "IBM-1047", "The Codeset to use.")
	filesCHtagCmd.Flags().Bool("follow-symlinks", true, "If true apply tag action to file/directory pointed to by links.")
	filesCHtagCmd.Flags().BoolP("recursive", "r", false, "If true apply the change recursively.")

	// util extattr
	filesExtAttrCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesExtAttrCmd.MarkFlagRequired("zunix-file-path")
	filesExtAttrCmd.Flags().StringP("action", "a", "", "The action to perform (set/reset).")
	filesExtAttrCmd.Flags().String("attributes", "", "The extended attributes to set (a,l,p,s).")

	// Wire up command tree
	filesUtilCmd.AddCommand(filesChmodCmd)
	filesUtilCmd.AddCommand(filesChownCmd)
	filesUtilCmd.AddCommand(filesCHtagCmd)
	filesUtilCmd.AddCommand(filesExtAttrCmd)

	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesRetrieveCmd)
	filesCmd.AddCommand(filesWriteCmd)
	filesCmd.AddCommand(filesCreateCmd)
	filesCmd.AddCommand(filesDeleteCmd)
	filesCmd.AddCommand(filesUtilCmd)

	rootCmd.AddCommand(filesCmd)
}
