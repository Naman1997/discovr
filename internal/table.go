package internal

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Base style for table borders
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#975b85ff"))

// Table Model
type tableModel struct {
	table table.Model
	data  interface{}
	width int
}

// Create a new table model from any struct slice
func NewTableModel(data interface{}, width int) *tableModel {
	columns, rows := BuildTable(data, width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+3),
		table.WithFocused(false),
	)

	// Styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6e6f6eff")).
		BorderBottom(true).
		Align(lipgloss.Right).
		Bold(true)
	s.Selected = s.Selected.
		BorderForeground(lipgloss.Color("")).
		Foreground(lipgloss.Color("")).
		Background(lipgloss.Color("")).
		Bold(false)
	t.SetStyles(s)

	return &tableModel{table: t, data: data, width: width}
}

// Bubble Tea Methods
func (m *tableModel) Init() tea.Cmd {
	return nil
}

func (m *tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		columns, rows := BuildTable(m.data, m.width)
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

// Text Wrapping
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

// Width Calculation
func ComputeColumnWidths(data interface{}, maxTotalWidth int) []int {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	elemType := v.Index(0).Type()
	numCols := elemType.NumField()
	widths := make([]int, numCols)

	// Small value for separator/padding between columns
	const divider = 1

	// Compute equal widths
	baseWidth := (maxTotalWidth / numCols) - divider
	if baseWidth < 1 { // enforce minimum width
		baseWidth = 1
	}

	for i := 0; i < numCols; i++ {
		widths[i] = baseWidth
	}

	return widths
}

// Dynamic Table Builder
func BuildTable(data interface{}, maxWidth int) ([]table.Column, []table.Row) {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil, nil
	}

	elemType := v.Index(0).Type()
	numCols := elemType.NumField()
	colWidths := ComputeColumnWidths(data, maxWidth)

	centerStyle := lipgloss.NewStyle().Align(lipgloss.Center)

	// Headers
	columns := []table.Column{}
	for i := 0; i < numCols; i++ {
		title := elemType.Field(i).Name
		centeredTitle := centerStyle.Width(colWidths[i]).Render(title)

		columns = append(columns, table.Column{
			Title: centeredTitle,
			Width: colWidths[i],
		})
	}

	// Rows
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
				var cell string
				if line < len(wrappedCols[j]) {
					cell = centerStyle.Width(colWidths[j]).Render(wrappedCols[j][line])
				} else {
					cell = centerStyle.Width(colWidths[j]).Render("")
				}
				row = append(row, cell)
			}
			rows = append(rows, row)
		}
	}

	return columns, rows
}

// Result Display Function
func ShowResults[T any](data []T) {
	m := NewTableModel(data, 120)
	fmt.Println(m.View())
}
