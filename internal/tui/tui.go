package tui

import (
	"fmt"
	logger "zcli/internal/logging"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

var (
	CustomBorder = table.Border{
		Top:    "─",
		Left:   "│",
		Right:  "│",
		Bottom: "─",

		TopRight:       "╮",
		TopLeft:        "╭",
		BottomRight:    "╯",
		BottomLeft:     "╰",
		TopJunction:    "─",
		LeftJunction:   "├",
		RightJunction:  "┤",
		BottomJunction: "─",
		InnerJunction:  "─",
		InnerDivider:   " ",
	}
	CustomBrowserBorder = table.Border{
		Top:    " ",
		Left:   " ",
		Right:  " ",
		Bottom: " ",

		TopRight:       " ",
		TopLeft:        " ",
		BottomRight:    " ",
		BottomLeft:     " ",
		TopJunction:    " ",
		LeftJunction:   " ",
		RightJunction:  " ",
		BottomJunction: " ",
		InnerJunction:  " ",
		InnerDivider:   " ",
	}
)

// Job Column definitions
const (
	ColumnID              = "ID"
	ColumnName            = "Name"
	ColumnType            = "Type"
	JobColumnJobname      = "Jobname"
	ColumnStatus          = "Status"
	JobColumnJobID        = "JobID"
	JobColumnOwner        = "Owner"
	JobColumnClass        = "Class"
	JobColumnRetcode      = "Retcode"
	JobColumnStarted      = "Started"
	JobColumnEnded        = "Ended"
	JobColumnMember       = "Member"
	JobColumnSystem       = "System"
	FileColumnDDName      = "DD-Name"
	FileColumnStepName    = "Stepname"
	FileColumnProcStep    = "Procstep"
	FileColumnFID         = "DSID"
	FileColumnBC          = "Byte-Cnt"
	FileColumnRC          = "Rec-Cnt"
	FileColumnLrecl       = "Rec-Length"
	BrowserColumnRec      = "File"
	BrowserColumnID       = "Line"
	FileLink              = "Link"
	MfscolumnFSMountPoint = "Mountpoint"
	MfscolumnFSType       = "FS-Type"
	MfscolumnFSStatus     = "Status"
	SubsysColumnActive    = "Active"
	SubsysColumnPrimary   = "Primary"
	SubsysColumnDynamic   = "Dynamic"
	SubsysColumnFuncs     = "Functions"
	SysvarColumnValue     = "Value"
	SysvarColumnDesc      = "Description"
)

// Generate columns based on how many are critical to show some summary
func GenerateJobColumns() []table.Column {

	columns := []table.Column{
		table.NewColumn(ColumnID, ColumnID, 3),
		table.NewColumn(JobColumnJobname, JobColumnJobname, 8).WithFiltered(true),
		table.NewColumn(JobColumnJobID, JobColumnJobID, 8).WithFiltered(true),
		table.NewColumn(JobColumnOwner, JobColumnOwner, 8).WithFiltered(true),
		table.NewColumn(JobColumnClass, JobColumnClass, 8),
		table.NewColumn(ColumnStatus, JobColumnEnded, 8).WithFiltered(true),
		table.NewColumn(JobColumnRetcode, JobColumnRetcode, 12).WithFiltered(true),
		table.NewColumn(ColumnType, ColumnType, 8),
		table.NewColumn(JobColumnStarted, JobColumnStarted, 24),
		table.NewColumn(JobColumnEnded, JobColumnEnded, 24),
		table.NewColumn(JobColumnMember, JobColumnMember, 8).WithFiltered(true),
		table.NewColumn(JobColumnSystem, JobColumnSystem, 8).WithFiltered(true),
	}

	return columns
}

func GenerateFileColumns() []table.Column {

	columns := []table.Column{
		table.NewColumn(ColumnID, ColumnID, 3),
		table.NewColumn(FileColumnDDName, FileColumnDDName, 8),
		table.NewColumn(FileColumnStepName, FileColumnStepName, 8),
		table.NewColumn(FileColumnProcStep, FileColumnProcStep, 8),
		table.NewColumn(FileColumnFID, FileColumnFID, 8),
		table.NewColumn(JobColumnJobID, JobColumnJobID, 8),
		table.NewColumn(JobColumnClass, JobColumnClass, 8),
		table.NewColumn(FileColumnRC, FileColumnRC, 8),
		table.NewColumn(FileColumnBC, FileColumnBC, 8),
		table.NewColumn(FileColumnLrecl, FileColumnLrecl, 8),
	}

	return columns
}

func GenerateFileBrowserColumns(ddName string) []table.Column {
	logger.Log.Debug("Generating Browser Columns")
	columns := []table.Column{
		table.NewColumn(BrowserColumnID, BrowserColumnID, 6),
		//table.NewColumn(tui.BrowserColumnRec, fmt.Sprintf("Browsing file %s", filename), 160),
		table.NewColumn(BrowserColumnRec, fmt.Sprintf("%s<----+----1----+----2----+----3----+----4----+----5----+----6----+----7----+----8----+----9----+---10----+---11----+---12----+---13----+---14----+---15>", ddName), 153),
	}

	return columns
}

func GenerateMfsColumns() []table.Column {
	columns := []table.Column{
		table.NewColumn(ColumnID, ColumnID, 3),
		table.NewColumn(MfscolumnFSMountPoint, MfscolumnFSMountPoint, 44),
		table.NewColumn(ColumnName, ColumnName, 44),
		table.NewColumn(MfscolumnFSType, MfscolumnFSType, 8),
		table.NewColumn(MfscolumnFSStatus, MfscolumnFSStatus, 10),
	}
	return columns
}

func GenerateDetailsMfsColumns() []table.Column {
	columns := []table.Column{
		table.NewColumn(ColumnID, ColumnID, 3),
		table.NewColumn(MfscolumnFSMountPoint, MfscolumnFSMountPoint, 44),
		table.NewColumn(ColumnName, ColumnName, 44),
		table.NewColumn(MfscolumnFSType, MfscolumnFSType, 8),
		table.NewColumn(MfscolumnFSStatus, MfscolumnFSStatus, 10),
	}
	return columns
}

func GenerateSubsysColumns() []table.Column {
	columns := []table.Column{
		table.NewColumn(ColumnID, ColumnID, 3),
		table.NewColumn(ColumnName, "Subsystem", 10).WithFiltered(true),
		table.NewColumn(SubsysColumnActive, SubsysColumnActive, 8),
		table.NewColumn(SubsysColumnPrimary, SubsysColumnPrimary, 8),
		table.NewColumn(SubsysColumnDynamic, SubsysColumnDynamic, 8),
		table.NewColumn(SubsysColumnFuncs, SubsysColumnFuncs, 40),
	}
	return columns
}

func GenerateSysvarColumns() []table.Column {
	columns := []table.Column{
		table.NewColumn(ColumnID, ColumnID, 3),
		table.NewColumn(ColumnName, ColumnName, 30).WithFiltered(true),
		table.NewColumn(SysvarColumnValue, SysvarColumnValue, 30).WithFiltered(true),
		table.NewColumn(SysvarColumnDesc, SysvarColumnDesc, 50),
	}
	return columns
}

// Map keys
func RemapTableKeys() table.KeyMap {
	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down")
	keys.RowUp.SetKeys("k", "up")
	keys.ScrollRight.SetKeys("l", "right")
	keys.ScrollLeft.SetKeys("h", "left")
	keys.PageDown.SetKeys("f8")
	keys.PageUp.SetKeys("f7")
	return keys
}

func TableLayout(c []table.Column, r []table.Row, k table.KeyMap, sel bool, fil bool, foc bool, ps int) table.Model {
	t := table.New(c).
		WithRows(r).
		Border(CustomBorder).
		//HeaderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#fdbf3b")).Bold(true)).
		HeaderStyle(lipgloss.NewStyle().Bold(true)).
		SelectableRows(sel).
		Filtered(fil).
		Focused(foc).
		WithKeyMap(k).
		WithStaticFooter("").
		WithPageSize(ps).
		WithSelectedText(" ", "✓").
		WithBaseStyle(
			lipgloss.NewStyle().
				BorderForeground(lipgloss.Color("#71fff1")).
				Foreground(lipgloss.Color("#92b2ff")).
				Align(lipgloss.Left),
		).
		WithMissingDataIndicatorStyled(table.StyledCell{
			Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#92b2ff")),
			Data:  "",
		})
	return t
}
