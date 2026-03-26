package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"zcli/internal/tui"
	"zcli/internal/zosmf"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/spf13/cobra"
)

// ---------- Bubbletea messages ----------

type jobListMsg []table.Row
type jobFilesMsg []table.Row
type spoolContentMsg []table.Row
type jobActionMsg struct {
	jobname string
	jobid   string
	ok      bool
	message string
}

// ---------- Bubbletea model ----------

type jobModel struct {
	jobTable     table.Model
	fileTable    table.Model
	spoolBrowser table.Model

	statusMsg string
	view      string // "jobs", "files", "spool"

	owner  string
	prefix string
	status string
	jtype  string

	jobname string
	jobid   string
	ddname  string
	url     string
	action  string
}

func newJobModel(owner, prefix, status, jtype string) jobModel {
	m := jobModel{
		view:   "jobs",
		owner:  owner,
		prefix: prefix,
		status: status,
		jtype:  jtype,
	}
	m.jobTable = buildJobTable(nil)
	m.fileTable = buildFileTable(nil)
	m.spoolBrowser = buildSpoolBrowser("", nil)
	return m
}

func (m jobModel) Init() tea.Cmd {
	return fetchJobList(m.owner, m.prefix, m.status, m.jtype)
}

func (m jobModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.jobTable = m.jobTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)
		m.fileTable = m.fileTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)
		m.spoolBrowser = m.spoolBrowser.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)

	case jobListMsg:
		m.jobTable = m.jobTable.WithRows([]table.Row(msg))
		m.statusMsg = fmt.Sprintf("  %d jobs loaded", len(msg))

	case jobFilesMsg:
		m.fileTable = m.fileTable.WithRows([]table.Row(msg))

	case spoolContentMsg:
		m.spoolBrowser = m.spoolBrowser.WithRows([]table.Row(msg))

	case jobActionMsg:
		if msg.ok {
			m.statusMsg = fmt.Sprintf("  %s %s(%s) successful", m.action, msg.jobname, msg.jobid)
		} else {
			m.statusMsg = fmt.Sprintf("  %s %s(%s) failed: %s", m.action, msg.jobname, msg.jobid, msg.message)
		}
		m.action = ""
		cmds = append(cmds, fetchJobList(m.owner, m.prefix, m.status, m.jtype))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+r":
			if m.view == "jobs" {
				cmds = append(cmds, fetchJobList(m.owner, m.prefix, m.status, m.jtype))
			}

		case "ctrl+s", "enter":
			if m.view == "jobs" {
				row := m.jobTable.HighlightedRow()
				m.jobname = str(row.Data[tui.JobColumnJobname])
				m.jobid = str(row.Data[tui.JobColumnJobID])
				m.view = "files"
				m.jobTable = m.jobTable.Focused(false)
				m.fileTable = m.fileTable.Focused(true)
				cmds = append(cmds, fetchJobFiles(m.jobname, m.jobid))
			} else if m.view == "files" {
				row := m.fileTable.HighlightedRow()
				m.ddname = str(row.Data[tui.FileColumnDDName])
				m.url = str(row.Data[tui.FileLink])
				m.view = "spool"
				m.fileTable = m.fileTable.Focused(false)
				m.spoolBrowser = m.spoolBrowser.Focused(true)
				cmds = append(cmds, fetchSpoolContent(m.url))
			}

		case "ctrl+p", "esc":
			if m.view == "files" {
				m.view = "jobs"
				m.fileTable = m.fileTable.Focused(false)
				m.jobTable = m.jobTable.Focused(true)
			} else if m.view == "spool" {
				m.view = "files"
				m.spoolBrowser = m.spoolBrowser.Focused(false)
				m.fileTable = m.fileTable.Focused(true)
			}

		case "ctrl+l":
			if m.view == "jobs" {
				row := m.jobTable.HighlightedRow()
				m.jobname = str(row.Data[tui.JobColumnJobname])
				m.jobid = str(row.Data[tui.JobColumnJobID])
				m.action = "cancel"
				cmds = append(cmds, doJobAction(m.jobname, m.jobid, "cancel"))
			}

		case "ctrl+h":
			if m.view == "jobs" {
				row := m.jobTable.HighlightedRow()
				m.jobname = str(row.Data[tui.JobColumnJobname])
				m.jobid = str(row.Data[tui.JobColumnJobID])
				m.action = "hold"
				cmds = append(cmds, doJobAction(m.jobname, m.jobid, "hold"))
			}

		case "ctrl+e":
			if m.view == "jobs" {
				row := m.jobTable.HighlightedRow()
				m.jobname = str(row.Data[tui.JobColumnJobname])
				m.jobid = str(row.Data[tui.JobColumnJobID])
				m.action = "release"
				cmds = append(cmds, doJobAction(m.jobname, m.jobid, "release"))
			}
		}
	}

	// Update the active table
	switch m.view {
	case "jobs":
		m.updateJobFooter()
		m.jobTable, cmd = m.jobTable.Update(msg)
	case "files":
		m.updateFileFooter()
		m.fileTable, cmd = m.fileTable.Update(msg)
	case "spool":
		m.spoolBrowser, cmd = m.spoolBrowser.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m jobModel) View() string {
	var b strings.Builder
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg + "\n")
	}
	pad := lipgloss.NewStyle().Padding(0)
	switch m.view {
	case "jobs":
		b.WriteString(pad.Render(m.jobTable.View()))
	case "files":
		b.WriteString(pad.Render(m.fileTable.View()))
	case "spool":
		b.WriteString(pad.Render(m.spoolBrowser.View()))
	}
	return b.String()
}

