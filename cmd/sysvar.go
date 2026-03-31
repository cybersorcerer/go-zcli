package cmd

import (
	"fmt"
	"os"
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

type sysvarListMsg []table.Row
type sysvarErrorMsg struct{ err error }
type sysvarActionDoneMsg struct{ info string }

// ---------- Bubbletea model ----------

type sysvarView int

const (
	sysvarViewList sysvarView = iota
	sysvarViewCreate
	sysvarViewConfirmDelete
)

type sysvarModel struct {
	varTable     table.Model
	statusMsg    string
	view         sysvarView
	plexName     string
	systemName   string
	selectedName string

	// create form
	createInputs []textinput.Model
	createFocus  int

	// confirm delete
	confirmYes bool
}

func newSysvarModel(plexName, systemName string) sysvarModel {
	return sysvarModel{
		varTable:   buildSysvarTable(nil),
		view:       sysvarViewList,
		plexName:   plexName,
		systemName: systemName,
	}
}

func (m sysvarModel) Init() tea.Cmd {
	return fetchSysvarList(m.plexName, m.systemName)
}

func (m sysvarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.varTable = m.varTable.WithMaxTotalWidth(msg.Width).WithPageSize(msg.Height - 10)

	case sysvarErrorMsg:
		m.statusMsg = fmt.Sprintf("  Error: %v", msg.err)

	case sysvarListMsg:
		m.varTable = m.varTable.WithRows([]table.Row(msg))
		m.statusMsg = fmt.Sprintf("  %d variables loaded", len(msg))

	case sysvarActionDoneMsg:
		m.statusMsg = "  " + msg.info
		m.view = sysvarViewList
		m.varTable = m.varTable.Focused(true)
		cmds = append(cmds, fetchSysvarList(m.plexName, m.systemName))

	case tea.KeyMsg:
		switch m.view {
		case sysvarViewList:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+r":
				cmds = append(cmds, fetchSysvarList(m.plexName, m.systemName))
			case "ctrl+n":
				m.view = sysvarViewCreate
				m.varTable = m.varTable.Focused(false)
				m.createInputs = makeCreateInputs()
				m.createFocus = 0
				m.createInputs[0].Focus()
			case "ctrl+x":
				row := m.varTable.HighlightedRow()
				name := fmt.Sprintf("%v", row.Data[tui.ColumnName])
				if name != "" {
					m.selectedName = name
					m.view = sysvarViewConfirmDelete
					m.confirmYes = false
					m.varTable = m.varTable.Focused(false)
				}
			}

		case sysvarViewCreate:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3", "esc":
				m.view = sysvarViewList
				m.varTable = m.varTable.Focused(true)
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
				value := m.createInputs[1].Value()
				desc := m.createInputs[2].Value()
				if name == "" || value == "" {
					m.statusMsg = "  Name and Value are required"
				} else {
					cmds = append(cmds, doSysvarCreate(m.plexName, m.systemName, name, value, desc))
				}
			default:
				for i := range m.createInputs {
					m.createInputs[i], cmd = m.createInputs[i].Update(msg)
					cmds = append(cmds, cmd)
				}
			}

		case sysvarViewConfirmDelete:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "f3", "esc", "n", "N":
				m.view = sysvarViewList
				m.varTable = m.varTable.Focused(true)
				m.statusMsg = "  Delete cancelled"
			case "left", "right", "h", "l", "tab":
				m.confirmYes = !m.confirmYes
			case "y", "Y":
				m.confirmYes = true
				cmds = append(cmds, doSysvarDelete(m.plexName, m.systemName, m.selectedName))
			case "enter":
				if m.confirmYes {
					cmds = append(cmds, doSysvarDelete(m.plexName, m.systemName, m.selectedName))
				} else {
					m.view = sysvarViewList
					m.varTable = m.varTable.Focused(true)
					m.statusMsg = "  Delete cancelled"
				}
			}
		}
	}

	if m.view == sysvarViewList {
		m.updateSysvarFooter()
		m.varTable, cmd = m.varTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m sysvarModel) View() string {
	var b strings.Builder

	if m.statusMsg != "" {
		b.WriteString(m.statusMsg + "\n")
	}

	pad := lipgloss.NewStyle().Padding(0)

	switch m.view {
	case sysvarViewList:
		b.WriteString(pad.Render(m.varTable.View()))

	case sysvarViewCreate:
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#71fff1")).
			Padding(1, 2)

		var form strings.Builder
		form.WriteString("Create / Update Variable\n\n")
		labels := []string{"Name:        ", "Value:       ", "Description: "}
		for i, input := range m.createInputs {
			form.WriteString(labels[i] + input.View() + "\n")
		}
		form.WriteString("\n[enter] save  [tab] next field  [F3] cancel")
		b.WriteString(style.Render(form.String()))

	case sysvarViewConfirmDelete:
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
		dialog.WriteString(fmt.Sprintf("Delete variable %q?\n\n", m.selectedName))
		dialog.WriteString("  " + yesStyle.Render("[Y]es") + "    " + noStyle.Render("[N]o"))
		dialog.WriteString("\n\n[y/n] or [left/right] + [enter]")
		b.WriteString(style.Render(dialog.String()))
	}

	return b.String()
}

