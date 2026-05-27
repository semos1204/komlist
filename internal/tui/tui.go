// Package tui provides an interactive terminal UI over the task service,
// built with Bubble Tea and sharing the board's rendering (internal/render).
//
// V1 is read + status changes: navigate with j/k, cycle status with space,
// mark done with d, reload with r, quit with q. Editing titles or adding
// tasks is done from the regular CLI.
package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/render"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	helpStyle  = lipgloss.NewStyle().Faint(true)
)

type model struct {
	svc      *service.TaskService
	ctx      context.Context
	tasks    []task.Task
	blocked  map[int]bool
	cursor   int
	err      error
	quitting bool
}

func newModel(svc *service.TaskService) model {
	m := model{svc: svc, ctx: context.Background()}
	m.reload()
	return m
}

func (m *model) reload() {
	tasks, err := m.svc.List(m.ctx, service.ListFilter{Sort: service.SortByUrgency})
	if err != nil {
		m.err = err
		return
	}
	blocked, err := m.svc.BlockedSet(m.ctx)
	if err != nil {
		m.err = err
		return
	}
	m.tasks = tasks
	m.blocked = blocked
	if m.cursor >= len(tasks) {
		m.cursor = max(0, len(tasks)-1)
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case " ", "enter":
		m.setStatus(nextStatus(m.currentStatus()))
	case "d":
		m.setStatus(task.StatusDone)
	case "r":
		m.reload()
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(" komlist") + "\n\n")
	if m.err != nil {
		b.WriteString("  error: " + m.err.Error() + "\n\n")
	}
	if len(m.tasks) == 0 {
		b.WriteString("  (no tasks)\n")
	}
	for i, t := range m.tasks {
		prefix := "  "
		if i == m.cursor {
			prefix = render.SelectedStyle.Render("> ")
		}
		b.WriteString(prefix + render.TaskLine(t, m.blocked[t.ID]) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render(" j/k move · space cycle · d done · r reload · q quit") + "\n")
	return b.String()
}

func (m *model) currentStatus() task.Status {
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return task.StatusTodo
	}
	return m.tasks[m.cursor].Status
}

func (m *model) setStatus(st task.Status) {
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return
	}
	id := m.tasks[m.cursor].ID
	if _, err := m.svc.ChangeStatus(m.ctx, id, st); err != nil {
		m.err = err
		return
	}
	m.reload()
}

func nextStatus(s task.Status) task.Status {
	switch s {
	case task.StatusTodo:
		return task.StatusInProgress
	case task.StatusInProgress:
		return task.StatusDone
	default:
		return task.StatusTodo
	}
}

// Run starts the interactive program and blocks until the user quits.
func Run(svc *service.TaskService) error {
	_, err := tea.NewProgram(newModel(svc)).Run()
	return err
}