func (m *jobModel) updateJobFooter() {
	m.jobTable = m.jobTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d | ctrl+c quit | ctrl+r refresh | ctrl+s select | ctrl+l cancel | ctrl+h hold | ctrl+e release | ↑/↓ move | f7/f8 page",
		m.jobTable.CurrentPage(), m.jobTable.MaxPages(),
	))
}

func (m *jobModel) updateFileFooter() {
	m.fileTable = m.fileTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d [%s/%s] | ctrl+s select | esc back | ctrl+c quit",
		m.fileTable.CurrentPage(), m.fileTable.MaxPages(), m.jobname, m.jobid,
	))
}

// ---------- Bubbletea commands ----------

func fetchJobList(owner, prefix, status, jtype string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		query := fmt.Sprintf("?owner=%s&prefix=%s&status=%s&exec-data=y", owner, prefix, status)
		resp, err := client.Get("/restjobs/jobs"+query, nil)
		if err != nil {
			return jobListMsg(nil)
		}
		if zosmf.CheckResponse(resp, 200) != nil {
			return jobListMsg(nil)
		}

		var jobs []struct {
			Jobname    string `json:"jobname"`
			Jobid      string `json:"jobid"`
			Owner      string `json:"owner"`
			Class      string `json:"class"`
			Status     string `json:"status"`
			Retcode    string `json:"retcode"`
			Type       string `json:"type"`
			ExecStart  string `json:"exec-started"`
			ExecEnd    string `json:"exec-ended"`
			ExecMember string `json:"exec-member"`
			ExecSystem string `json:"exec-system"`
		}
		if err := resp.JSON(&jobs); err != nil {
			return jobListMsg(nil)
		}

		var rows []table.Row
		i := 0
		for _, j := range jobs {
			if jtype != "*" && j.Type != jtype {
				continue
			}
			i++
			rc := j.Retcode
			if rc == "" {
				rc = "n/a"
			}
			started := strings.ReplaceAll(strings.ReplaceAll(j.ExecStart, "T", " "), "Z", "")
			ended := strings.ReplaceAll(strings.ReplaceAll(j.ExecEnd, "T", " "), "Z", "")

			rd := table.RowData{
				tui.ColumnID:         fmt.Sprintf("%3d", i),
				tui.JobColumnJobname: j.Jobname,
				tui.JobColumnJobID:   j.Jobid,
				tui.JobColumnOwner:   j.Owner,
				tui.JobColumnClass:   j.Class,
				tui.ColumnStatus:     j.Status,
				tui.JobColumnRetcode: rc,
				tui.ColumnType:       j.Type,
				tui.JobColumnStarted: started,
				tui.JobColumnEnded:   ended,
				tui.JobColumnMember:  j.ExecMember,
				tui.JobColumnSystem:  j.ExecSystem,
			}
			row := table.NewRow(rd)
			switch {
			case strings.Contains(rc, "ABEND"):
				row = row.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C")))
			case strings.Contains(rc, "JCL"):
				row = row.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF066")))
			case strings.Contains(j.Status, "ACTIVE"):
				row = row.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#00a65a")))
			}
			rows = append(rows, row)
		}
		return jobListMsg(rows)
	}
}

