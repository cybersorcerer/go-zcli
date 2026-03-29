package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"zcli/internal/tui"
	"zcli/internal/zosmf"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/spf13/cobra"
)

// ---------- Bubbletea messages ----------

type fsErrorMsg struct{ err error }
type fsActionMsg struct{ msg string }

// ---------- Filesystem item ----------

type fsItem struct {
	Name       string   `json:"name"`
	Mountpoint string   `json:"mountpoint"`
	FSTname    string   `json:"fstname"`
	Status     string   `json:"status"`
	Mode       []string `json:"mode"`
	Dev        int      `json:"dev"`
	Sysname    string   `json:"sysname"`
	Readibc    int      `json:"readibc"`
	Writeibc   int      `json:"writeibc"`
	Diribc     int      `json:"diribc"`
}

// ---------- Bubbletea model ----------

type fsView int

const (
	fsViewList fsView = iota
	fsViewDetails
	fsViewCreate
	fsViewConfirmCreate
	fsViewConfirmDelete
	fsViewConfirmUnmount
)

type fsModel struct {
	fsTable      table.Model
	detailTable  table.Model
	statusMsg    string
	view         fsView
	selectedName string
	selectedMP   string
	items        []fsItem

	// create form
	createInputs []textinput.Model
	createFocus  int
	createName   string
	createPri    string
	createSec    string
	createOwner  string
	createGroup  string
	createPerms  string
	createSC     string
	createMC     string
	createDC     string
	createVols   string
	createTmout  string

	// confirm dialog
	confirmYes bool
}

func newFsModel() fsModel {
	return fsModel{
		fsTable:     buildFsTable(nil),
		detailTable: buildFsDetailTable(nil),
		view:        fsViewList,
	}
}

func (m fsModel) Init() tea.Cmd {
	return fetchFsList()
}

