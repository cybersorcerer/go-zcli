package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"zcli/internal/zosmf"

	"github.com/spf13/cobra"
)

var knownParmlibTypes = map[string]bool{
	"ALLOC": true, "AUTOR": true, "AXR": true, "BPXPRM": true,
	"CEAPRM": true, "CEEPARM": true, "CLOCK": true, "COMMND": true,
	"CONSOL": true, "COUPLE": true, "CUNUNI": true, "DEVSUP": true,
	"DIAG": true, "EXIT": true, "FXEPRM": true, "GRSCNF": true,
	"GRSRNL": true, "IEAFIX": true, "IEALPA": true, "IEAPAK": true,
	"IEASVC": true, "IEASYM": true, "IEASYS": true, "IEFOPZ": true,
	"IEFSSN": true, "IFAPRD": true, "IGDSMS": true, "IGGCAT": true,
	"IKJTSO": true, "IQPPRM": true, "IRRPRM": true, "IXGCNF": true,
	"IZUPRM": true, "LOAD": true, "LPALST": true, "PROG": true,
	"SCHED": true, "VATLST": true,
}

var parmlibCmd = &cobra.Command{
	Use:   "parmlib",
	Short: "Work with z/OS Parmlib management services",
	Long: `
DESCRIPTION
-----------
Work with z/OS Parmlib management services on a z/OS system.
Provides syntax validation for all 38 supported parmlib member types.`,
}