func fetchJobFiles(jobname, jobid string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restjobs/jobs/%s/%s/files", jobname, jobid)
		resp, err := client.Get(path, nil)
		if err != nil || zosmf.CheckResponse(resp, 200) != nil {
			return jobFilesMsg(nil)
		}

		var files []struct {
			Ddname      string `json:"ddname"`
			Stepname    string `json:"stepname"`
			Procstep    string `json:"procstep"`
			ID          int    `json:"id"`
			Jobid       string `json:"jobid"`
			Class       string `json:"class"`
			RecordCount int    `json:"record-count"`
			ByteCount   int    `json:"byte-count"`
			Lrecl       int    `json:"lrecl"`
			RecordsURL  string `json:"records-url"`
		}
		if err := resp.JSON(&files); err != nil {
			return jobFilesMsg(nil)
		}

		var rows []table.Row
		for i, f := range files {
			rd := table.RowData{
				tui.ColumnID:           fmt.Sprintf("%3d", i+1),
				tui.FileColumnDDName:   f.Ddname,
				tui.FileColumnStepName: f.Stepname,
				tui.FileColumnProcStep: f.Procstep,
				tui.FileColumnFID:      f.ID,
				tui.JobColumnJobID:     f.Jobid,
				tui.JobColumnClass:     f.Class,
				tui.FileColumnRC:       fmt.Sprintf("%v", f.RecordCount),
				tui.FileColumnBC:       fmt.Sprintf("%v", f.ByteCount),
				tui.FileColumnLrecl:    fmt.Sprintf("%v", f.Lrecl),
				tui.FileLink:           f.RecordsURL,
			}
			rows = append(rows, table.NewRow(rd))
		}
		return jobFilesMsg(rows)
	}
}

func fetchSpoolContent(recordsURL string) tea.Cmd {
	return func() tea.Msg {
		if recordsURL == "" {
			return spoolContentMsg(nil)
		}
		client := Profile.NewZosmfClient()
		// recordsURL is a full URL from z/OSMF; extract the path after /zosmf
		path := recordsURL
		if idx := strings.Index(recordsURL, "/zosmf"); idx >= 0 {
			path = recordsURL[idx+len("/zosmf"):]
		}
		resp, err := client.Get(path, nil)
		if err != nil || zosmf.CheckResponse(resp, 200) != nil {
			return spoolContentMsg(nil)
		}
		lines := strings.Split(resp.BodyString(), "\n")
		var rows []table.Row
		for i, line := range lines {
			rd := table.RowData{
				tui.BrowserColumnID:  fmt.Sprintf("%5d", i+1),
				tui.BrowserColumnRec: line,
			}
			rows = append(rows, table.NewRow(rd))
		}
		return spoolContentMsg(rows)
	}
}

func doJobAction(jobname, jobid, action string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restjobs/jobs/%s/%s", jobname, jobid)
		payload := map[string]string{"request": action, "version": "2.0"}
		resp, err := client.Put(path, payload, nil)
		if err != nil {
			return jobActionMsg{jobname: jobname, jobid: jobid, ok: false, message: err.Error()}
		}
		if zosmf.CheckResponse(resp, 200, 202) != nil {
			return jobActionMsg{jobname: jobname, jobid: jobid, ok: false, message: resp.BodyString()}
		}
		return jobActionMsg{jobname: jobname, jobid: jobid, ok: true}
	}
}

// ---------- Table builders ----------

func buildJobTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateJobColumns()
	return tui.TableLayout(cols, rows, keys, false, true, true, 15)
}

func buildFileTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateFileColumns()
	return tui.TableLayout(cols, rows, keys, false, false, false, 15)
}

func buildSpoolBrowser(ddname string, rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateFileBrowserColumns(ddname)
	return tui.TableLayout(cols, rows, keys, false, false, false, 30)
}