func (m fsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.fsTable = m.fsTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)
		m.detailTable = m.detailTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)

	case fsErrorMsg:
		m.statusMsg = fmt.Sprintf("  Error: %v", msg.err)

	case fsListWithItemsMsg:
		m.fsTable = m.fsTable.WithRows(msg.rows)
		m.items = msg.items
		m.statusMsg = fmt.Sprintf("  %d filesystems loaded", len(msg.rows))

	case fsActionMsg:
		m.statusMsg = "  " + msg.msg
		m.view = fsViewList
		m.fsTable = m.fsTable.Focused(true)
		cmds = append(cmds, fetchFsList())

	case tea.KeyMsg:
		switch m.view {
		case fsViewList:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+r":
				cmds = append(cmds, fetchFsList())
			case "enter":
				row := m.fsTable.HighlightedRow()
				name := fmt.Sprintf("%v", row.Data[tui.ColumnName])
				m.selectedName = name
				m.selectedMP = fmt.Sprintf("%v", row.Data[tui.MfscolumnFSMountPoint])
				for _, item := range m.items {
					if item.Name == name {
						m.detailTable = m.detailTable.WithRows(buildFsDetailRows(item))
						break
					}
				}
				m.view = fsViewDetails
				m.fsTable = m.fsTable.Focused(false)
				m.detailTable = m.detailTable.Focused(true)
				m.statusMsg = fmt.Sprintf("  %s", name)
			case "ctrl+u":
				row := m.fsTable.HighlightedRow()
				name := fmt.Sprintf("%v", row.Data[tui.ColumnName])
				if name != "" {
					m.selectedName = name
					m.view = fsViewConfirmUnmount
					m.confirmYes = false
					m.fsTable = m.fsTable.Focused(false)
				}
			case "ctrl+x":
				row := m.fsTable.HighlightedRow()
				name := fmt.Sprintf("%v", row.Data[tui.ColumnName])
				if name != "" {
					m.selectedName = name
					m.view = fsViewConfirmDelete
					m.confirmYes = false
					m.fsTable = m.fsTable.Focused(false)
				}
			case "ctrl+n":
				m.view = fsViewCreate
				m.fsTable = m.fsTable.Focused(false)
				m.createInputs = makeFsCreateInputs()
				m.createFocus = 0
				m.createInputs[0].Focus()
			}

		case fsViewDetails:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3":
				m.view = fsViewList
				m.detailTable = m.detailTable.Focused(false)
				m.fsTable = m.fsTable.Focused(true)
			}

		case fsViewCreate:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3":
				m.view = fsViewList
				m.fsTable = m.fsTable.Focused(true)
			case "tab", "down":
				m.createInputs[m.createFocus].Blur()
				m.createFocus = (m.createFocus + 1) % len(m.createInputs)
				m.createInputs[m.createFocus].Focus()
			case "shift+tab", "up":
				m.createInputs[m.createFocus].Blur()
				m.createFocus = (m.createFocus - 1 + len(m.createInputs)) % len(m.createInputs)
				m.createInputs[m.createFocus].Focus()
			case "enter":
				name := m.createInputs[0].Value()
				cylsPri := m.createInputs[1].Value()
				if name == "" || cylsPri == "" {
					m.statusMsg = "  Name and Primary Cylinders are required"
				} else {
					m.createName = name
					m.createPri = cylsPri
					m.createSec = m.createInputs[2].Value()
					m.createOwner = m.createInputs[3].Value()
					m.createGroup = m.createInputs[4].Value()
					m.createPerms = m.createInputs[5].Value()
					m.createSC = m.createInputs[6].Value()
					m.createMC = m.createInputs[7].Value()
					m.createDC = m.createInputs[8].Value()
					m.createVols = m.createInputs[9].Value()
					m.createTmout = m.createInputs[10].Value()
					m.view = fsViewConfirmCreate
					m.confirmYes = false
				}
			default:
				for i := range m.createInputs {
					m.createInputs[i], cmd = m.createInputs[i].Update(msg)
					cmds = append(cmds, cmd)
				}
			}

		case fsViewConfirmDelete:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3", "n", "N":
				m.view = fsViewList
				m.fsTable = m.fsTable.Focused(true)
				m.statusMsg = "  Delete cancelled"
			case "left", "right", "h", "l", "tab":
				m.confirmYes = !m.confirmYes
			case "y", "Y":
				m.confirmYes = true
				cmds = append(cmds, doFsDelete(m.selectedName))
			case "enter":
				if m.confirmYes {
					cmds = append(cmds, doFsDelete(m.selectedName))
				} else {
					m.view = fsViewList
					m.fsTable = m.fsTable.Focused(true)
					m.statusMsg = "  Delete cancelled"
				}
			}

		case fsViewConfirmCreate:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3", "n", "N":
				m.view = fsViewCreate
				m.statusMsg = "  Create cancelled"
			case "left", "right", "h", "l", "tab":
				m.confirmYes = !m.confirmYes
			case "y", "Y":
				m.confirmYes = true
				cmds = append(cmds, doFsCreate(m.fsCreateParams()))
			case "enter":
				if m.confirmYes {
					cmds = append(cmds, doFsCreate(m.fsCreateParams()))
				} else {
					m.view = fsViewCreate
					m.statusMsg = "  Create cancelled"
				}
			}

		case fsViewConfirmUnmount:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3", "n", "N":
				m.view = fsViewList
				m.fsTable = m.fsTable.Focused(true)
				m.statusMsg = "  Unmount cancelled"
			case "left", "right", "h", "l", "tab":
				m.confirmYes = !m.confirmYes
			case "y", "Y":
				m.confirmYes = true
				cmds = append(cmds, doFsUnmount(m.selectedName))
			case "enter":
				if m.confirmYes {
					cmds = append(cmds, doFsUnmount(m.selectedName))
				} else {
					m.view = fsViewList
					m.fsTable = m.fsTable.Focused(true)
					m.statusMsg = "  Unmount cancelled"
				}
			}
		}
	}

	switch m.view {
	case fsViewList:
		m.updateFsFooter()
		m.fsTable, cmd = m.fsTable.Update(msg)
	case fsViewDetails:
		m.updateFsDetailFooter()
		m.detailTable, cmd = m.detailTable.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m fsModel) View() string {
	var b strings.Builder
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg + "\n")
	}
	pad := lipgloss.NewStyle().Padding(0)
	switch m.view {
	case fsViewList:
		b.WriteString(pad.Render(m.fsTable.View()))
	case fsViewDetails:
		b.WriteString(pad.Render(m.detailTable.View()))
	case fsViewCreate:
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#71fff1")).
			Padding(1, 2)
		var form strings.Builder
		form.WriteString("Create zFS Filesystem (* required)\n\n")
		labels := []string{
			"* Dataset Name:      ",
			"* Primary Cylinders: ",
			"  Secondary Cyl.:    ",
			"  Owner:             ",
			"  Group:             ",
			"  Permissions:       ",
			"  Storage Class:     ",
			"  Management Class:  ",
			"  Data Class:        ",
			"  Volumes:           ",
			"  Timeout:           ",
		}
		for i, input := range m.createInputs {
			form.WriteString(labels[i] + input.View() + "\n")
		}
		form.WriteString("\n[enter] create  [tab/↑/↓] navigate  [F3] cancel")
		b.WriteString(style.Render(form.String()))
	case fsViewConfirmDelete:
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#ff5555")).
			Padding(1, 2)
		yesStyle := lipgloss.NewStyle()
		noStyle := lipgloss.NewStyle()
		if m.confirmYes {
			yesStyle = yesStyle.Foreground(lipgloss.Color("#ff5555")).Bold(true).Underline(true)
		} else {
			noStyle = noStyle.Foreground(lipgloss.Color("#71fff1")).Bold(true).Underline(true)
		}
		var dialog strings.Builder
		dialog.WriteString(fmt.Sprintf("Delete filesystem %q?\n", m.selectedName))
		dialog.WriteString("The filesystem must not be mounted.\n\n")
		dialog.WriteString("  " + yesStyle.Render("[Y]es") + "    " + noStyle.Render("[N]o"))
		dialog.WriteString("\n\n[y/n] or [left/right] + [enter]")
		b.WriteString(style.Render(dialog.String()))
	case fsViewConfirmCreate:
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#71fff1")).
			Padding(1, 2)
		yesStyle := lipgloss.NewStyle()
		noStyle := lipgloss.NewStyle()
		if m.confirmYes {
			yesStyle = yesStyle.Foreground(lipgloss.Color("#00a65a")).Bold(true).Underline(true)
		} else {
			noStyle = noStyle.Foreground(lipgloss.Color("#71fff1")).Bold(true).Underline(true)
		}
		var dialog strings.Builder
		dialog.WriteString("Create zFS filesystem?\n\n")
		dialog.WriteString(fmt.Sprintf("  Dataset:          %s\n", m.createName))
		dialog.WriteString(fmt.Sprintf("  Primary Cyl.:     %s\n", m.createPri))
		optionals := []struct{ label, val string }{
			{"Secondary Cyl.", m.createSec},
			{"Owner", m.createOwner},
			{"Group", m.createGroup},
			{"Permissions", m.createPerms},
			{"Storage Class", m.createSC},
			{"Management Class", m.createMC},
			{"Data Class", m.createDC},
			{"Volumes", m.createVols},
			{"Timeout", m.createTmout},
		}
		for _, o := range optionals {
			if o.val != "" {
				dialog.WriteString(fmt.Sprintf("  %-18s%s\n", o.label+":", o.val))
			}
		}
		dialog.WriteString("\n  " + yesStyle.Render("[Y]es") + "    " + noStyle.Render("[N]o"))
		dialog.WriteString("\n\n[y/n] or [left/right] + [enter]")
		b.WriteString(style.Render(dialog.String()))
	case fsViewConfirmUnmount:
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFF066")).
			Padding(1, 2)
		yesStyle := lipgloss.NewStyle()
		noStyle := lipgloss.NewStyle()
		if m.confirmYes {
			yesStyle = yesStyle.Foreground(lipgloss.Color("#FFF066")).Bold(true).Underline(true)
		} else {
			noStyle = noStyle.Foreground(lipgloss.Color("#71fff1")).Bold(true).Underline(true)
		}
		var dialog strings.Builder
		dialog.WriteString(fmt.Sprintf("Unmount filesystem %q?\n\n", m.selectedName))
		dialog.WriteString("  " + yesStyle.Render("[Y]es") + "    " + noStyle.Render("[N]o"))
		dialog.WriteString("\n\n[y/n] or [left/right] + [enter]")
		b.WriteString(style.Render(dialog.String()))
	}
	return b.String()
}

