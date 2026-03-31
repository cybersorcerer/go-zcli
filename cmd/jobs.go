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

		case "enter":
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

		case "f3", "esc":
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
		m.updateSpoolFooter()
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
		"Pg. %d/%d | enter select | ctrl+r refresh | ctrl+l cancel | ctrl+h hold | ctrl+e release | ↑/↓ move | F7/F8 page | ctrl+c quit",
		m.jobTable.CurrentPage(), m.jobTable.MaxPages(),
	))
}

func (m *jobModel) updateFileFooter() {
	m.fileTable = m.fileTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d [%s/%s] | enter select | F3 back | ↑/↓ move | F7/F8 page | ctrl+c quit",
		m.fileTable.CurrentPage(), m.fileTable.MaxPages(), m.jobname, m.jobid,
	))
}

func (m *jobModel) updateSpoolFooter() {
	m.spoolBrowser = m.spoolBrowser.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d [%s] | F3 back | ↑/↓ move | G/g bottom/top | F7/F8 page | ctrl+c quit",
		m.spoolBrowser.CurrentPage(), m.spoolBrowser.MaxPages(), m.ddname,
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
	Use:   "list",
	Short: "Get a list of z/OS JES Jobs, STCs and TSO Users",
	Long: `
DESCRIPTION
-----------
zcli jobs list gets the state of jobs, stcs and/or TSO users from z/OS.
You can specify filter criteria such as prefix, owner or type. If used
with the --tui option zcli will present the resulting job list in an easy
to read table. In TUI mode action commands like cancel, hold and release
are available using key combinations.`,
	Example: `
  zcli jobs list --tui --prefix 'TEST*'
  zcli jobs list --prefix 'TEST*' --status active --ojob`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		prefix, _ := cmd.Flags().GetString("prefix")
		status, _ := cmd.Flags().GetString("status")
		maxJobs, _ := cmd.Flags().GetInt("max-jobs")
		jobID, _ := cmd.Flags().GetString("jobid")
		userCorrelator, _ := cmd.Flags().GetString("user-correlator")
		execData, _ := cmd.Flags().GetString("exec-data")
		secondaryJES, _ := cmd.Flags().GetString("secondary-jes")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

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

		// Build query parameters
		var params []string
		params = append(params, "owner="+owner)
		params = append(params, "prefix="+prefix)
		if status != "" {
			params = append(params, "status="+status)
		}
		params = append(params, fmt.Sprintf("max-jobs=%d", maxJobs))
		if execData != "" {
			params = append(params, "exec-data="+execData)
		}
		if jobID != "" {
			params = append(params, "jobid="+jobID)
		}
		if userCorrelator != "" {
			params = append(params, "user-correlator="+userCorrelator)
		}

		// Build path with optional secondary JES
		basePath := "/restjobs/jobs"
		if secondaryJES != "" {
			basePath = fmt.Sprintf("/restjobs/jobs/-%s", secondaryJES)
		}
		path := basePath + "?" + strings.Join(params, "&")

		// Build headers
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

		client := Profile.NewZosmfClient()
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

var jobsDDNamesCmd = &cobra.Command{
	Use:   "ddnames",
	Short: "List spool files for a job",
	Long: `
DESCRIPTION
-----------
List the spool files (DD names) for a job and return them as JSON.`,
	Example: `
  zcli jobs ddnames --job-name TESTJOB --job-id JOB00030
  zcli jobs ddnames --job-correlator J0003456
  zcli jobs ddnames --job-name TESTJOB --job-id JOB00030 --secondary-jes JES3`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		secondaryJES, _ := cmd.Flags().GetString("secondary-jes")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		basePath, err := resolveJobFilesPath(cmd)
		if err != nil {
			return err
		}
		if secondaryJES != "" {
			basePath = rebuildPathWithJES(basePath, secondaryJES)
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

		client := Profile.NewZosmfClient()
		resp, err := client.Get(basePath, headers)
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
	Use:   "spool",
	Short: "Retrieve spool file content or JCL",
	Long: `
DESCRIPTION
-----------
Retrieve the content of a specific spool file by its file ID, or the submitted
JCL by specifying --jcl. Supports text/binary/record modes, search, and
record range filtering.`,
	Example: `
  zcli jobs spool --job-name TESTJOB --job-id JOB00030 --file-id 3
  zcli jobs spool --job-name TESTJOB --job-id JOB00030 --jcl
  zcli jobs spool --job-name TESTJOB --job-id JOB00030 --file-id 3 --record-range 0-99
  zcli jobs spool --job-name TESTJOB --job-id JOB00030 --file-id 3 --search 'IEF403I'
  zcli jobs spool --job-name TESTJOB --job-id JOB00030 --file-id 3 --mode binary`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileID, _ := cmd.Flags().GetInt("file-id")
		jcl, _ := cmd.Flags().GetBool("jcl")
		mode, _ := cmd.Flags().GetString("mode")
		fileEncoding, _ := cmd.Flags().GetString("file-encoding")
		search, _ := cmd.Flags().GetString("search")
		research, _ := cmd.Flags().GetString("research")
		insensitive, _ := cmd.Flags().GetString("insensitive")
		maxReturnSize, _ := cmd.Flags().GetString("max-return-size")
		recordRange, _ := cmd.Flags().GetString("record-range")
		secondaryJES, _ := cmd.Flags().GetString("secondary-jes")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		basePath, err := resolveJobIdentPath(cmd)
		if err != nil {
			return err
		}
		if secondaryJES != "" {
			basePath = rebuildPathWithJES(basePath, secondaryJES)
		}

		// Build file path
		var filePart string
		if jcl {
			filePart = "/files/JCL/records"
		} else {
			filePart = fmt.Sprintf("/files/%d/records", fileID)
		}

		// Build query parameters
		var params []string
		if mode != "" {
			params = append(params, "mode="+mode)
		}
		if fileEncoding != "" {
			params = append(params, "fileEncoding="+fileEncoding)
		}
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

		path := basePath + filePart
		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		// Build headers
		headers := make(map[string]string)
		if recordRange != "" {
			headers["X-IBM-Record-Range"] = recordRange
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

var jobsSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit JCL to z/OS for execution",
	Long: `
DESCRIPTION
-----------
Submits a job to z/OS and returns Jobname and JobID. You can submit JCL
inline from a local file (--file-name), or reference a data set or UNIX
file on the host (--remote-file). When using --remote-file, specify a
fully-qualified data set as "//'HLQ.JCL(MBR)'" or a UNIX path as
"/u/myjobs/job1".

JCL symbols can be passed with --jcl-symbol NAME=VALUE (repeatable).`,
	Example: `
  zcli jobs submit --file-name /u/jobs/testjob.jcl
  zcli jobs submit --file-name /u/jobs/testjob.jcl --intrdr-class B
  zcli jobs submit --remote-file "//'MYJOBS.TEST.CNTL(TESTJOBX)'"
  zcli jobs submit --remote-file /u/myjobs/job1 --intrdr-mode TEXT
  zcli jobs submit --file-name job.jcl --jcl-symbol MBR=ABC --jcl-symbol ENV=PROD
  zcli jobs submit --file-name job.jcl --notification-url https://hook.example.com`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName, _ := cmd.Flags().GetString("file-name")
		remoteFile, _ := cmd.Flags().GetString("remote-file")
		jesName, _ := cmd.Flags().GetString("secondary-jes")
		intrdrClass, _ := cmd.Flags().GetString("intrdr-class")
		intrdrRecfm, _ := cmd.Flags().GetString("intrdr-recfm")
		intrdrLrecl, _ := cmd.Flags().GetString("intrdr-lrecl")
		intrdrMode, _ := cmd.Flags().GetString("intrdr-mode")
		intrdrEncoding, _ := cmd.Flags().GetString("intrdr-file-encoding")
		userCorrelator, _ := cmd.Flags().GetString("user-correlator")
		jclSymbols, _ := cmd.Flags().GetStringArray("jcl-symbol")
		notifURL, _ := cmd.Flags().GetString("notification-url")
		notifOpts, _ := cmd.Flags().GetString("notification-options")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		if fileName == "" && remoteFile == "" {
			return fmt.Errorf("either --file-name (local JCL) or --remote-file (host dataset/UNIX path) is required")
		}
		if fileName != "" && remoteFile != "" {
			return fmt.Errorf("--file-name and --remote-file are mutually exclusive")
		}

		client := Profile.NewZosmfClient()
		path := jobPath(jesName)
		headers := map[string]string{}

		// Custom headers
		if intrdrClass != "" {
			headers["X-IBM-Intrdr-Class"] = intrdrClass
		}
		if intrdrRecfm != "" {
			headers["X-IBM-Intrdr-Recfm"] = intrdrRecfm
		}
		if intrdrLrecl != "" {
			headers["X-IBM-Intrdr-Lrecl"] = intrdrLrecl
		}
		if intrdrMode != "" {
			headers["X-IBM-Intrdr-Mode"] = intrdrMode
		}
		if intrdrEncoding != "" {
			headers["X-IBM-Intrdr-File-Encoding"] = intrdrEncoding
		}
		if userCorrelator != "" {
			headers["X-IBM-User-Correlator"] = userCorrelator
		}
		for _, sym := range jclSymbols {
			parts := strings.SplitN(sym, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --jcl-symbol format %q, expected NAME=VALUE", sym)
			}
			headers["X-IBM-JCL-Symbol-"+parts[0]] = parts[1]
		}
		if notifURL != "" {
			headers["X-IBM-Notification-URL"] = notifURL
		}
		if notifOpts != "" {
			headers["X-IBM-Notification-Options"] = notifOpts
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

		var resp *zosmf.Response
		var err error

		if remoteFile != "" {
			// Submit from host dataset or UNIX file via JSON document
			headers["Content-Type"] = "application/json"
			payload := map[string]interface{}{
				"file": remoteFile,
			}
			resp, err = client.Put(path, payload, headers)
		} else {
			// Submit inline JCL from local file
			data, readErr := os.ReadFile(fileName)
			if readErr != nil {
				return fmt.Errorf("error reading JCL file %s: %w", fileName, readErr)
			}

			mode := strings.ToUpper(intrdrMode)
			if mode == "RECORD" || mode == "BINARY" {
				headers["Content-Type"] = "application/octet-stream"
			} else {
				headers["Content-Type"] = "text/plain"
			}

			resp, err = client.PutRaw(path, data, headers)
		}
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
Hold a job that has been submitted but not yet selected for processing.
When held, a job is not eligible for selection. Identify the job by
--job-name/--job-id or --job-correlator.

By default synchronous processing (version 2.0) is used. Use --async
for asynchronous processing (version 1.0), which returns HTTP 202 only.`,
	Args: cobra.NoArgs,
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
	Args: cobra.NoArgs,
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		purge, _ := cmd.Flags().GetBool("purge")
		jesName, _ := cmd.Flags().GetString("secondary-jes")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")
		client := Profile.NewZosmfClient()

		if purge {
			basePath, err := resolveJobIdentPath(cmd)
			if err != nil {
				return err
			}
			path := basePath
			if jesName != "" {
				path = rebuildPathWithJES(basePath, jesName)
			}
			headers := map[string]string{"X-IBM-Job-Modify-Version": "2.0"}
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
Change the JES job class of an existing job. Identify the job by
--job-name/--job-id or --job-correlator. The specified job class is
not validated on input; verify with "zcli jobs list" afterwards.

By default synchronous processing (version 2.0) is used. Use --async
for asynchronous processing (version 1.0), which returns HTTP 202 only.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		newClass, _ := cmd.Flags().GetString("new-class")
		jesName, _ := cmd.Flags().GetString("secondary-jes")
		async, _ := cmd.Flags().GetBool("async")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		basePath, err := resolveJobIdentPath(cmd)
		if err != nil {
			return err
		}
		path := basePath
		if jesName != "" {
			path = rebuildPathWithJES(basePath, jesName)
		}

		version := "2.0"
		if async {
			version = "1.0"
		}
		payload := map[string]string{"class": newClass, "version": version}

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
	async, _ := cmd.Flags().GetBool("async")
	targetSystem, _ := cmd.Flags().GetString("target-system")
	targetUser, _ := cmd.Flags().GetString("target-user")
	targetPassword, _ := cmd.Flags().GetString("target-password")

	client := Profile.NewZosmfClient()
	basePath, err := resolveJobIdentPath(cmd)
	if err != nil {
		return err
	}
	path := basePath
	if jesName != "" {
		path = rebuildPathWithJES(basePath, jesName)
	}

	version := "2.0"
	if async {
		version = "1.0"
	}
	payload := map[string]string{"request": action, "version": version}

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
	// list flags
	jobsListCmd.Flags().StringP("owner", "o", "*", "Job owner (default: z/OS user ID, max 8 chars, wildcards allowed).")
	jobsListCmd.Flags().StringP("prefix", "p", "*", "Job name prefix (max 8 chars, wildcards allowed).")
	jobsListCmd.Flags().StringP("status", "s", "", "Status filter: ACTIVE to limit to active jobs.")
	jobsListCmd.Flags().IntP("max-jobs", "m", 1000, "Maximum jobs returned (1-1000).")
	jobsListCmd.Flags().String("jobid", "", "Job ID filter (max 8 chars, mutually exclusive with --user-correlator).")
	jobsListCmd.Flags().String("user-correlator", "", "User portion of job correlator (1-32 chars, JES2 only).")
	jobsListCmd.Flags().String("exec-data", "", "Return execution data: Y or N.")
	jobsListCmd.Flags().BoolP("ojob", "b", false, "Display z/OS batch jobs only.")
	jobsListCmd.Flags().BoolP("ostc", "c", false, "Display z/OS started tasks only.")
	jobsListCmd.Flags().BoolP("otsu", "u", false, "Display z/OS TSO users only.")
	jobsListCmd.MarkFlagsMutuallyExclusive("ojob", "ostc", "otsu")
	jobsListCmd.MarkFlagsMutuallyExclusive("jobid", "user-correlator")
	jobsListCmd.Flags().BoolP("tui", "t", false, "Activate terminal user interface.")
	addJESFlag(jobsListCmd)
	jobsListCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsListCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsListCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// ddnames flags (list spool files)
	addJobIdentFlags(jobsDDNamesCmd)
	addJESFlag(jobsDDNamesCmd)
	jobsDDNamesCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsDDNamesCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsDDNamesCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// spool flags (retrieve spool content)
	addJobIdentFlags(jobsFilesCmd)
	addJESFlag(jobsFilesCmd)
	jobsFilesCmd.Flags().IntP("file-id", "i", 0, "Spool file ID to retrieve.")
	jobsFilesCmd.Flags().Bool("jcl", false, "Retrieve submitted JCL instead of spool file.")
	jobsFilesCmd.Flags().String("mode", "", "Conversion mode: text (default), binary, or record.")
	jobsFilesCmd.Flags().String("file-encoding", "", "EBCDIC code page for spool file (default: IBM-1047, text mode only).")
	jobsFilesCmd.Flags().String("search", "", "Search for first record containing this string (text mode only).")
	jobsFilesCmd.Flags().String("research", "", "Search using extended regular expression (text mode only).")
	jobsFilesCmd.Flags().String("insensitive", "", "Case insensitive search: true (default) or false.")
	jobsFilesCmd.Flags().String("max-return-size", "", "Max records to return for search/research (default 100).")
	jobsFilesCmd.Flags().String("record-range", "", "Record range: SSS-EEE or SSS,NNN (0-based).")
	jobsFilesCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsFilesCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsFilesCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// submit flags
	jobsSubmitCmd.Flags().StringP("file-name", "f", "", "Local file containing JCL to submit inline.")
	jobsSubmitCmd.Flags().StringP("remote-file", "r", "", "Host dataset or UNIX file path (e.g. \"//'HLQ.JCL(MBR)'\" or \"/u/jobs/job1\").")
	addJESFlag(jobsSubmitCmd)
	jobsSubmitCmd.Flags().String("intrdr-class", "", "Internal reader class (single char, default A). Defines default MSGCLASS.")
	jobsSubmitCmd.Flags().String("intrdr-recfm", "", "Internal reader record format: F or V.")
	jobsSubmitCmd.Flags().String("intrdr-lrecl", "", "Internal reader logical record length (default 80).")
	jobsSubmitCmd.Flags().String("intrdr-mode", "", "Input job format: TEXT, RECORD, or BINARY (default TEXT).")
	jobsSubmitCmd.Flags().String("intrdr-file-encoding", "", "EBCDIC code page for internal reader (default IBM-1047).")
	jobsSubmitCmd.Flags().String("user-correlator", "", "User portion of the job correlator (1-32 chars).")
	jobsSubmitCmd.Flags().StringArrayP("jcl-symbol", "s", nil, "JCL symbol as NAME=VALUE (repeatable, up to 128).")
	jobsSubmitCmd.Flags().String("notification-url", "", "URL to receive HTTP POST on job events.")
	jobsSubmitCmd.Flags().String("notification-options", "", "JSON string specifying job events, e.g. '{\"events\":[\"active\",\"complete\"]}'.")
	jobsSubmitCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsSubmitCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsSubmitCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// hold flags
	addJobIdentFlags(jobsHoldCmd)
	addJESFlag(jobsHoldCmd)
	jobsHoldCmd.Flags().Bool("async", false, "Use asynchronous processing (version 1.0).")
	jobsHoldCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsHoldCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsHoldCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// release flags
	addJobIdentFlags(jobsReleaseCmd)
	addJESFlag(jobsReleaseCmd)
	jobsReleaseCmd.Flags().Bool("async", false, "Use asynchronous processing (version 1.0).")
	jobsReleaseCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsReleaseCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsReleaseCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// cancel flags
	addJobIdentFlags(jobsCancelCmd)
	addJESFlag(jobsCancelCmd)
	jobsCancelCmd.Flags().Bool("purge", false, "Purge output of cancelled job")
	jobsCancelCmd.Flags().Bool("async", false, "Use asynchronous processing (version 1.0).")
	jobsCancelCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsCancelCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsCancelCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// change-class flags
	addJobIdentFlags(jobsChangeClassCmd)
	addJESFlag(jobsChangeClassCmd)
	jobsChangeClassCmd.Flags().StringP("new-class", "n", "", "New JES job class")
	jobsChangeClassCmd.MarkFlagRequired("new-class")
	jobsChangeClassCmd.Flags().Bool("async", false, "Use asynchronous processing (version 1.0).")
	jobsChangeClassCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	jobsChangeClassCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	jobsChangeClassCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// Wire up
	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsDDNamesCmd)
	jobsCmd.AddCommand(jobsFilesCmd)
	jobsCmd.AddCommand(jobsSubmitCmd)
	jobsCmd.AddCommand(jobsHoldCmd)
	jobsCmd.AddCommand(jobsReleaseCmd)
	jobsCmd.AddCommand(jobsCancelCmd)
	jobsCmd.AddCommand(jobsChangeClassCmd)

	rootCmd.AddCommand(jobsCmd)
}