func str(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// ---------- Helper: build job path with optional JES name ----------

func jobPath(jesName string) string {
	if jesName != "" {
		return fmt.Sprintf("/restjobs/jobs/-%s", jesName)
	}
	return "/restjobs/jobs"
}

// ---------- Cobra commands ----------

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage z/OS Batch Jobs / Started Tasks / TSO Users",
	Long: `
DESCRIPTION
-----------
Manage z/OS jobs/stcs and/or tso users and their associated spool files on the JES spool
queues. Get lists of jobs and use actions like cancel, delete, hold, release and submit.`,
}

var jobsListCmd = &cobra.Command{
	Use:   "ls",
	Short: "Get a list of z/OS JES Jobs, STCs and TSO Users",
	Long: `
DESCRIPTION
-----------
zcli jobs ls gets the state of jobs, stcs and/or TSO users from z/OS.
You can specify filter criteria such as prefix, owner or type. If used
with the --tui option zcli will present the resulting job list in an easy
to read table. In TUI mode action commands like cancel, hold and release
are available using key combinations.`,
	Example: `
  zcli jobs ls --tui --prefix 'TEST*'
  zcli jobs ls --prefix 'TEST*' --status active --ojob`,
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		prefix, _ := cmd.Flags().GetString("prefix")
		status, _ := cmd.Flags().GetString("status")
		maxJobs, _ := cmd.Flags().GetInt("max-jobs")

		jtype := "*"
		if v, _ := cmd.Flags().GetBool("ojob"); v {
			jtype = "JOB"
		} else if v, _ := cmd.Flags().GetBool("ostc"); v {
			jtype = "STC"
		} else if v, _ := cmd.Flags().GetBool("otsu"); v {
			jtype = "TSU"
		}

		if v, _ := cmd.Flags().GetBool("tui"); v {
			if _, err := tea.NewProgram(newJobModel(owner, prefix, status, jtype)).Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

		// Non-TUI: JSON output
		client := Profile.NewZosmfClient()
		query := fmt.Sprintf("?owner=%s&prefix=%s&status=%s&exec-data=y&max-jobs=%d", owner, prefix, status, maxJobs)
		resp, err := client.Get("/restjobs/jobs"+query, nil)
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

var jobsDDNamesCmd = &cobra.Command{
	Use:   "ddnames",
	Short: "List spool file DD names of a job",
	Long: `
DESCRIPTION
-----------
Get spool file DD Names of a job and return them as JSON.`,
	Example: `
  zcli jobs ddnames --job-name TESTJOB --job-id JOB00030
  zcli jobs ddnames --job-correlator J0003456`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()
		path, err := resolveJobFilesPath(cmd)
		if err != nil {
			return err
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

var jobsFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Retrieve spool file content",
	Long: `
DESCRIPTION
-----------
Retrieve the content of a specific spool file by its file ID.`,
	Example: `
  zcli jobs files --job-name TESTJOB --job-id JOB00030 --file-id 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileID, _ := cmd.Flags().GetInt("file-id")
		client := Profile.NewZosmfClient()
		basePath, err := resolveJobIdentPath(cmd)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/files/%d/records", basePath, fileID)
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

var jobsJCLCmd = &cobra.Command{
	Use:   "jcl",
	Short: "Retrieve the JCL of a job",
	Long: `
DESCRIPTION
-----------
Retrieve the submitted JCL for a job.`,
	Example: `
  zcli jobs jcl --job-name TESTJOB --job-id JOB00030`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := Profile.NewZosmfClient()
		basePath, err := resolveJobIdentPath(cmd)
		if err != nil {
			return err
		}
		path := basePath + "/files/JCL/records"
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

var jobsSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit JCL to z/OS for execution",
	Long: `
DESCRIPTION
-----------
Submits a job to z/OS and returns Jobname and JobID.`,
	Example: `
  zcli jobs submit --file-name /u/jobs/testjob.jcl
  zcli jobs submit --file-name /u/jobs/testjob.jcl --secondary-jes JES3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, _ := cmd.Flags().GetString("file-name")
		jesName, _ := cmd.Flags().GetString("secondary-jes")
		inline, _ := cmd.Flags().GetBool("inline")

		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("error reading JCL file %s: %w", fileName, err)
		}

		client := Profile.NewZosmfClient()
		path := jobPath(jesName)

		headers := map[string]string{}
		if inline {
			headers["Content-Type"] = "text/plain"
		}

		resp, err := client.PutRaw(path, data, headers)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 201); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}

		var result struct {
			Jobname string `json:"jobname"`
			Jobid   string `json:"jobid"`
		}
		if err := resp.JSON(&result); err != nil {
			fmt.Println(resp.BodyString())
		} else {
			fmt.Printf("Job %s submitted, jobID is %s\n", result.Jobname, result.Jobid)
		}
		return nil
	},
}

