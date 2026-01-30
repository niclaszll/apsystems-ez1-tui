package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/niclaszll/apsystems-ez1-tui/pkg/apsystems"
)

type View int

const (
	ViewDashboard View = iota
	ViewDeviceInfo
	ViewAlarms
	ViewPowerControl
)

type keyMap struct {
	Help        key.Binding
	Quit        key.Binding
	Refresh     key.Binding
	Tab         key.Binding
	PowerOn     key.Binding
	PowerOff    key.Binding
	PowerSleep  key.Binding
	IncreasePwr key.Binding
	DecreasePwr key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Tab, k.Refresh}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Refresh, k.Help, k.Quit},
		{k.PowerOn, k.PowerOff, k.PowerSleep},
		{k.IncreasePwr, k.DecreasePwr},
	}
}

var keys = keyMap{
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
	PowerOn: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "power on"),
	),
	PowerOff: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "power off"),
	),
	PowerSleep: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sleep mode"),
	),
	IncreasePwr: key.NewBinding(
		key.WithKeys("+", "="),
		key.WithHelp("+", "increase power"),
	),
	DecreasePwr: key.NewBinding(
		key.WithKeys("-", "_"),
		key.WithHelp("-", "decrease power"),
	),
}

type Model struct {
	client      *apsystems.Client
	currentView View
	spinner     spinner.Model
	help        help.Model
	keys        keyMap
	loading     bool
	err         error
	stats       *apsystems.Statistics
	deviceInfo  *apsystems.DeviceInfo
	alarmInfo   *apsystems.AlarmInfo
	powerStatus *apsystems.PowerStatus
	powerLimit  *apsystems.PowerLimit
	width       int
	height      int
	showHelp    bool
}

type tickMsg time.Time
type statsMsg *apsystems.Statistics
type deviceInfoMsg *apsystems.DeviceInfo
type alarmInfoMsg *apsystems.AlarmInfo
type powerStatusMsg *apsystems.PowerStatus
type powerLimitMsg *apsystems.PowerLimit
type errMsg error

func NewModel(client *apsystems.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		client:      client,
		currentView: ViewDashboard,
		spinner:     s,
		help:        help.New(),
		keys:        keys,
		loading:     true,
		showHelp:    false,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		fetchStats(m.client),
		fetchDeviceInfo(m.client),
		fetchPowerStatus(m.client),
		fetchPowerLimit(m.client),
		fetchAlarmInfo(m.client),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchStats(client *apsystems.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		stats, err := client.GetStatistics(ctx)
		if err != nil {
			return errMsg(err)
		}
		return statsMsg(stats)
	}
}

func fetchDeviceInfo(client *apsystems.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		info, err := client.GetDeviceInfo(ctx)
		if err != nil {
			return errMsg(err)
		}
		return deviceInfoMsg(info)
	}
}

func fetchAlarmInfo(client *apsystems.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		info, err := client.GetAlarmInfo(ctx)
		if err != nil {
			return errMsg(err)
		}
		return alarmInfoMsg(info)
	}
}

func fetchPowerStatus(client *apsystems.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		status, err := client.GetDevicePowerStatus(ctx)
		if err != nil {
			return errMsg(err)
		}
		return powerStatusMsg(status)
	}
}

func fetchPowerLimit(client *apsystems.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		limit, err := client.GetMaxPower(ctx)
		if err != nil {
			return errMsg(err)
		}
		return powerLimitMsg(limit)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keys.Tab):
			m.currentView = (m.currentView + 1) % 4
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			return m, tea.Batch(
				fetchStats(m.client),
				fetchDeviceInfo(m.client),
				fetchPowerStatus(m.client),
				fetchPowerLimit(m.client),
				fetchAlarmInfo(m.client),
			)
		case key.Matches(msg, m.keys.PowerOn):
			if m.currentView == ViewPowerControl {
				return m, m.setPowerStatus("ON")
			}
		case key.Matches(msg, m.keys.PowerOff):
			if m.currentView == ViewPowerControl {
				return m, m.setPowerStatus("OFF")
			}
		case key.Matches(msg, m.keys.PowerSleep):
			if m.currentView == ViewPowerControl {
				return m, m.setPowerStatus("SLEEP")
			}
		case key.Matches(msg, m.keys.IncreasePwr):
			if m.currentView == ViewPowerControl && m.powerLimit != nil {
				newPower := int(m.powerLimit.Data.MaxPower) + 50
				if newPower <= 800 {
					return m, m.setMaxPower(newPower)
				}
			}
		case key.Matches(msg, m.keys.DecreasePwr):
			if m.currentView == ViewPowerControl && m.powerLimit != nil {
				newPower := int(m.powerLimit.Data.MaxPower) - 50
				if newPower >= 30 {
					return m, m.setMaxPower(newPower)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tickMsg:
		return m, fetchStats(m.client)

	case statsMsg:
		m.stats = msg
		m.loading = false
		m.err = nil
		return m, tickCmd()

	case deviceInfoMsg:
		m.deviceInfo = msg
		return m, nil

	case alarmInfoMsg:
		m.alarmInfo = msg
		return m, nil

	case powerStatusMsg:
		m.powerStatus = msg
		return m, nil

	case powerLimitMsg:
		m.powerLimit = msg
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, tickCmd()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) setPowerStatus(status string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := m.client.SetDevicePowerStatus(ctx, status)
		if err != nil {
			return errMsg(err)
		}
		return fetchPowerStatus(m.client)()
	}
}

