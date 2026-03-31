package cmd

import (
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

type subsysListMsg []table.Row
type subsysErrorMsg struct{ err error }

// ---------- Bubbletea model ----------

type subsysModel struct {
	subsysTable table.Model
	statusMsg   string
}

func newSubsysModel() subsysModel {
	return subsysModel{
		subsysTable: buildSubsysTable(nil),
	}
}

func (m subsysModel) Init() tea.Cmd {
	return fetchSubsysList("")
}

func (m subsysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.subsysTable = m.subsysTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)

	case subsysErrorMsg:
		m.statusMsg = fmt.Sprintf("  Error: %v", msg.err)

	case subsysListMsg:
		m.subsysTable = m.subsysTable.WithRows([]table.Row(msg))
		m.statusMsg = fmt.Sprintf("  %d subsystems loaded", len(msg))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+r":
			cmds = append(cmds, fetchSubsysList(""))
		}
	}

	m.updateSubsysFooter()
	m.subsysTable, cmd = m.subsysTable.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m subsysModel) View() string {
	var b strings.Builder
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg + "\n")
	}
	pad := lipgloss.NewStyle().Padding(0)
	b.WriteString(pad.Render(m.subsysTable.View()))
	return b.String()
}

func (m *subsysModel) updateSubsysFooter() {
	m.subsysTable = m.subsysTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d | ctrl+r refresh | ↑/↓ move | F7/F8 page | ctrl+c quit",
		m.subsysTable.CurrentPage(), m.subsysTable.MaxPages(),
	))
}

// ---------- Bubbletea commands ----------

func fetchSubsysList(ssid string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := "/rest/mvssubs"
		if ssid != "" {
			path = fmt.Sprintf("/rest/mvssubs?ssid=%s", ssid)
		}
		resp, err := client.Get(path, nil)
		if err != nil {
			return subsysErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			return subsysErrorMsg{err: apiErr}
		}

		var result struct {
			Items []struct {
				Subsys  string `json:"subsys"`
				Active  bool   `json:"active"`
				Primary bool   `json:"primary"`
				Dynamic bool   `json:"dynamic"`
				Funcs   []int  `json:"funcs"`
			} `json:"items"`
		}
		if err := resp.JSON(&result); err != nil {
			return subsysErrorMsg{err: fmt.Errorf("JSON parse error: %w (body: %.200s)", err, resp.BodyString())}
		}

		var rows []table.Row
		for i, ss := range result.Items {
			funcs := make([]string, len(ss.Funcs))
			for j, f := range ss.Funcs {
				funcs[j] = fmt.Sprintf("%d", f)
			}
			rd := table.RowData{
				tui.ColumnID:            fmt.Sprintf("%3d", i+1),
				tui.ColumnName:          ss.Subsys,
				tui.SubsysColumnActive:  boolToStr(ss.Active),
				tui.SubsysColumnPrimary: boolToStr(ss.Primary),
				tui.SubsysColumnDynamic: boolToStr(ss.Dynamic),
				tui.SubsysColumnFuncs:   strings.Join(funcs, ","),
			}
			rows = append(rows, table.NewRow(rd))
		}
		return subsysListMsg(rows)
	}
}

func boolToStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// ---------- Table builder ----------

func buildSubsysTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateSubsysColumns()
	return tui.TableLayout(cols, rows, keys, false, true, true, 15)
}

// ---------- Cobra commands ----------

var subsystemsCmd = &cobra.Command{
	Use:   "subsystems",
	Short: "List z/OS subsystems",
	Long: `
DESCRIPTION
-----------
This service lists the subsystems on a z/OS system.`,
}

var subsystemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get information about z/OS subsystems",
	Long: `
DESCRIPTION
-----------
You can use the list subcommand to get information about the
subsystems on a z/OS system. You can filter the returned list
of subsystems by specifying a subsystem id or wild-card.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		ssid, _ := cmd.Flags().GetString("ssid")

		if useTUI {
			if _, err := tea.NewProgram(newSubsysModel()).Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

		client := Profile.NewZosmfClient()
		path := "/rest/mvssubs"
		if ssid != "" {
			path = fmt.Sprintf("/rest/mvssubs?ssid=%s", ssid)
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
	subsystemsListCmd.Flags().StringP("ssid", "s", "", "Subsystem ID or wildcard filter, if empty all subsystems are returned.")
	subsystemsListCmd.Flags().BoolP("tui", "t", false, "Activate terminal user interface")

	subsystemsCmd.AddCommand(subsystemsListCmd)
	rootCmd.AddCommand(subsystemsCmd)
}