var jobsHoldCmd = &cobra.Command{
	Use:   "hold",
	Short: "Hold a job on z/OS",
	Long: `
DESCRIPTION
-----------
Hold a job on z/OS using the z/OSMF REST API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runJobModify(cmd, "hold")
	},
}

var jobsReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a held job on z/OS",
	Long: `
DESCRIPTION
-----------
Release a held job on z/OS using the z/OSMF REST API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runJobModify(cmd, "release")
	},
}

var jobsCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a job on z/OS",
	Long: `
DESCRIPTION
-----------
Cancel a job currently executing on z/OS. With --purge the job output is also deleted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		purge, _ := cmd.Flags().GetBool("purge")
		jesName, _ := cmd.Flags().GetString("secondary-jes")
		client := Profile.NewZosmfClient()

		if purge {
			basePath, err := resolveJobIdentPath(cmd)
			if err != nil {
				return err
			}
			path := basePath
			if jesName != "" {
				// Rebuild path with JES name prefix
				path = rebuildPathWithJES(basePath, jesName)
			}
			headers := map[string]string{"X-IBM-Job-Modify-Version": "2.0"}
			resp, err := client.Delete(path, headers)
			if err != nil {
				return err
			}
			if apiErr := zosmf.CheckResponse(resp, 200, 202); apiErr != nil {
				fmt.Fprintln(os.Stderr, apiErr)
				os.Exit(8)
			}
			fmt.Println(resp.BodyString())
		} else {
			return runJobModify(cmd, "cancel")
		}
		return nil
	},
}

var jobsChangeClassCmd = &cobra.Command{
	Use:   "change-class",
	Short: "Change the job class of a job",
	Long: `
DESCRIPTION
-----------
Change the JES job class of an existing job.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		newClass, _ := cmd.Flags().GetString("new-class")
		jesName, _ := cmd.Flags().GetString("secondary-jes")
		client := Profile.NewZosmfClient()
		basePath, err := resolveJobIdentPath(cmd)
		if err != nil {
			return err
		}
		path := basePath
		if jesName != "" {
			path = rebuildPathWithJES(basePath, jesName)
		}
		payload := map[string]string{"class": newClass, "version": "2.0"}
		resp, err := client.Put(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 202); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println(resp.BodyString())
		return nil
	},
}

// ---------- Helpers ----------

func runJobModify(cmd *cobra.Command, action string) error {
	jesName, _ := cmd.Flags().GetString("secondary-jes")
	client := Profile.NewZosmfClient()
	basePath, err := resolveJobIdentPath(cmd)
	if err != nil {
		return err
	}
	path := basePath
	if jesName != "" {
		path = rebuildPathWithJES(basePath, jesName)
	}
	payload := map[string]string{"request": action, "version": "2.0"}
	resp, err := client.Put(path, payload, nil)
	if err != nil {
		return err
	}
	if apiErr := zosmf.CheckResponse(resp, 200, 202); apiErr != nil {
		fmt.Fprintln(os.Stderr, apiErr)
		os.Exit(8)
	}
	fmt.Println(resp.BodyString())
	return nil
}

func resolveJobIdentPath(cmd *cobra.Command) (string, error) {
	correlator, _ := cmd.Flags().GetString("job-correlator")
	if correlator != "" {
		return fmt.Sprintf("/restjobs/jobs/%s", correlator), nil
	}
	jobName, _ := cmd.Flags().GetString("job-name")
	jobID, _ := cmd.Flags().GetString("job-id")
	if jobName == "" || jobID == "" {
		return "", fmt.Errorf("either --job-correlator or both --job-name and --job-id are required")
	}
	return fmt.Sprintf("/restjobs/jobs/%s/%s", jobName, jobID), nil
}