var parmlibValidateCmd = &cobra.Command{
	Use:   "validate <membertype>",
	Short: "Validate parmlib member syntax",
	Long: `
DESCRIPTION
-----------
Validate the syntax of z/OS parmlib members using the z/OSMF REST API.

MEMBER TYPES (specify without the xx suffix)
--------------------------------------------
ALLOC, AUTOR, AXR, BPXPRM, CEAPRM, CEEPARM, CLOCK, COMMND,
CONSOL, COUPLE, CUNUNI, DEVSUP, DIAG, EXIT, FXEPRM, GRSCNF,
GRSRNL, IEAFIX, IEALPA, IEAPAK, IEASVC, IEASYM, IEASYS,
IEFOPZ, IEFSSN, IFAPRD, IGDSMS, IGGCAT, IKJTSO, IQPPRM,
IRRPRM, IXGCNF, IZUPRM, LOAD, LPALST, PROG, SCHED, VATLST

SCENARIOS
---------
1. Validate all active members of a type:
   zcli parmlib validate BPXPRM

2. Validate all members of a type via specific LOADxx:
   zcli parmlib validate BPXPRM --load-member LOADAC --load-dataset SYS1.PARMLIB

3. Validate all parmlib types deeply (LOAD only):
   zcli parmlib validate LOAD --deep

4. Validate a specific member by name:
   zcli parmlib validate BPXPRM --member BPXPRM01 --dataset SYS1.PARMLIB

5. Validate parmlib content from stdin:
   cat BPXPRM01 | zcli parmlib validate BPXPRM`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		memberType := strings.ToUpper(args[0])
		member, _ := cmd.Flags().GetString("member")
		dataset, _ := cmd.Flags().GetString("dataset")
		volser, _ := cmd.Flags().GetString("volser")
		loadMember, _ := cmd.Flags().GetString("load-member")
		loadDataset, _ := cmd.Flags().GetString("load-dataset")
		loadVolser, _ := cmd.Flags().GetString("load-volser")
		deep, _ := cmd.Flags().GetBool("deep")
		system, _ := cmd.Flags().GetString("system")
		text, _ := cmd.Flags().GetBool("text")

		if !knownParmlibTypes[memberType] {
			fmt.Fprintf(os.Stderr, "Warning: %q is not a known parmlib member type.\n", memberType)
		}

		if deep && memberType != "LOAD" {
			return fmt.Errorf("--deep is only valid with member type LOAD")
		}

		// Read stdin if piped (Scenario 5 / 3-2)
		var stdinBody []byte
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			var err error
			stdinBody, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/parmlib/v1/%s/validate", memberType)

		var params []string
		if system != "" {
			params = append(params, "system="+system)
		}
		if deep {
			params = append(params, "deep=true")
		}
		if member != "" {
			params = append(params, "member="+member)
			if dataset != "" {
				params = append(params, "dataset="+dataset)
			}
			if volser != "" {
				params = append(params, "volser="+volser)
			}
		} else if len(stdinBody) == 0 {
			if loadMember != "" {
				params = append(params, "memLOAD="+loadMember)
				if loadDataset != "" {
					params = append(params, "dsnLOAD="+loadDataset)
				}
				if loadVolser != "" {
					params = append(params, "volsLOAD="+loadVolser)
				}
			}
		}

		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		var resp *zosmf.Response
		var err error
		if len(stdinBody) > 0 {
			resp, err = client.PutRaw(path, stdinBody, map[string]string{"Content-Type": "text/plain"})
		} else {
			resp, err = client.Put(path, nil, nil)
		}
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		if text {
			printParmlibValidateText(resp.Body)
			return nil
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

func printParmlibValidateText(body []byte) {
	var result struct {
		Result                string `json:"result"`
		NumberOfMembers       int    `json:"numberOfMembers"`
		NumberOfFailedMembers int    `json:"numberOfFailedMembers"`
		ParmlibDatasets       []struct {
			Volser  string `json:"volser"`
			Dataset string `json:"dataset"`
		} `json:"parmlibDatasets"`
		Details json.RawMessage `json:"details"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println(string(body))
		return
	}

	mark := "✓"
	if result.Result == "failed" {
		mark = "✗"
	}
	fmt.Printf("RESULT: %s %s  (%d validated, %d failed)\n",
		mark, strings.ToUpper(result.Result),
		result.NumberOfMembers, result.NumberOfFailedMembers)

	if len(result.ParmlibDatasets) > 0 {
		fmt.Println("\nParmlib Datasets:")
		for _, ds := range result.ParmlibDatasets {
			fmt.Printf("  %s", ds.Dataset)
			if ds.Volser != "" && ds.Volser != "null" {
				fmt.Printf("  (volser: %s)", ds.Volser)
			}
			fmt.Println()
		}
	}

	if result.Details == nil {
		return
	}

	// Scenario 3: details is a single member object
	var single struct {
		Member           string `json:"member"`
		Dataset          string `json:"dataset"`
		ValidationResult string `json:"validationResult"`
		Warning          string `json:"warning"`
		NumOfFailure     int    `json:"numOfFailure"`
		Failures         []parmlibFailure `json:"failures"`
	}
	if err := json.Unmarshal(result.Details, &single); err == nil && single.ValidationResult != "" {
		fmt.Println()
		printMemberResult("", single.Member, single.Dataset, single.ValidationResult, single.Warning, single.Failures)
		return
	}

	// Scenario 1/2: details is a map of membertype → object or array
	var multi map[string]json.RawMessage
	if err := json.Unmarshal(result.Details, &multi); err != nil {
		fmt.Println(string(result.Details))
		return
	}

	fmt.Println()
	for typeName, typeRaw := range multi {
		fmt.Printf("  %s\n", typeName)

		var members []json.RawMessage
		if err := json.Unmarshal(typeRaw, &members); err != nil {
			// Single warning object (no members of this type found)
			var warn struct {
				Warning string `json:"warning"`
			}
			if err2 := json.Unmarshal(typeRaw, &warn); err2 == nil && warn.Warning != "" {
				fmt.Printf("    ! %s\n", warn.Warning)
			}
			continue
		}

		for _, memberRaw := range members {
			var m struct {
				Member           string           `json:"member"`
				Dataset          string           `json:"dataset"`
				ValidationResult string           `json:"validationResult"`
				Warning          string           `json:"warning"`
				NumOfFailure     int              `json:"numOfFailure"`
				Failures         []parmlibFailure `json:"failures"`
			}
			if err := json.Unmarshal(memberRaw, &m); err != nil {
				continue
			}
			printMemberResult("  ", m.Member, m.Dataset, m.ValidationResult, m.Warning, m.Failures)
		}
	}
}

type parmlibFailure struct {
	Index    int `json:"index"`
	Location struct {
		Line     int `json:"line"`
		Position int `json:"position"`
	} `json:"location"`
	CurrentField  string   `json:"currentField"`
	ExpectedField []string `json:"expectedField"`
	AfterField    string   `json:"afterField"`
	Message       string   `json:"message"`
}

func printMemberResult(indent, member, dataset, validationResult, warning string, failures []parmlibFailure) {
	mark := "✓"
	if validationResult == "failed" {
		mark = "✗"
	} else if validationResult == "N/A" {
		mark = "-"
	}

	line := member
	if dataset != "" {
		line += "  " + dataset
	}
	if len(failures) > 0 {
		fmt.Printf("%s  %s %-30s (%d failure(s))\n", indent, mark, line, len(failures))
	} else {
		fmt.Printf("%s  %s %s\n", indent, mark, line)
	}

	if warning != "" {
		fmt.Printf("%s    ! %s\n", indent, warning)
	}

	for _, f := range failures {
		fmt.Printf("%s    [Line %d, Pos %d] %s after %s → expected: %s\n",
			indent, f.Location.Line, f.Location.Position,
			f.CurrentField, f.AfterField,
			strings.Join(f.ExpectedField, ", "))
	}
}

func init() {
	parmlibValidateCmd.Flags().StringP("member", "m", "", "Specific member name to validate (e.g. BPXPRM01).")
	parmlibValidateCmd.Flags().String("dataset", "", "Dataset containing the member (e.g. SYS1.PARMLIB).")
	parmlibValidateCmd.Flags().String("volser", "", "Volser for dataset if not cataloged.")
	parmlibValidateCmd.Flags().String("load-member", "", "LOADxx member name for type resolution (e.g. LOADAC).")
	parmlibValidateCmd.Flags().String("load-dataset", "", "Dataset containing the LOADxx member (e.g. SYS1.PARMLIB).")
	parmlibValidateCmd.Flags().String("load-volser", "", "Volser for load-dataset if not cataloged.")
	parmlibValidateCmd.Flags().Bool("deep", false, "Validate all types deeply (only valid with LOAD).")
	parmlibValidateCmd.Flags().String("system", "", "Remote system name in the sysplex.")
	parmlibValidateCmd.Flags().Bool("text", false, "Display the response as formatted text.")

	parmlibCmd.AddCommand(parmlibValidateCmd)
	rootCmd.AddCommand(parmlibCmd)
}