func (m *fsModel) updateFsFooter() {
	m.fsTable = m.fsTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d | enter details | ctrl+n create | ctrl+x delete | ctrl+u unmount | ctrl+r refresh | ↑/↓ move | F7/F8 page | ctrl+c quit",
		m.fsTable.CurrentPage(), m.fsTable.MaxPages(),
	))
}

func (m *fsModel) updateFsDetailFooter() {
	m.detailTable = m.detailTable.WithStaticFooter(fmt.Sprintf(
		"[%s] | F3 back | ↑/↓ move | F7/F8 page | ctrl+c quit",
		m.selectedName,
	))
}

// ---------- Bubbletea commands ----------

func fetchFsList() tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		resp, err := client.Get("/restfiles/mfs", nil)
		if err != nil {
			return fsErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			return fsErrorMsg{err: apiErr}
		}

		var result struct {
			Items []fsItem `json:"items"`
		}
		if err := resp.JSON(&result); err != nil {
			return fsErrorMsg{err: fmt.Errorf("JSON parse error: %w", err)}
		}

		var rows []table.Row
		for i, fs := range result.Items {
			modeStr := strings.Join(fs.Mode, ",")
			rd := table.RowData{
				tui.ColumnID:              fmt.Sprintf("%3d", i+1),
				tui.MfscolumnFSMountPoint: fs.Mountpoint,
				tui.ColumnName:            fs.Name,
				tui.MfscolumnFSType:       fs.FSTname,
				tui.MfscolumnFSStatus:     fs.Status,
				"Mode":                    modeStr,
			}
			row := table.NewRow(rd)
			if strings.Contains(modeStr, "rdwr") {
				row = row.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#00a65a")))
			}
			rows = append(rows, row)
		}

		// Store items in model via a closure trick - we return both rows and items
		// Actually we need to store items. Use a combined message.
		return fsListWithItemsMsg{rows: rows, items: result.Items}
	}
}