func (m *sysvarModel) updateSysvarFooter() {
	m.varTable = m.varTable.WithStaticFooter(fmt.Sprintf(
		"Pg. %d/%d | ctrl+n create | ctrl+x delete | ctrl+r refresh | ↑/↓ move | F7/F8 page | ctrl+c quit",
		m.varTable.CurrentPage(), m.varTable.MaxPages(),
	))
}

// ---------- Text inputs for create form ----------

func makeCreateInputs() []textinput.Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "Variable name"
	nameInput.CharLimit = 64
	nameInput.Width = 40

	valueInput := textinput.New()
	valueInput.Placeholder = "Variable value"
	valueInput.CharLimit = 256
	valueInput.Width = 40

	descInput := textinput.New()
	descInput.Placeholder = "Description (optional)"
	descInput.CharLimit = 256
	descInput.Width = 40

	return []textinput.Model{nameInput, valueInput, descInput}
}

// ---------- Bubbletea commands ----------

func fetchSysvarList(plexName, systemName string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s", plexName, systemName)

		resp, err := client.Get(path, nil)
		if err != nil {
			return sysvarErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 200); apiErr != nil {
			return sysvarErrorMsg{err: apiErr}
		}

		var result struct {
			VarList    []struct {
				Name        string `json:"name"`
				Value       string `json:"value"`
				Description string `json:"description"`
			} `json:"system-variable-list"`
			SymList []struct {
				Name        string `json:"name"`
				Value       string `json:"value"`
				Description string `json:"description"`
			} `json:"system-symbol-list"`
		}
		if err := resp.JSON(&result); err != nil {
			return sysvarErrorMsg{err: fmt.Errorf("JSON parse error: %w (body: %.200s)", err, resp.BodyString())}
		}

		items := result.VarList
		if len(items) == 0 {
			items = result.SymList
		}

		var rows []table.Row
		for i, v := range items {
			rd := table.RowData{
				tui.ColumnID:         fmt.Sprintf("%3d", i+1),
				tui.ColumnName:       v.Name,
				tui.SysvarColumnValue: v.Value,
				tui.SysvarColumnDesc:  v.Description,
			}
			rows = append(rows, table.NewRow(rd))
		}
		return sysvarListMsg(rows)
	}
}

func doSysvarCreate(plexName, systemName, name, value, desc string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s", plexName, systemName)

		entry := map[string]string{
			"name":  name,
			"value": value,
		}
		if desc != "" {
			entry["description"] = desc
		}
		payload := map[string]interface{}{
			"system-variable-list": []map[string]string{entry},
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return sysvarErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			return sysvarErrorMsg{err: apiErr}
		}
		return sysvarActionDoneMsg{info: fmt.Sprintf("Variable %q created/updated", name)}
	}
}

func doSysvarDelete(plexName, systemName, name string) tea.Cmd {
	return func() tea.Msg {
		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s", plexName, systemName)

		resp, err := client.DeleteWithBody(path, []string{name}, nil)
		if err != nil {
			return sysvarErrorMsg{err: err}
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			return sysvarErrorMsg{err: apiErr}
		}
		return sysvarActionDoneMsg{info: fmt.Sprintf("Variable %q deleted", name)}
	}
}

// ---------- Table builder ----------

func buildSysvarTable(rows []table.Row) table.Model {
	keys := tui.RemapTableKeys()
	cols := tui.GenerateSysvarColumns()
	return tui.TableLayout(cols, rows, keys, false, true, true, 15)
}

// ---------- Cobra commands ----------

var sysvarCmd = &cobra.Command{
	Use:   "sysvar",
	Short: "Interact with z/OSMF and System variables",
	Long: `
DESCRIPTION
-----------
Interact with the z/OSMF and System variables.`,
}

var sysvarGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve z/OSMF and system variables",
	Long: `
DESCRIPTION
-----------
Use this command to retrieve z/OSMF and system variables.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		plexName, _ := cmd.Flags().GetString("plex-name")
		systemName, _ := cmd.Flags().GetString("system-name")
		varNames, _ := cmd.Flags().GetStringArray("var-name")
		source, _ := cmd.Flags().GetString("source")

		if useTUI {
			if _, err := tea.NewProgram(newSysvarModel(plexName, systemName)).Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s", plexName, systemName)

		query := ""
		for _, v := range varNames {
			if query == "" {
				query = "?"
			} else {
				query += "&"
			}
			query += "var-name=" + v
		}
		if source != "" {
			if query == "" {
				query = "?"
			} else {
				query += "&"
			}
			query += "source=" + source
		}
		path += query

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

var sysvarCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create or update z/OSMF system variables",
	Long: `