func (m Model) setMaxPower(watts int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := m.client.SetMaxPower(ctx, watts)
		if err != nil {
			return errMsg(err)
		}
		return fetchPowerLimit(m.client)()
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var content string

	switch m.currentView {
	case ViewDashboard:
		content = m.renderDashboard()
	case ViewDeviceInfo:
		content = m.renderDeviceInfo()
	case ViewAlarms:
		content = m.renderAlarms()
	case ViewPowerControl:
		content = m.renderPowerControl()
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	// Apply some padding to align with header
	contentStyle := lipgloss.NewStyle().Padding(0, 1)
	content = contentStyle.Render(content)
	footer = contentStyle.Render(footer)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m Model) renderHeader() string {
	tabs := []string{"Dashboard", "Device Info", "Alarms", "Power Control"}
	var renderedTabs []string

	for i, tab := range tabs {
		style := lipgloss.NewStyle().Bold(true).Padding(0, 1).Margin(0, 1, 0, 0)
		if View(i) == m.currentView {
			style = style.
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4"))
		} else {
			style = style.Foreground(lipgloss.Color("#666666"))
		}
		renderedTabs = append(renderedTabs, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m Model) renderFooter() string {
	if m.showHelp {
		return "\n" + m.help.FullHelpView(m.keys.FullHelp())
	}
	return "\n" + m.help.ShortHelpView(m.keys.ShortHelp())
}

func (m Model) renderDashboard() string {
	if m.loading && m.stats == nil {
		return fmt.Sprintf("\n%s Loading...", m.spinner.View())
	}

	if m.stats == nil {
		if m.err != nil {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)
			return errorStyle.Render(fmt.Sprintf("\nError: %v\n\nRetrying...", m.err))
		}
		return "\nNo data available"
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Width(25)

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	powerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	lines := []string{
		"",
		labelStyle.Render("Current Power Output:") + powerStyle.Render(fmt.Sprintf("%d W", m.stats.TotalPower)),
		labelStyle.Render("Energy Today:") + valueStyle.Render(fmt.Sprintf("%.3f kWh", m.stats.TotalEnergyToday)),
		labelStyle.Render("Lifetime Energy:") + valueStyle.Render(fmt.Sprintf("%.3f kWh", m.stats.TotalEnergyLifetime)),
		"",
		labelStyle.Render("Last Update:") + valueStyle.Render(m.stats.LastUpdate.Format("15:04:05")),
	}

	if m.powerStatus != nil {
		var statusText string
		switch int(m.powerStatus.Data.Status) {
		case 0:
			statusText = "ON"
		case 1:
			statusText = "OFF"
		case 2:
			statusText = "SLEEP"
		default:
			statusText = "UNKNOWN"
		}
		lines = append(lines, labelStyle.Render("Power Status:")+valueStyle.Render(statusText))
	}

	if m.powerLimit != nil {
		lines = append(lines, labelStyle.Render("Max Power Limit:")+valueStyle.Render(fmt.Sprintf("%d W", int(m.powerLimit.Data.MaxPower))))
	}

	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6600")).
			Italic(true)
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(fmt.Sprintf("⚠ Last refresh failed: %v", m.err)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderDeviceInfo() string {
	if m.deviceInfo == nil {
		return "\nLoading device information..."
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	lines := []string{
		"",
		labelStyle.Render("Device ID:") + valueStyle.Render(m.deviceInfo.Data.DeviceID),
		labelStyle.Render("Firmware:") + valueStyle.Render(m.deviceInfo.Data.Firmware),
		"",
		labelStyle.Render("IP Address:") + valueStyle.Render(m.deviceInfo.Data.IPAddr),
		labelStyle.Render("SSID:") + valueStyle.Render(m.deviceInfo.Data.SSIDName),
		"",
		labelStyle.Render("Min Power:") + valueStyle.Render(fmt.Sprintf("%d W", int(m.deviceInfo.Data.MinPower))),
		labelStyle.Render("Max Power:") + valueStyle.Render(fmt.Sprintf("%d W", int(m.deviceInfo.Data.MaxPower))),
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderAlarms() string {
	if m.alarmInfo == nil {
		return "\nLoading alarm information..."
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Width(25)

	okStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	alarmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	renderStatus := func(value int) string {
		if value == 0 {
			return okStyle.Render("OK")
		}
		return alarmStyle.Render("ALARM")
	}

	lines := []string{
		"",
		labelStyle.Render("Grid Fault:") + renderStatus(int(m.alarmInfo.Data.Og)),
		labelStyle.Render("PV1 Short Circuit:") + renderStatus(int(m.alarmInfo.Data.Isce1)),
		labelStyle.Render("PV2 Short Circuit:") + renderStatus(int(m.alarmInfo.Data.Isce2)),
		labelStyle.Render("Output Error:") + renderStatus(int(m.alarmInfo.Data.Oe)),
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderPowerControl() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true)

	lines := []string{""}

	if m.powerStatus != nil {
		var statusText string
		switch int(m.powerStatus.Data.Status) {
		case 0:
			statusText = "ON (Normal)"
		case 1:
			statusText = "OFF"
		case 2:
			statusText = "SLEEP"
		default:
			statusText = "UNKNOWN"
		}
		lines = append(lines, labelStyle.Render("Current Status:")+valueStyle.Render(statusText))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Press 'o' for ON, 'f' for OFF, 's' for SLEEP"))
		lines = append(lines, "")
	}

	if m.powerLimit != nil {
		lines = append(lines, labelStyle.Render("Max Power Limit:")+valueStyle.Render(fmt.Sprintf("%d W", int(m.powerLimit.Data.MaxPower))))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Press '+' to increase by 50W, '-' to decrease by 50W"))
		lines = append(lines, helpStyle.Render("Range: 30-800 W"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