type fsListWithItemsMsg struct {
	rows  []table.Row
	items []fsItem
}

type fsCreateOpts struct {
	name    string
	cylsPri string
	cylsSec string
	owner   string
	group   string
	perms   string
	sc      string
	mc      string
	dc      string
	volumes string
	timeout string
}

func (m fsModel) fsCreateParams() fsCreateOpts {
	return fsCreateOpts{
		name:    m.createName,
		cylsPri: m.createPri,
		cylsSec: m.createSec,
		owner:   m.createOwner,
		group:   m.createGroup,
		perms:   m.createPerms,
		sc:      m.createSC,
		mc:      m.createMC,
		dc:      m.createDC,
		volumes: m.createVols,
		timeout: m.createTmout,
	}
}

func doFsCreate(opts fsCreateOpts) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/zfs/%s", opts.name)
		if opts.timeout != "" {
			path += "?timeout=" + opts.timeout
		}

		payload := map[string]interface{}{
			"cylsPri": fsAtoi(opts.cylsPri),
		}
		if opts.cylsSec != "" {
			payload["cylsSec"] = fsAtoi(opts.cylsSec)
		}
		if opts.owner != "" {
			payload["owner"] = opts.owner
		}
		if opts.group != "" {
			payload["group"] = opts.group
		}
		if opts.perms != "" {
			payload["perms"] = fsAtoi(opts.perms)
		}
		if opts.sc != "" {
			payload["storageClass"] = opts.sc
		}
		if opts.mc != "" {
			payload["managementClass"] = opts.mc
		}
		if opts.dc != "" {
			payload["dataClass"] = opts.dc
		}
		if opts.volumes != "" {
			payload["volumes"] = strings.Split(opts.volumes, ",")
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return fsErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 201); apiErr != nil {
			return fsErrorMsg{err: apiErr}
		}
		return fsActionMsg{msg: fmt.Sprintf("Created %s", opts.name)}
	}
}

