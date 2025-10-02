package internal

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Asset struct {
	Name string
	IP   string
	MAC  string
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#975b85ff"))

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
			//		case "enter":
			//			return m, tea.Printf("Selected asset: %s", m.table.SelectedRow()[0])
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func RenderAssetsTable(columns []table.Column, rows []table.Row, height int) {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
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

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func RenderAssetsTableOnce(columns []table.Column, rows []table.Row, height int) {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(height),
	)

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

	fmt.Println(baseStyle.Render(t.View()))
}

func ShowPassiveScanResults() {
	columns := []table.Column{
		{Title: "Src IP", Width: 15},
		{Title: "Protocol", Width: 10},
		{Title: "Src MAC", Width: 20},
		{Title: "Dst MAC", Width: 20},
		{Title: "EthType", Width: 12},
	}

	var rows []table.Row
	for _, r := range passive_results {
		rows = append(rows, table.Row{
			r.SrcIP,
			r.Protocol,
			r.SrcMAC,
			r.DstMAC,
			r.EthernetType,
		})
	}
	RenderAssetsTable(columns, rows, 10)
}

func ShowAzureResultsTable() {
	columns := []table.Column{
		{Title: "Field", Width: 20},
		{Title: "Value", Width: 100},
	}

	var rows []table.Row
	for _, r := range azure_results {
		rows = append(rows, table.Row{"Name         │", r.Name})
		rows = append(rows, table.Row{"UniqueID     │", r.UniqueID})
		rows = append(rows, table.Row{"Location     │", r.Location})
		rows = append(rows, table.Row{"ResourceGroup│", r.ResourceGroup})
		rows = append(rows, table.Row{"NIC          │", r.NIC})
		rows = append(rows, table.Row{"MAC          │", r.MAC})
		rows = append(rows, table.Row{"Subnet       │", r.Subnet})
		rows = append(rows, table.Row{"Vnet         │", r.Vnet})
		rows = append(rows, table.Row{
			lipgloss.NewStyle().Foreground(lipgloss.Color("#7fff00")).Render("IPs"),
			lipgloss.NewStyle().Width(120).Foreground(lipgloss.Color("#7fff00")).Render(fmt.Sprintf("Private: %s\nPublic: %s", r.PrivateIP, r.PublicIP)),
		})
		rows = append(rows, table.Row{"", ""})
	}

	RenderAssetsTable(columns, rows, 20)
}

func ShowActiveResults() {
	columns := []table.Column{
		{Title: "Interfaces", Width: 15},
		{Title: "IP Address  │", Width: 20},
		{Title: "MAC Address", Width: 20},
	}

	var rows []table.Row
	for _, r := range defaultscan_results {
		rows = append(rows, table.Row{
			r.Interface,
			r.Dest_IP,
			r.Dest_Mac,
		})
	}

	RenderAssetsTable(columns, rows, 15)
}

func ShowNmapScanResults() {
	columns := []table.Column{
		{Title: "Port", Width: 8},
		{Title: "Protocol", Width: 10},
		{Title: "State", Width: 12},
		{Title: "Service", Width: 15},
		{Title: "Product", Width: 25},
	}

	var rows []table.Row
	for _, r := range active_results {
		rows = append(rows, table.Row{
			r.Port,
			r.Protocol,
			r.State,
			r.Service,
			r.Product,
		})
	}
	RenderAssetsTableOnce(columns, rows, 15)
}
