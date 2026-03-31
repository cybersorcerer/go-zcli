package cmd

import (
	"fmt"
	"os"
	"strings"
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pathName, _ := cmd.Flags().GetString("path-name")
		maxItems, _ := cmd.Flags().GetString("max-items")
		lstat, _ := cmd.Flags().GetBool("lstat")
		// Filter parameters
		group, _ := cmd.Flags().GetString("group")
		mtime, _ := cmd.Flags().GetString("mtime")
		name, _ := cmd.Flags().GetString("name")
		size, _ := cmd.Flags().GetString("size")
		perm, _ := cmd.Flags().GetString("perm")
		fileType, _ := cmd.Flags().GetString("type")
		user, _ := cmd.Flags().GetString("user")
		// Tree traversal parameters
		depth, _ := cmd.Flags().GetString("depth")
		limit, _ := cmd.Flags().GetString("limit")
		filesys, _ := cmd.Flags().GetString("filesys")
		symlinks, _ := cmd.Flags().GetString("symlinks")

		if pathName == "" {
			pathName = fmt.Sprintf("/u/%s", Profile.User)
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs?path=%s", pathName)

		// Filter parameters
		if group != "" {
			path += "&group=" + group
		}
		if mtime != "" {
			path += "&mtime=" + mtime
		}
		if name != "" {
			path += "&name=" + name
		}
		if size != "" {
			path += "&size=" + size
		}
		if perm != "" {
			path += "&perm=" + perm
		}
		if fileType != "" {
			path += "&type=" + fileType
		}
		if user != "" {
			path += "&user=" + user
		}
		// Tree traversal parameters (only if filters are present)
		if depth != "" {
			path += "&depth=" + depth
		}
		if limit != "" {
			path += "&limit=" + limit
		}
		if filesys != "" {
			path += "&filesys=" + filesys
		}
		if symlinks != "" {
			path += "&symlinks=" + symlinks
		}

		headers := map[string]string{}
		if maxItems != "" {
			headers["X-IBM-Max-Items"] = maxItems
		}
		if lstat {
			headers["X-IBM-Lstat"] = "true"
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

var filesRetrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve a file from z/Unix",
	Long: `
DESCRIPTION
-----------
You can use this command to retrieve a file from z/Unix. To write the
retrieved data to the local file system also specify --local-file-name.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFileName, _ := cmd.Flags().GetString("zunix-file-name")
		localFileName, _ := cmd.Flags().GetString("local-file-name")
		fileType, _ := cmd.Flags().GetString("file-type")
		encoding, _ := cmd.Flags().GetString("encoding")
		charset, _ := cmd.Flags().GetString("charset")
		search, _ := cmd.Flags().GetString("search")
		research, _ := cmd.Flags().GetString("research")
		insensitive, _ := cmd.Flags().GetString("insensitive")
		maxReturnSize, _ := cmd.Flags().GetString("max-return-size")
		recordRange, _ := cmd.Flags().GetString("record-range")
		byteRange, _ := cmd.Flags().GetString("byte-range")
		ifNoneMatch, _ := cmd.Flags().GetString("if-none-match")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFileName)

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

		// Build X-IBM-Data-Type header
		var dataTypeHeader string
		if fileType == "text" {
			if encoding != "IBM-1047" {
				dataTypeHeader = fmt.Sprintf("text;fileEncoding=%s", encoding)
			} else {
				dataTypeHeader = "text"
			}
		} else {
			dataTypeHeader = fileType
		}

		headers := map[string]string{
			"X-IBM-Data-Type": dataTypeHeader,
		}
		if fileType == "text" {
			headers["Content-Type"] = fmt.Sprintf("text/plain;charset=%s", charset)
		}
		if recordRange != "" {
			headers["X-IBM-Record-Range"] = recordRange
		}
		if byteRange != "" {
			headers["Range"] = "bytes=" + byteRange
		}
		if ifNoneMatch != "" {
			headers["If-None-Match"] = ifNoneMatch
		}

		resp, err := client.Get(path, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 206, 304); apiErr != nil {
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFileName, _ := cmd.Flags().GetString("zunix-file-name")
		localFileName, _ := cmd.Flags().GetString("local-file-name")
		etag, _ := cmd.Flags().GetString("etag")
		fileType, _ := cmd.Flags().GetString("file-type")
		encoding, _ := cmd.Flags().GetString("encoding")
		charset, _ := cmd.Flags().GetString("charset")
		crlf, _ := cmd.Flags().GetBool("crlf")

		data, err := os.ReadFile(localFileName)
		if err != nil {
			return fmt.Errorf("error reading local file %s: %w", localFileName, err)
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFileName)

		// Build X-IBM-Data-Type header
		var dataTypeHeader string
		if fileType == "text" {
			var opts []string
			if encoding != "IBM-1047" {
				opts = append(opts, "fileEncoding="+encoding)
			}
			if crlf {
				opts = append(opts, "crlf=true")
			}
			if len(opts) > 0 {
				dataTypeHeader = "text;" + strings.Join(opts, ";")
			} else {
				dataTypeHeader = "text"
			}
		} else {
			dataTypeHeader = fileType
		}

		headers := map[string]string{
			"X-IBM-Data-Type": dataTypeHeader,
		}
		if fileType == "text" {
			headers["Content-Type"] = fmt.Sprintf("text/plain;charset=%s", charset)
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
	Args: cobra.NoArgs,
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		recursive, _ := cmd.Flags().GetBool("recursive")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		headers := make(map[string]string)
		if recursive {
			headers["X-IBM-Option"] = "recursive"
		}
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
			headers["X-IBM-Target-System-Password"] = targetPassword
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
file or directory. Operations include: chmod, chown, chtag, copy, extattr,
getfacl, move, setfacl, link, and unlink.`,
}

var filesChmodCmd = &cobra.Command{
	Use:   "chmod",
	Short: "Change mode (chmod) of z/UNIX file objects",
	Long: `
DESCRIPTION
-----------
chmod is the command used to change the access permissions and
the special mode flags of file system objects (files and directories).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		permissions, _ := cmd.Flags().GetString("permissions")
		followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")
		recursive, _ := cmd.Flags().GetBool("recursive")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

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

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesChownCmd = &cobra.Command{
	Use:   "chown",
	Short: "Change Owner (chown) of z/UNIX file objects",
	Long: `
DESCRIPTION
-----------
Used to change the owner of file system files and directories.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		owner, _ := cmd.Flags().GetString("owner")
		group, _ := cmd.Flags().GetString("group")
		followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")
		recursive, _ := cmd.Flags().GetBool("recursive")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		links := "follow"
		if !followSymlinks {
			links = "change"
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"request":   "chown",
			"owner":     owner,
			"links":     links,
			"recursive": recursive,
		}
		if group != "" {
			payload["group"] = group
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesCHtagCmd = &cobra.Command{
	Use:   "chtag",
	Short: "Change File Tagging (chtag) of z/UNIX file objects",
	Long: `
DESCRIPTION
-----------
chtag allows you to set, modify, remove, or display information in a file tag.
A file tag is composed of a text flag (txtflag) and a coded character set.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		action, _ := cmd.Flags().GetString("action")
		fileType, _ := cmd.Flags().GetString("file-type")
		codeset, _ := cmd.Flags().GetString("codeset")
		followSymlinks, _ := cmd.Flags().GetBool("follow-symlinks")
		recursive, _ := cmd.Flags().GetBool("recursive")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

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

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		action, _ := cmd.Flags().GetString("action")
		attributes, _ := cmd.Flags().GetString("attributes")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)
		payload := map[string]interface{}{
			"request": "extattr",
		}
		if action != "" && attributes != "" {
			payload[action] = attributes
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy a z/UNIX file, directory, or dataset to a UNIX file",
	Long: `
DESCRIPTION
-----------
Copy a UNIX file or directory, or a dataset/member, to a target UNIX file path.
Use --from for a UNIX source or --from-dataset for a dataset source.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		from, _ := cmd.Flags().GetString("from")
		fromDataset, _ := cmd.Flags().GetString("from-dataset")
		fromMember, _ := cmd.Flags().GetString("from-member")
		fromType, _ := cmd.Flags().GetString("from-type")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		recursive, _ := cmd.Flags().GetBool("recursive")
		links, _ := cmd.Flags().GetString("links")
		preserve, _ := cmd.Flags().GetString("preserve")
		bpxkAutocvt, _ := cmd.Flags().GetString("bpxk-autocvt")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		payload := map[string]interface{}{
			"request": "copy",
		}

		if from != "" {
			payload["from"] = from
			payload["overwrite"] = overwrite
			payload["recursive"] = recursive
			if links != "" {
				payload["links"] = links
			}
			if preserve != "" {
				payload["preserve"] = preserve
			}
		} else if fromDataset != "" {
			ds := map[string]interface{}{
				"dsn": fromDataset,
			}
			if fromMember != "" {
				ds["member"] = fromMember
			}
			if fromType != "" {
				ds["type"] = fromType
			}
			payload["from-dataset"] = ds
		}

		headers := make(map[string]string)
		if bpxkAutocvt != "" {
			headers["X-IBM-BPXK-AUTOCVT"] = bpxkAutocvt
		}
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesMoveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a z/UNIX file or directory",
	Long: `
DESCRIPTION
-----------
Move a UNIX file or directory to the target path.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		from, _ := cmd.Flags().GetString("from")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		payload := map[string]interface{}{
			"request":   "move",
			"from":      from,
			"overwrite": overwrite,
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesGetfaclCmd = &cobra.Command{
	Use:   "getfacl",
	Short: "Get file access control lists (ACLs)",
	Long: `
DESCRIPTION
-----------
Retrieve access control list (ACL) entries for a UNIX file or directory.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		aclType, _ := cmd.Flags().GetString("type")
		user, _ := cmd.Flags().GetString("user")
		useCommas, _ := cmd.Flags().GetBool("use-commas")
		suppressHeader, _ := cmd.Flags().GetBool("suppress-header")
		suppressBaseacl, _ := cmd.Flags().GetBool("suppress-baseacl")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		payload := map[string]interface{}{
			"request": "getfacl",
		}
		if aclType != "" {
			payload["type"] = aclType
		}
		if user != "" {
			payload["user"] = user
		}
		if useCommas {
			payload["use-commas"] = true
		}
		if suppressHeader {
			payload["suppress-header"] = true
		}
		if suppressBaseacl {
			payload["suppress-baseacl"] = true
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesSetfaclCmd = &cobra.Command{
	Use:   "setfacl",
	Short: "Set file access control lists (ACLs)",
	Long: `
DESCRIPTION
-----------
Set, modify, or delete access control list (ACL) entries for a UNIX file or directory.
Use one of: --set, --modify, --delete, or --delete-type.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		abort, _ := cmd.Flags().GetBool("abort")
		links, _ := cmd.Flags().GetString("links")
		deleteType, _ := cmd.Flags().GetString("delete-type")
		set, _ := cmd.Flags().GetString("set")
		modify, _ := cmd.Flags().GetString("modify")
		del, _ := cmd.Flags().GetString("delete")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		payload := map[string]interface{}{
			"request": "setfacl",
		}
		if abort {
			payload["abort"] = true
		}
		if links != "" {
			payload["links"] = links
		}
		if deleteType != "" {
			payload["delete-type"] = deleteType
		}
		if set != "" {
			payload["set"] = set
		}
		if modify != "" {
			payload["modify"] = modify
		}
		if del != "" {
			payload["delete"] = del
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Create a symbolic or external link",
	Long: `
DESCRIPTION
-----------
Create a symbolic link or external link to a UNIX file or directory.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		from, _ := cmd.Flags().GetString("from")
		linkType, _ := cmd.Flags().GetString("type")
		recursive, _ := cmd.Flags().GetBool("recursive")
		force, _ := cmd.Flags().GetBool("force")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		payload := map[string]interface{}{
			"request":   "link",
			"from":      from,
			"type":      linkType,
			"recursive": recursive,
			"force":     force,
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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

var filesUnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlink a file",
	Long: `
DESCRIPTION
-----------
Unlink (remove a link to) a UNIX file.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zunixFilePath, _ := cmd.Flags().GetString("zunix-file-path")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/fs%s", zunixFilePath)

		payload := map[string]interface{}{
			"request": "unlink",
		}

		headers := make(map[string]string)
		if targetSystem != "" {
			headers["X-IBM-Target-System"] = targetSystem
		}
		if targetUser != "" {
			headers["X-IBM-Target-System-User"] = targetUser
		}
		if targetPassword != "" {
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
	// files list
	filesListCmd.Flags().StringP("path-name", "p", "", "UNIX directory or file path. DEFAULT /u/<USERID>")
	// Custom headers
	filesListCmd.Flags().String("max-items", "", "Maximum number of items to return (0 = all, default 1000).")
	filesListCmd.Flags().Bool("lstat", false, "Perform lstat() instead of stat() on the path.")
	// Filter parameters
	filesListCmd.Flags().String("group", "", "Filter by group owner name or GID.")
	filesListCmd.Flags().String("mtime", "", "Filter by modification time in days (e.g. -1, +7, 3).")
	filesListCmd.Flags().String("name", "", "Filter by filename pattern (fnmatch rules).")
	filesListCmd.Flags().String("size", "", "Filter by file size (e.g. +1M, -500K, 1024). Suffixes: K, M, G.")
	filesListCmd.Flags().String("perm", "", "Filter by permissions (octal, prefix - for all-bits-set match).")
	filesListCmd.Flags().String("type", "", "Filter by type: c (char), d (dir), f (file), l (symlink), p (pipe), s (socket).")
	filesListCmd.Flags().String("user", "", "Filter by user owner name or UID.")
	// Tree traversal parameters (used with filters)
	filesListCmd.Flags().String("depth", "", "Max directory depth (0 = unlimited, 1 = path only).")
	filesListCmd.Flags().String("limit", "", "Max items to return (overrides --max-items if both given).")
	filesListCmd.Flags().String("filesys", "", "Filesystem traversal: same (default) or all.")
	filesListCmd.Flags().String("symlinks", "", "Symlink handling: follow (default) or report.")

	// files retrieve
	filesRetrieveCmd.Flags().StringP("zunix-file-name", "z", "", "Full path to file name to retrieve.")
	filesRetrieveCmd.MarkFlagRequired("zunix-file-name")
	filesRetrieveCmd.Flags().StringP("local-file-name", "l", "", "Full path to file on local file system to write to.")
	filesRetrieveCmd.Flags().String("file-type", "text", "Data type: text or binary.")
	filesRetrieveCmd.Flags().StringP("encoding", "e", "IBM-1047", "EBCDIC code page on z/Unix (text mode).")
	filesRetrieveCmd.Flags().StringP("charset", "c", "ISO8859-1", "Local charset for text conversion.")
	// Query parameters
	filesRetrieveCmd.Flags().String("search", "", "Search for the first line containing this string.")
	filesRetrieveCmd.Flags().String("research", "", "Search using an extended regular expression.")
	filesRetrieveCmd.Flags().String("insensitive", "", "Case insensitive search: true (default) or false.")
	filesRetrieveCmd.Flags().String("max-return-size", "", "Max lines to return for search/research (default 100).")
	// Headers
	filesRetrieveCmd.Flags().String("record-range", "", "Record/line range: SSS-EEE or SSS,NNN (0-based).")
	filesRetrieveCmd.Flags().String("byte-range", "", "Byte range for binary mode: e.g. 0-499, 500-, -500.")
	filesRetrieveCmd.Flags().String("if-none-match", "", "ETag token for conditional retrieval (HTTP 304 if unchanged).")

	// files write
	filesWriteCmd.Flags().StringP("zunix-file-name", "z", "", "Full path to file name of z/Unix file to write to.")
	filesWriteCmd.MarkFlagRequired("zunix-file-name")
	filesWriteCmd.Flags().StringP("local-file-name", "l", "", "Full path to local file name.")
	filesWriteCmd.MarkFlagRequired("local-file-name")
	filesWriteCmd.Flags().String("etag", "", "ETag for conditional write (If-Match header).")
	filesWriteCmd.Flags().String("file-type", "text", "Data type for conversion: text or binary.")
	filesWriteCmd.Flags().StringP("encoding", "e", "IBM-1047", "z/OS UNIX file encoding (X-IBM-Data-Type fileEncoding).")
	filesWriteCmd.Flags().StringP("charset", "c", "ISO8859-1", "Local file charset for Content-Type header.")
	filesWriteCmd.Flags().Bool("crlf", false, "Use CRLF line terminators (X-IBM-Data-Type crlf=true).")

	// files create
	filesCreateCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file to create.")
	filesCreateCmd.MarkFlagRequired("zunix-file-path")
	filesCreateCmd.Flags().String("file-type", "", "File type to create (file/dir).")
	filesCreateCmd.MarkFlagRequired("file-type")
	filesCreateCmd.Flags().StringP("mode", "m", "rw-rw-rw-", "The file mode to use.")

	// files delete
	filesDeleteCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory to delete.")
	filesDeleteCmd.MarkFlagRequired("zunix-file-path")
	filesDeleteCmd.Flags().BoolP("recursive", "r", false, "Delete non-empty directory and all contents.")
	filesDeleteCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesDeleteCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesDeleteCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util chmod
	filesChmodCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesChmodCmd.MarkFlagRequired("zunix-file-path")
	filesChmodCmd.Flags().StringP("permissions", "p", "644", "Mode value (POSIX symbolic form or octal).")
	filesChmodCmd.Flags().Bool("follow-symlinks", true, "Follow symlinks (false = suppress).")
	filesChmodCmd.Flags().BoolP("recursive", "r", false, "Apply change recursively (chmod -R).")
	filesChmodCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesChmodCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesChmodCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util chown
	filesChownCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesChownCmd.MarkFlagRequired("zunix-file-path")
	filesChownCmd.Flags().StringP("owner", "o", "", "User ID or UID.")
	filesChownCmd.MarkFlagRequired("owner")
	filesChownCmd.Flags().StringP("group", "g", "", "Group ID or GID.")
	filesChownCmd.Flags().Bool("follow-symlinks", true, "Follow symlinks (false = change link itself, chown -h).")
	filesChownCmd.Flags().BoolP("recursive", "r", false, "Apply change recursively (chown -R).")
	filesChownCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesChownCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesChownCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util chtag
	filesCHtagCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesCHtagCmd.MarkFlagRequired("zunix-file-path")
	filesCHtagCmd.Flags().StringP("action", "a", "", "Tag action: set, remove, or list.")
	filesCHtagCmd.MarkFlagRequired("action")
	filesCHtagCmd.Flags().String("file-type", "mixed", "File type: binary, mixed, or text (set action only).")
	filesCHtagCmd.Flags().StringP("codeset", "c", "IBM-1047", "Coded character set (set action only).")
	filesCHtagCmd.Flags().Bool("follow-symlinks", true, "Follow symlinks (false = suppress).")
	filesCHtagCmd.Flags().BoolP("recursive", "r", false, "Apply change recursively (chtag -R).")
	filesCHtagCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesCHtagCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesCHtagCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util extattr
	filesExtAttrCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesExtAttrCmd.MarkFlagRequired("zunix-file-path")
	filesExtAttrCmd.Flags().StringP("action", "a", "", "Action: set or reset (omit to display).")
	filesExtAttrCmd.Flags().String("attributes", "", "Extended attributes: a (APF), l (shared lib), p (program-ctrl), s (shared addr).")
	filesExtAttrCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesExtAttrCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesExtAttrCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util copy
	filesCopyCmd.Flags().StringP("zunix-file-path", "z", "", "Target UNIX file or directory path.")
	filesCopyCmd.MarkFlagRequired("zunix-file-path")
	filesCopyCmd.Flags().String("from", "", "Source UNIX file or directory to copy.")
	filesCopyCmd.Flags().String("from-dataset", "", "Source dataset name (fully qualified).")
	filesCopyCmd.Flags().String("from-member", "", "Source dataset member name.")
	filesCopyCmd.Flags().String("from-type", "", "Dataset type: binary, executable, or text.")
	filesCopyCmd.Flags().Bool("overwrite", true, "Overwrite target if it exists (UNIX source only).")
	filesCopyCmd.Flags().BoolP("recursive", "r", false, "Copy recursively (cp -R, UNIX source only).")
	filesCopyCmd.Flags().String("links", "", "Symlink handling: none, src (cp -H), or all (cp -L).")
	filesCopyCmd.Flags().String("preserve", "", "Preserve attributes: none, modtime (cp -m), or all (cp -p).")
	filesCopyCmd.Flags().String("bpxk-autocvt", "", "Auto conversion: on/all or off (X-IBM-BPXK-AUTOCVT).")
	filesCopyCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesCopyCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesCopyCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util move
	filesMoveCmd.Flags().StringP("zunix-file-path", "z", "", "Target UNIX file or directory path.")
	filesMoveCmd.MarkFlagRequired("zunix-file-path")
	filesMoveCmd.Flags().String("from", "", "Source UNIX file or directory to move.")
	filesMoveCmd.MarkFlagRequired("from")
	filesMoveCmd.Flags().Bool("overwrite", true, "Overwrite target if it exists.")
	filesMoveCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesMoveCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesMoveCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util getfacl
	filesGetfaclCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesGetfaclCmd.MarkFlagRequired("zunix-file-path")
	filesGetfaclCmd.Flags().String("type", "", "ACL type: access (default), dir, or file.")
	filesGetfaclCmd.Flags().String("user", "", "Filter ACL entries for this user ID or UID.")
	filesGetfaclCmd.Flags().Bool("use-commas", false, "Separate ACL entries with commas instead of newlines.")
	filesGetfaclCmd.Flags().Bool("suppress-header", false, "Do not display the comment header (getfacl -m).")
	filesGetfaclCmd.Flags().Bool("suppress-baseacl", false, "Display only extended ACL entries (getfacl -o).")
	filesGetfaclCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesGetfaclCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesGetfaclCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util setfacl
	filesSetfaclCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file or directory.")
	filesSetfaclCmd.MarkFlagRequired("zunix-file-path")
	filesSetfaclCmd.Flags().Bool("abort", false, "Abort on error or warning (setfacl -a).")
	filesSetfaclCmd.Flags().String("links", "", "Symlink handling: follow (default) or suppress.")
	filesSetfaclCmd.Flags().String("delete-type", "", "Delete all extended ACLs by type: access, dir, file, or every.")
	filesSetfaclCmd.Flags().String("set", "", "Replace all ACLs with these entries (setfacl -s).")
	filesSetfaclCmd.Flags().String("modify", "", "Modify ACL entries.")
	filesSetfaclCmd.Flags().String("delete", "", "Delete specified ACL entries.")
	filesSetfaclCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesSetfaclCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesSetfaclCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util link
	filesLinkCmd.Flags().StringP("zunix-file-path", "z", "", "Target link path.")
	filesLinkCmd.MarkFlagRequired("zunix-file-path")
	filesLinkCmd.Flags().String("from", "", "Source file or directory to link.")
	filesLinkCmd.MarkFlagRequired("from")
	filesLinkCmd.Flags().String("type", "", "Link type: symbol or external.")
	filesLinkCmd.MarkFlagRequired("type")
	filesLinkCmd.Flags().BoolP("recursive", "r", false, "Link files recursively (ln -R).")
	filesLinkCmd.Flags().BoolP("force", "f", false, "Force link, delete conflicting paths (ln -f).")
	filesLinkCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesLinkCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesLinkCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// util unlink
	filesUnlinkCmd.Flags().StringP("zunix-file-path", "z", "", "Path to the file to unlink.")
	filesUnlinkCmd.MarkFlagRequired("zunix-file-path")
	filesUnlinkCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesUnlinkCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesUnlinkCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// Wire up command tree
	filesUtilCmd.AddCommand(filesChmodCmd)
	filesUtilCmd.AddCommand(filesChownCmd)
	filesUtilCmd.AddCommand(filesCHtagCmd)
	filesUtilCmd.AddCommand(filesExtAttrCmd)
	filesUtilCmd.AddCommand(filesCopyCmd)
	filesUtilCmd.AddCommand(filesMoveCmd)
	filesUtilCmd.AddCommand(filesGetfaclCmd)
	filesUtilCmd.AddCommand(filesSetfaclCmd)
	filesUtilCmd.AddCommand(filesLinkCmd)
	filesUtilCmd.AddCommand(filesUnlinkCmd)

	filesCmd.AddCommand(filesListCmd)
	filesCmd.AddCommand(filesRetrieveCmd)
	filesCmd.AddCommand(filesWriteCmd)
	filesCmd.AddCommand(filesCreateCmd)
	filesCmd.AddCommand(filesDeleteCmd)
	filesCmd.AddCommand(filesUtilCmd)

	rootCmd.AddCommand(filesCmd)
}