func doFsDelete(fsname string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/zfs/%s", fsname)
		resp, err := client.Delete(path, nil)
		if err != nil {
			return fsErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 204); apiErr != nil {
			return fsErrorMsg{err: apiErr}
		}
		return fsActionMsg{msg: fmt.Sprintf("Deleted %s", fsname)}
	}
}

func doFsUnmount(fsname string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		payload := map[string]interface{}{
			"action": "unmount",
		}
		resp, err := client.Put(fmt.Sprintf("/restfiles/mfs/%s", fsname), payload, nil)
		if err != nil {
			return fsErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 200, 204); apiErr != nil {
			return fsErrorMsg{err: apiErr}
		}
		return fsActionMsg{msg: fmt.Sprintf("Unmounted %s", fsname)}
	}
}

// ---------- Detail rows ----------

func buildFsDetailRows(fs fsItem) []table.Row {
	details := []struct {
		key   string
		value string
	}{
		{"Name", fs.Name},
		{"Mountpoint", fs.Mountpoint},
		{"FS Type", fs.FSTname},
		{"Status", fs.Status},
		{"Mode", strings.Join(fs.Mode, ", ")},
		{"Device", fmt.Sprintf("%d", fs.Dev)},
		{"System", fs.Sysname},
		{"Read I/O", fmt.Sprintf("%d", fs.Readibc)},
		{"Write I/O", fmt.Sprintf("%d", fs.Writeibc)},
		{"Dir I/O", fmt.Sprintf("%d", fs.Diribc)},
	}

	var rows []table.Row
	for i, d := range details {
		rd := table.RowData{
			tui.ColumnID:   fmt.Sprintf("%3d", i+1),
			tui.ColumnName: d.key,
			tui.ColumnType: d.value,
		}
		rows = append(rows, table.NewRow(rd))
	}
	return rows
}

// ---------- Table builders ----------

func buildFsTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateMfsColumns()
	return tui.TableLayout(cols, rows, keys, false, true, true, 15)
}

func buildFsDetailTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := []table.Column{
		table.NewColumn(tui.ColumnID, tui.ColumnID, 3),
		table.NewColumn(tui.ColumnName, "Property", 15),
		table.NewColumn(tui.ColumnType, "Value", 60),
	}
	return tui.TableLayout(cols, rows, keys, false, false, true, 15)
}

// ---------- Create form inputs ----------