DESCRIPTION
-----------
Use this command to create or update system variables in the
variable pool. If the pool does not exist, it will be created.
Variables are specified with --var name=value (repeatable).
Optional descriptions can be added with --desc (same order as --var).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		plexName, _ := cmd.Flags().GetString("plex-name")
		systemName, _ := cmd.Flags().GetString("system-name")
		vars, _ := cmd.Flags().GetStringArray("var")
		descs, _ := cmd.Flags().GetStringArray("desc")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s", plexName, systemName)

		var varList []map[string]string
		for i, v := range vars {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable format %q, expected name=value", v)
			}
			entry := map[string]string{
				"name":  parts[0],
				"value": parts[1],
			}
			if i < len(descs) {
				entry["description"] = descs[i]
			}
			varList = append(varList, entry)
		}

		var payload interface{}
		if len(varList) > 0 {
			payload = map[string]interface{}{
				"system-variable-list": varList,
			}
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println("System variables created/updated successfully.")
		return nil
	},
}

var sysvarImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import system variables from a CSV file",
	Long: `
DESCRIPTION
-----------
Use this command to import z/OSMF system variables from a
CSV file on the z/OS system. The file must contain variable
definitions in CSV format (name, value, description per row)
with no header row.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		plexName, _ := cmd.Flags().GetString("plex-name")
		systemName, _ := cmd.Flags().GetString("system-name")
		importFile, _ := cmd.Flags().GetString("file")
		localFile, _ := cmd.Flags().GetString("local")

		client := Profile.NewZosmfClient()

		if localFile != "" {
			data, err := os.ReadFile(localFile)
			if err != nil {
				return fmt.Errorf("error reading local file %s: %w", localFile, err)
			}
			uploadPath := fmt.Sprintf("/restfiles/fs%s", importFile)
			headers := map[string]string{
				"X-IBM-Data-Type": "text;fileEncoding=UTF-8",
				"Content-Type":    "text/plain;charset=UTF-8",
			}
			resp, err := client.PutRaw(uploadPath, data, headers)
			if err != nil {
				return fmt.Errorf("error uploading file to z/OS: %w", err)
			}
			if apiErr := zosmf.CheckResponse(resp, 201, 204); apiErr != nil {
				return fmt.Errorf("upload failed: %w", apiErr)
			}

			tagPayload := map[string]interface{}{
				"request": "chtag",
				"action":  "set",
				"type":    "text",
				"codeset": "ISO8859-1",
			}
			tagResp, err := client.Put(uploadPath, tagPayload, nil)
			if err != nil {
				return fmt.Errorf("error tagging file on z/OS: %w", err)
			}
			if apiErr := zosmf.CheckResponse(tagResp, 200); apiErr != nil {
				return fmt.Errorf("chtag failed: %w", apiErr)
			}

			fmt.Printf("Local file %s uploaded and tagged (ISO8859-1) to %s\n", localFile, importFile)
		}

		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s/actions/import", plexName, systemName)
		payload := map[string]string{
			"variables-import-file": importFile,
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println("System variables imported successfully.")
		return nil
	},
}

var sysvarExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export system variables to a CSV file",
	Long: `
DESCRIPTION
-----------
Use this command to export z/OSMF system variables to a CSV
file on the z/OS system. The exported file can be re-imported
with the import command.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		plexName, _ := cmd.Flags().GetString("plex-name")
		systemName, _ := cmd.Flags().GetString("system-name")
		exportFile, _ := cmd.Flags().GetString("file")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		localFile, _ := cmd.Flags().GetString("local")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s/actions/export", plexName, systemName)

		payload := map[string]interface{}{
			"variables-export-file": exportFile,
			"overwrite":             overwrite,
		}

		resp, err := client.Post(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println("System variables exported successfully.")

		if localFile != "" {
			downloadPath := fmt.Sprintf("/restfiles/fs%s", exportFile)
			headers := map[string]string{
				"X-IBM-Data-Type": "text;fileEncoding=UTF-8",
				"Content-Type":    "text/plain;charset=UTF-8",
			}
			dlResp, err := client.Get(downloadPath, headers)
			if err != nil {
				return fmt.Errorf("error downloading file from z/OS: %w", err)
			}
			if apiErr := zosmf.CheckResponse(dlResp, 200); apiErr != nil {
				return fmt.Errorf("download failed: %w", apiErr)
			}
			if err := os.WriteFile(localFile, dlResp.Body, 0644); err != nil {
				return fmt.Errorf("error writing local file %s: %w", localFile, err)
			}
			fmt.Printf("Downloaded %s to %s\n", exportFile, localFile)
		}

		return nil
	},
}

var sysvarDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete system variables from the variable pool",
	Long: `
DESCRIPTION
-----------
Use this command to delete z/OSMF system variables from the
variable pool. Specify variable names with --var (repeatable).
Without --var, the entire variable pool is deleted.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		plexName, _ := cmd.Flags().GetString("plex-name")
		systemName, _ := cmd.Flags().GetString("system-name")
		varNames, _ := cmd.Flags().GetStringArray("var")

		client := Profile.NewZosmfClient()
		path := fmt.Sprintf("/variables/rest/1.0/systems/%s.%s", plexName, systemName)

		var payload interface{}
		if len(varNames) > 0 {
			payload = varNames
		}

		resp, err := client.DeleteWithBody(path, payload, nil)
		if err != nil {
			return err
		}
		if apiErr := zosmf.CheckResponse(resp, 204); apiErr != nil {
			fmt.Fprintln(os.Stderr, apiErr)
			os.Exit(8)
		}
		fmt.Println("System variables deleted successfully.")
		return nil
	},
}

func init() {
	sysvarGetCmd.Flags().StringP("plex-name", "x", "", "The name of a z/OS sysplex.")
	sysvarGetCmd.MarkFlagRequired("plex-name")
	sysvarGetCmd.Flags().StringP("system-name", "y", "", "The name of a z/OS system.")
	sysvarGetCmd.MarkFlagRequired("system-name")
	sysvarGetCmd.Flags().StringArrayP("var-name", "v", nil, "Variable or symbol name (repeatable).")
	sysvarGetCmd.Flags().StringP("source", "o", "", "Source type: variable (default) or symbol.")
	sysvarGetCmd.Flags().BoolP("tui", "t", false, "Activate terminal user interface")

	sysvarCreateCmd.Flags().StringP("plex-name", "x", "", "The name of a z/OS sysplex.")
	sysvarCreateCmd.MarkFlagRequired("plex-name")
	sysvarCreateCmd.Flags().StringP("system-name", "y", "", "The name of a z/OS system.")
	sysvarCreateCmd.MarkFlagRequired("system-name")
	sysvarCreateCmd.Flags().StringArrayP("var", "v", nil, "Variable in name=value format (repeatable).")
	sysvarCreateCmd.Flags().StringArrayP("desc", "e", nil, "Description for each variable (same order as --var).")

	sysvarImportCmd.Flags().StringP("plex-name", "x", "", "The name of a z/OS sysplex.")
	sysvarImportCmd.MarkFlagRequired("plex-name")
	sysvarImportCmd.Flags().StringP("system-name", "y", "", "The name of a z/OS system.")
	sysvarImportCmd.MarkFlagRequired("system-name")
	sysvarImportCmd.Flags().StringP("file", "f", "", "Path to the CSV file on z/OS containing the variables to import.")
	sysvarImportCmd.MarkFlagRequired("file")
	sysvarImportCmd.Flags().StringP("local", "l", "", "Local file to upload to z/OS before importing.")

	sysvarExportCmd.Flags().StringP("plex-name", "x", "", "The name of a z/OS sysplex.")
	sysvarExportCmd.MarkFlagRequired("plex-name")
	sysvarExportCmd.Flags().StringP("system-name", "y", "", "The name of a z/OS system.")
	sysvarExportCmd.MarkFlagRequired("system-name")
	sysvarExportCmd.Flags().StringP("file", "f", "", "Path to the export CSV file on z/OS.")
	sysvarExportCmd.MarkFlagRequired("file")
	sysvarExportCmd.Flags().BoolP("overwrite", "w", false, "Overwrite the file if it already exists.")
	sysvarExportCmd.Flags().StringP("local", "l", "", "Local file path to download the exported CSV to.")

	sysvarDeleteCmd.Flags().StringP("plex-name", "x", "", "The name of a z/OS sysplex.")
	sysvarDeleteCmd.MarkFlagRequired("plex-name")
	sysvarDeleteCmd.Flags().StringP("system-name", "y", "", "The name of a z/OS system.")
	sysvarDeleteCmd.MarkFlagRequired("system-name")
	sysvarDeleteCmd.Flags().StringArrayP("var", "v", nil, "Variable name to delete (repeatable). Omit to delete the entire pool.")

	sysvarCmd.AddCommand(sysvarGetCmd)
	sysvarCmd.AddCommand(sysvarCreateCmd)
	sysvarCmd.AddCommand(sysvarImportCmd)
	sysvarCmd.AddCommand(sysvarExportCmd)
	sysvarCmd.AddCommand(sysvarDeleteCmd)
	rootCmd.AddCommand(sysvarCmd)
}
