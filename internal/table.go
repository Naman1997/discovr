package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Base style for table borders
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#975b85ff"))

// -------------------- Table Model --------------------
type tableModel struct {
	table table.Model
	data  interface{}
	width int
}

// Create a new table model from any struct slice
func NewTableModel(data interface{}, width int) *tableModel {
	columns, rows := BuildDynamicTableWithWrap(data, width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+3),
		table.WithFocused(true),
	)

	// Styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6e6f6eff")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		BorderForeground(lipgloss.Color("#6e6f6eff")).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#49306d")).
		Bold(false)
	t.SetStyles(s)

	return &tableModel{table: t, data: data, width: width}
}

// -------------------- Bubble Tea Methods --------------------
func (m *tableModel) Init() tea.Cmd {
	return nil
}

func (m *tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		columns, rows := BuildDynamicTableWithWrap(m.data, m.width)
		m.table.SetColumns(columns)
		m.table.SetRows(rows)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *tableModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

// -------------------- Text Wrapping --------------------
func WrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	for len(text) > 0 {
		if len(text) <= maxWidth {
			lines = append(lines, text)
			break
		}
		lines = append(lines, text[:maxWidth])
		text = text[maxWidth:]
	}
	return lines
}

// -------------------- Dynamic Width Compute --------------------
func ComputeColumnWidths(data interface{}, maxTotalWidth int) []int {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	elemType := v.Index(0).Type()
	numCols := elemType.NumField()
	widths := make([]int, numCols)

	// First, take header widths
	for i := 0; i < numCols; i++ {
		widths[i] = len(elemType.Field(i).Name)
	}

	// Then, scan data to find max content width for each column
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		for j := 0; j < numCols; j++ {
			fieldVal := fmt.Sprintf("%v", elem.Field(j).Interface())
			lines := strings.Split(fieldVal, "\n")
			for _, line := range lines {
				if len(line) > widths[j] {
					widths[j] = len(line)
				}
			}
		}
	}

	// Compute total natural width
	totalWidth := numCols - 1 // for column separators
	for _, w := range widths {
		totalWidth += w
	}

	// If too wide, split equally among columns
	if totalWidth > maxTotalWidth {
		equal := (maxTotalWidth - (numCols - 1)) / numCols
		if equal < 3 {
			equal = 3 // enforce minimum
		}
		for i := range widths {
			widths[i] = equal
		}
	}
	return widths
}

// -------------------- Dynamic Table Builder --------------------
func BuildDynamicTableWithWrap(data interface{}, maxWidth int) ([]table.Column, []table.Row) {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil, nil
	}

	elemType := v.Index(0).Type()
	numCols := elemType.NumField()
	colWidths := ComputeColumnWidths(data, maxWidth)

	// Columns
	columns := []table.Column{}
	for i := 0; i < numCols; i++ {
		columns = append(columns, table.Column{
			Title: elemType.Field(i).Name,
			Width: colWidths[i],
		})
	}

	// Rows with wrapping
	rows := []table.Row{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)

		// Wrap each column
		wrappedCols := make([][]string, numCols)
		maxLines := 0
		for j := 0; j < numCols; j++ {
			fieldVal := fmt.Sprintf("%v", elem.Field(j).Interface())
			wrappedCols[j] = WrapText(fieldVal, colWidths[j])
			if len(wrappedCols[j]) > maxLines {
				maxLines = len(wrappedCols[j])
			}
		}

		for line := 0; line < maxLines; line++ {
			row := table.Row{}
			for j := 0; j < numCols; j++ {
				if line < len(wrappedCols[j]) {
					row = append(row, wrappedCols[j][line])
				} else {
					row = append(row, "")
				}
			}
			rows = append(rows, row)
		}
	}
	return columns, rows
}

// -------------------- ShowResults Functions --------------------
func ShowResults[T any](data []T) {
	m := NewTableModel(data, 100)
	fmt.Println(m.View())
}
