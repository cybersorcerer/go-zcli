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

type mfsListMsg []table.Row
type mfsErrorMsg struct{ err error }
type mfsDetailsMsg struct {
	data map[string]interface{}
}

// ---------- Bubbletea model ----------

type mfsModel struct {
	mfsTable     table.Model
	detailTable  table.Model
	statusMsg    string
	view         string // "list" or "details"
	selectedName string
}

func newMfsModel() mfsModel {
	return mfsModel{
		mfsTable:    buildMfsTable(nil),
		detailTable: buildMfsDetailTable(nil),
		view:        "list",
	}
}

func (m mfsModel) Init() tea.Cmd {
	return fetchMfsList("", "")
}

func (m mfsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.mfsTable = m.mfsTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)
		m.detailTable = m.detailTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)

	case mfsErrorMsg:
		m.statusMsg = fmt.Sprintf("  Error: %v", msg.err)

	case mfsListMsg:
		m.mfsTable = m.mfsTable.WithRows([]table.Row(msg))
		m.statusMsg = fmt.Sprintf("  %d filesystems loaded", len(msg))

	case mfsDetailsMsg:
		rows := buildDetailRows(msg.data)
		m.detailTable = m.detailTable.WithRows(rows)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+r":
			if m.view == "list" {
				cmds = append(cmds, fetchMfsList("", ""))
			}

		case "ctrl+s", "enter":
			if m.view == "list" {
				row := m.mfsTable.HighlightedRow()
				m.selectedName = fmt.Sprintf("%v", row.Data[tui.ColumnName])
				m.view = "details"
				m.mfsTable = m.mfsTable.Focused(false)
				m.detailTable = m.detailTable.Focused(true)
				cmds = append(cmds, fetchMfsDetails(m.selectedName))
			}

		case "ctrl+p", "esc":
			if m.view == "details" {
				m.view = "list"
				m.detailTable = m.detailTable.Focused(false)
				m.mfsTable = m.mfsTable.Focused(true)
			}
		}
	}

	switch m.view {
	case "list":
		m.updateMfsFooter()
		m.mfsTable, cmd = m.mfsTable.Update(msg)
	case "details":
		m.detailTable, cmd = m.detailTable.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m mfsModel) View() string {
	var b strings.Builder
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg + "\n")
	}
	pad := lipgloss.NewStyle().Padding(0)
	switch m.view {
	case "list":
		b.WriteString(pad.Render(m.mfsTable.View()))
	case "details":
		b.WriteString(pad.Render(m.detailTable.View()))
	}
	return b.String()
}

func (m *mfsModel) updateMfsFooter() {
	m.mfsTable = m.mfsTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d | ctrl+c quit | ctrl+r refresh | ctrl+s select | ↑/↓ move | f7/f8 page",
		m.mfsTable.CurrentPage(), m.mfsTable.MaxPages(),
	))
}

// ---------- Bubbletea commands ----------

func fetchMfsList(path, fsname string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		query := ""
		if path != "" {
			query = "?path=" + path
		} else if fsname != "" {
			query = "?fsname=" + fsname
		}
		resp, err := client.Get("/restfiles/mfs"+query, nil)
		if err != nil {
			return mfsErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			return mfsErrorMsg{err: apiErr}
		}

		var result struct {
			Items []struct {
				Name       string `json:"name"`
				Mountpoint string `json:"mountpoint"`
				FSTname    string `json:"fstname"`
				Status     string `json:"status"`
				Mode       []string `json:"mode"`
				Dev        int    `json:"dev"`
				Sysname    string `json:"sysname"`
				Readibc    int    `json:"readibc"`
				Writeibc   int    `json:"writeibc"`
				Diribc     int    `json:"diribc"`
			} `json:"items"`
		}
		if err := resp.JSON(&result); err != nil {
			return mfsErrorMsg{err: fmt.Errorf("JSON parse error: %w (body: %.200s)", err, resp.BodyString())}
		}

		var rows []table.Row
		for i, fs := range result.Items {
			rd := table.RowData{
				tui.ColumnID:              fmt.Sprintf("%3d", i+1),
				tui.MfscolumnFSMountPoint: fs.Mountpoint,
				tui.ColumnName:            fs.Name,
				tui.MfscolumnFSType:       fs.FSTname,
				tui.MfscolumnFSStatus:     fs.Status,
			}
			rows = append(rows, table.NewRow(rd))
		}
		return mfsListMsg(rows)
	}
}

func fetchMfsDetails(fsname string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		resp, err := client.Get("/restfiles/mfs?fsname="+fsname, nil)
		if err != nil || zosmf.CheckResponse(resp, 200) != nil {
			return mfsDetailsMsg{}
		}
		var data map[string]interface{}
		if err := resp.JSON(&data); err != nil {
			return mfsDetailsMsg{}
		}
		return mfsDetailsMsg{data: data}
	}
}

func buildDetailRows(data map[string]interface{}) []table.Row {
	var rows []table.Row
	if data == nil {
		return rows
	}
	items, ok := data["items"]
	if !ok {
		return rows
	}
	itemList, ok := items.([]interface{})
	if !ok || len(itemList) == 0 {
		return rows
	}
	fs, ok := itemList[0].(map[string]interface{})
	if !ok {
		return rows
	}
	i := 0
	for k, v := range fs {
		i++
		rd := table.RowData{
			tui.ColumnID:   fmt.Sprintf("%3d", i),
			tui.ColumnName: k,
			tui.ColumnType: fmt.Sprintf("%v", v),
		}
		rows = append(rows, table.NewRow(rd))
	}
	return rows
}

// ---------- Table builders ----------

func buildMfsTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateMfsColumns()
	return tui.TableLayout(cols, rows, keys, false, true, true, 15)
}

func buildMfsDetailTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateDetailsMfsColumns()
	return tui.TableLayout(cols, rows, keys, false, true, true, 15)
}

// ---------- Cobra commands ----------

var mfsCmd = &cobra.Command{
	Use:   "mfs",
	Short: "Manage z/OS UNIX mounted filesystems",
	Long: `
DESCRIPTION
-----------
You can use the mfs command to list all mounted filesystems,
or a specific filesystem mounted at a given path, or the path
mounted to a given filesystem name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		mfsPath, _ := cmd.Flags().GetString("path")
		fsname, _ := cmd.Flags().GetString("fsname")

		if useTUI {
			if _, err := tea.NewProgram(newMfsModel()).Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

		// Non-TUI: JSON output
		client := Profile.NewZosmfClient()
		query := ""
		if mfsPath != "" {
			query = "?path=" + mfsPath
		} else if fsname != "" {
			query = "?fsname=" + fsname
		}
		resp, err := client.Get("/restfiles/mfs"+query, nil)
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

// Unused import guard for json
var _ = json.Marshal

func init() {
	mfsCmd.Flags().StringP("path", "p", "", "UNIX directory path to filter by")
	mfsCmd.Flags().StringP("fsname", "f", "", "Fully qualified filesystem name to filter by")
	mfsCmd.Flags().BoolP("tui", "t", false, "Activate terminal user interface")
	mfsCmd.MarkFlagsMutuallyExclusive("path", "fsname")

	rootCmd.AddCommand(mfsCmd)
}