func resolveJobFilesPath(cmd *cobra.Command) (string, error) {
	basePath, err := resolveJobIdentPath(cmd)
	if err != nil {
		return "", err
	}
	return basePath + "/files", nil
}

func rebuildPathWithJES(basePath, jesName string) string {
	// Insert -jesName after /restjobs/jobs/
	return strings.Replace(basePath, "/restjobs/jobs/", fmt.Sprintf("/restjobs/jobs/-%s", jesName), 1)
}

// ---------- Common flags for job-name/job-id/job-correlator ----------

func addJobIdentFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("job-name", "", "", "z/OS JES job name")
	cmd.Flags().StringP("job-id", "", "", "z/OS JES job ID")
	cmd.Flags().StringP("job-correlator", "", "", "z/OS JES job correlator")
	cmd.MarkFlagsMutuallyExclusive("job-correlator", "job-name")
	cmd.MarkFlagsMutuallyExclusive("job-correlator", "job-id")
}

func addJESFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("secondary-jes", "", "", "Secondary JES subsystem name")
}

// ---------- Unused import guard for json ----------
var _ = json.Marshal

// ---------- init ----------

func init() {
	// ls flags
	jobsListCmd.Flags().StringP("owner", "o", "*", "Owner of the requested z/OS JES jobs")
	jobsListCmd.Flags().StringP("prefix", "p", "*", "Job name prefix to search for")
	jobsListCmd.Flags().StringP("status", "s", "*", "Status filter (e.g. ACTIVE)")
	jobsListCmd.Flags().IntP("max-jobs", "m", 1000, "Maximum number of jobs returned")
	jobsListCmd.Flags().BoolP("ojob", "b", false, "Display z/OS batch jobs only")
	jobsListCmd.Flags().BoolP("ostc", "c", false, "Display z/OS started tasks only")
	jobsListCmd.Flags().BoolP("otsu", "u", false, "Display z/OS TSO users only")
	jobsListCmd.MarkFlagsMutuallyExclusive("ojob", "ostc", "otsu")
	jobsListCmd.Flags().BoolP("tui", "t", false, "Activate terminal user interface")

	// ddnames flags
	addJobIdentFlags(jobsDDNamesCmd)

	// files flags
	addJobIdentFlags(jobsFilesCmd)
	jobsFilesCmd.Flags().IntP("file-id", "i", 0, "Spool file ID to retrieve")
	jobsFilesCmd.MarkFlagRequired("file-id")

	// jcl flags
	addJobIdentFlags(jobsJCLCmd)

	// submit flags
	jobsSubmitCmd.Flags().StringP("file-name", "f", "", "File containing JCL to submit")
	jobsSubmitCmd.MarkFlagRequired("file-name")
	addJESFlag(jobsSubmitCmd)
	jobsSubmitCmd.Flags().Bool("inline", true, "Submit JCL inline (text/plain)")

	// hold flags
	addJobIdentFlags(jobsHoldCmd)
	addJESFlag(jobsHoldCmd)

	// release flags
	addJobIdentFlags(jobsReleaseCmd)
	addJESFlag(jobsReleaseCmd)

	// cancel flags
	addJobIdentFlags(jobsCancelCmd)
	addJESFlag(jobsCancelCmd)
	jobsCancelCmd.Flags().Bool("purge", false, "Purge output of cancelled job")

	// change-class flags
	addJobIdentFlags(jobsChangeClassCmd)
	addJESFlag(jobsChangeClassCmd)
	jobsChangeClassCmd.Flags().StringP("new-class", "n", "", "New JES job class")
	jobsChangeClassCmd.MarkFlagRequired("new-class")

	// Wire up
	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsDDNamesCmd)
	jobsCmd.AddCommand(jobsFilesCmd)
	jobsCmd.AddCommand(jobsJCLCmd)
	jobsCmd.AddCommand(jobsSubmitCmd)
	jobsCmd.AddCommand(jobsHoldCmd)
	jobsCmd.AddCommand(jobsReleaseCmd)
	jobsCmd.AddCommand(jobsCancelCmd)
	jobsCmd.AddCommand(jobsChangeClassCmd)

	rootCmd.AddCommand(jobsCmd)
}