func makeFsCreateInputs() []textinput.Model {
	newInput := func(placeholder string, charLimit, width int) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.CharLimit = charLimit
		ti.Width = width
		return ti
	}

	return []textinput.Model{
		newInput("e.g. USER.ZFS.AGGR01", 44, 44),          // 0: Dataset Name *
		newInput("e.g. 10", 6, 10),                         // 1: Primary Cylinders *
		newInput("optional", 6, 10),                         // 2: Secondary Cylinders
		newInput("optional, owner user ID", 8, 20),          // 3: Owner
		newInput("optional, group owner", 8, 20),            // 4: Group
		newInput("optional, e.g. 755", 4, 10),               // 5: Perms
		newInput("optional, DFSMS storage class", 30, 30),   // 6: Storage Class
		newInput("optional, DFSMS management class", 30, 30),// 7: Management Class
		newInput("optional, DFSMS data class", 30, 30),      // 8: Data Class
		newInput("optional, e.g. VOL001,VOL002", 60, 44),   // 9: Volumes
		newInput("optional, seconds (default 20)", 6, 10),   // 10: Timeout
	}
}

func fsAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// ---------- Cobra commands ----------

var filesystemsCmd = &cobra.Command{
	Use:   "filesystems",
	Short: "Interact with z/OS z/Unix filesystems",
	Long: `
DESCRIPTION
-----------
Interact with z/OS z/Unix filesystems. List, create, delete, mount, and unmount
z/OS UNIX file systems. Use --tui with the list subcommand for an interactive view
with unmount support.`,
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
		perms, _ := cmd.Flags().GetInt("perms")
		storageClass, _ := cmd.Flags().GetString("storage-class")
		managementClass, _ := cmd.Flags().GetString("management-class")
		dataClass, _ := cmd.Flags().GetString("data-class")
		cylsPri, _ := cmd.Flags().GetInt("primary-cylinders")
		cylsSec, _ := cmd.Flags().GetInt("secondary-cylinders")
		volumes, _ := cmd.Flags().GetStringSlice("volumes")
		timeout, _ := cmd.Flags().GetString("timeout")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/zfs/%s", zfsName)
		if timeout != "" {
			path += "?timeout=" + timeout
		}

		payload := map[string]interface{}{
			"cylsPri": cylsPri,
		}
		if cylsSec > 0 {
			payload["cylsSec"] = cylsSec
		}
		if owner != "" {
			payload["owner"] = owner
		}
		if group != "" {
			payload["group"] = group
		}
		if cmd.Flags().Changed("perms") {
			payload["perms"] = perms
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

var filesystemsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a zfs z/UNIX filesystem",
	Long: `
DESCRIPTION
-----------
The command will delete a zfs filesystem.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		zfsName, _ := cmd.Flags().GetString("zfs-dataset-name")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/zfs/%s", zfsName)

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

		resp, err := client.Delete(path, headers)
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
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

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
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/restfiles/mfs/%s", fsName)
		payload := map[string]interface{}{
			"action": "unmount",
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
List all mounted filesystems, or the specific filesystem mounted at a given path,
or the filesystem with a given filesystem name. Use --tui for an interactive view
with details and unmount support.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fsName, _ := cmd.Flags().GetString("fs-dataset-name")
		pathName, _ := cmd.Flags().GetString("path")
		maxItems, _ := cmd.Flags().GetString("max-items")
		targetSystem, _ := cmd.Flags().GetString("target-system")
		targetUser, _ := cmd.Flags().GetString("target-user")
		targetPassword, _ := cmd.Flags().GetString("target-password")

		if v, _ := cmd.Flags().GetBool("tui"); v {
			if _, err := tea.NewProgram(newFsModel()).Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

		client := Profile.NewZosmfClient()
		var path string
		if fsName != "" {
			path = fmt.Sprintf("/restfiles/mfs/?fsname=%s", fsName)
		} else if pathName != "" {
			path = fmt.Sprintf("/restfiles/mfs/?path=%s", pathName)
		} else {
			path = "/restfiles/mfs"
		}

		headers := make(map[string]string)
		if maxItems != "" {
			headers["X-IBM-Max-Items"] = maxItems
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

func init() {
	// filesystems create
	filesystemsCreateCmd.Flags().StringP("zfs-dataset-name", "z", "", "The zFS aggregate (VSAM linear dataset) name.")
	filesystemsCreateCmd.MarkFlagRequired("zfs-dataset-name")
	filesystemsCreateCmd.Flags().StringP("owner", "o", "", "Owner user ID (default: 755).")
	filesystemsCreateCmd.Flags().StringP("group", "g", "", "Group owner (default: 755).")
	filesystemsCreateCmd.Flags().Int("perms", 755, "Permissions (default: 755).")
	filesystemsCreateCmd.Flags().String("storage-class", "", "z/OS DFSMS storage class.")
	filesystemsCreateCmd.Flags().String("management-class", "", "z/OS DFSMS management class.")
	filesystemsCreateCmd.Flags().String("data-class", "", "z/OS DFSMS data class.")
	filesystemsCreateCmd.Flags().Int("primary-cylinders", 0, "Primary space in cylinders.")
	filesystemsCreateCmd.MarkFlagRequired("primary-cylinders")
	filesystemsCreateCmd.Flags().Int("secondary-cylinders", 0, "Secondary space in cylinders.")
	filesystemsCreateCmd.Flags().StringSlice("volumes", nil, "Volume serial numbers.")
	filesystemsCreateCmd.Flags().String("timeout", "", "Seconds to wait for zfsadm format (default: 20).")
	filesystemsCreateCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesystemsCreateCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesystemsCreateCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// filesystems delete
	filesystemsDeleteCmd.Flags().StringP("zfs-dataset-name", "z", "", "The zFS aggregate (VSAM linear dataset) name.")
	filesystemsDeleteCmd.MarkFlagRequired("zfs-dataset-name")
	filesystemsDeleteCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesystemsDeleteCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesystemsDeleteCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// filesystems mount
	filesystemsMountCmd.Flags().StringP("fs-dataset-name", "f", "", "The filesystem name to mount.")
	filesystemsMountCmd.MarkFlagRequired("fs-dataset-name")
	filesystemsMountCmd.Flags().String("fs-type", "zfs", "The file system type.")
	filesystemsMountCmd.Flags().StringP("mount-point", "t", "", "The mount point directory.")
	filesystemsMountCmd.MarkFlagRequired("mount-point")
	filesystemsMountCmd.Flags().StringP("mode", "m", "rdonly", "Mount mode (e.g. rdonly, rdwr).")
	filesystemsMountCmd.Flags().Bool("setuid", false, "Use setuid option in mount mode.")
	filesystemsMountCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesystemsMountCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesystemsMountCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// filesystems unmount
	filesystemsUnmountCmd.Flags().StringP("fs-dataset-name", "f", "", "The filesystem name to unmount.")
	filesystemsUnmountCmd.MarkFlagRequired("fs-dataset-name")
	filesystemsUnmountCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesystemsUnmountCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesystemsUnmountCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// filesystems list
	filesystemsListCmd.Flags().StringP("fs-dataset-name", "f", "", "Filesystem name to list (fsname query param).")
	filesystemsListCmd.Flags().StringP("path", "p", "", "Mount path to list (path query param).")
	filesystemsListCmd.Flags().String("max-items", "", "Maximum items to return (0 = all, default 1000).")
	filesystemsListCmd.Flags().BoolP("tui", "t", false, "Activate terminal user interface.")
	filesystemsListCmd.Flags().String("target-system", "", "Target system nickname for cross-system request.")
	filesystemsListCmd.Flags().String("target-user", "", "User ID for target system authentication.")
	filesystemsListCmd.Flags().String("target-password", "", "Password for target system authentication.")

	// Wire up
	filesystemsCmd.AddCommand(filesystemsCreateCmd)
	filesystemsCmd.AddCommand(filesystemsDeleteCmd)
	filesystemsCmd.AddCommand(filesystemsMountCmd)
	filesystemsCmd.AddCommand(filesystemsUnmountCmd)
	filesystemsCmd.AddCommand(filesystemsListCmd)

	rootCmd.AddCommand(filesystemsCmd)
}
