// Package tui provides an interactive terminal UI over the task service,
// built with Bubble Tea and sharing the board's rendering (internal/render).
//
// The model is a small state machine: normal mode for navigation and status
// changes; modal input modes (add, edit, tag filter) backed by a Bubble Tea
// text input; and a y/n confirm mode for deletions.
package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/render"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	helpStyle  = lipgloss.NewStyle().Faint(true)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.AdaptiveColor{Light: "252", Dark: "238"}).
				Foreground(lipgloss.AdaptiveColor{Light: "16", Dark: "231"})
)

const defaultRowWidth = 60

type mode int

const (
	modeNormal mode = iota
	modeAdd
	modeEdit
	modeFilterTag
	modeConfirmDelete
)

type model struct {
	svc *service.TaskService
	ctx context.Context

	tasks   []task.Task
	blocked map[int]bool
	cursor  int

	width   int
	grouped bool

	statusFilter *task.Status
	tagFilter    string

	mode      mode
	input     textinput.Model
	pendingID int

	err      error
	quitting bool
}

func newModel(svc *service.TaskService) model {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.CharLimit = 200
	m := model{svc: svc, ctx: context.Background(), input: ti}
	m.reload()
	return m
}

func (m *model) reload() {
	filter := service.ListFilter{Sort: service.SortByUrgency, Tag: m.tagFilter}
	if m.statusFilter != nil {
		filter.Status = m.statusFilter
	}
	tasks, err := m.svc.List(m.ctx, filter)
	if err != nil {
		m.err = err
		return
	}
	blocked, err := m.svc.BlockedSet(m.ctx)
	if err != nil {
		m.err = err
		return
	}
	if m.grouped {
		tasks = reorderByFirstTag(tasks)
	}
	m.tasks = tasks
	m.blocked = blocked
	if m.cursor >= len(tasks) {
		m.cursor = max(0, len(tasks)-1)
	}
	m.err = nil
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.input.Width = max(20, msg.Width-10)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		switch m.mode {
		case modeAdd, modeEdit, modeFilterTag:
			return m.handleInputKey(msg)
		case modeConfirmDelete:
			return m.handleConfirmKey(msg), nil
		default:
			return m.handleNormalKey(msg)
		}
	}
	return m, nil
}

// normalHandlers maps a key string (per tea.KeyMsg.String()) to the mutation
// it triggers on the model in normal mode. Centralising the bindings keeps
// the dispatch function simple and the table self-documenting.
var normalHandlers = map[string]func(*model){
	"j":     func(m *model) { m.moveCursor(+1) },
	"down":  func(m *model) { m.moveCursor(+1) },
	"k":     func(m *model) { m.moveCursor(-1) },
	"up":    func(m *model) { m.moveCursor(-1) },
	" ":     func(m *model) { m.setStatus(nextStatus(m.currentStatus())) },
	"enter": func(m *model) { m.setStatus(nextStatus(m.currentStatus())) },
	"d":     func(m *model) { m.setStatus(task.StatusDone) },
	"r":     (*model).reload,
	"a":     (*model).enterAdd,
	"e":     (*model).enterEdit,
	"x":     (*model).enterConfirmDelete,
	"g":     (*model).toggleGrouped,
	"s":     (*model).cycleStatusAndReload,
	"f":     (*model).enterFilterTag,
	"c":     (*model).clearFilters,
}

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if k == "q" || k == "esc" {
		m.quitting = true
		return m, tea.Quit
	}
	if h, ok := normalHandlers[k]; ok {
		h(&m)
	}
	return m, nil
}

func (m *model) moveCursor(delta int) {
	next := m.cursor + delta
	if next < 0 || next >= len(m.tasks) {
		return
	}
	m.cursor = next
}

func (m *model) toggleGrouped() {
	m.grouped = !m.grouped
	m.reload()
}

func (m *model) cycleStatusAndReload() {
	m.cycleStatusFilter()
	m.reload()
}

func (m *model) clearFilters() {
	m.statusFilter = nil
	m.tagFilter = ""
	m.reload()
}

func (m model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.exitInput()
		return m, nil
	case tea.KeyEnter:
		m.commitInput()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) handleConfirmKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "y":
		if err := m.svc.Delete(m.ctx, m.pendingID); err != nil {
			m.err = err
		}
		m.pendingID = 0
		m.mode = modeNormal
		m.reload()
	case "n", "esc":
		m.pendingID = 0
		m.mode = modeNormal
	}
	return m
}

func (m *model) enterAdd() {
	m.input.SetValue("")
	m.input.Placeholder = "new task title"
	m.input.Focus()
	m.mode = modeAdd
}

func (m *model) enterEdit() {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	m.input.SetValue(t.Title)
	m.input.Placeholder = ""
	m.input.Focus()
	m.pendingID = t.ID
	m.mode = modeEdit
}

func (m *model) enterFilterTag() {
	m.input.SetValue(m.tagFilter)
	m.input.Placeholder = "tag (empty clears)"
	m.input.Focus()
	m.mode = modeFilterTag
}

func (m *model) enterConfirmDelete() {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	m.pendingID = t.ID
	m.mode = modeConfirmDelete
}

func (m *model) exitInput() {
	m.input.Blur()
	m.input.SetValue("")
	m.input.Placeholder = ""
	m.pendingID = 0
	m.mode = modeNormal
}

func (m *model) commitInput() {
	val := strings.TrimSpace(m.input.Value())
	switch m.mode {
	case modeAdd:
		if val != "" {
			t, err := m.svc.Add(m.ctx, val)
			if err != nil {
				m.err = err
			} else {
				m.reload()
				m.focusTask(t.ID)
			}
		}
	case modeEdit:
		if val != "" && m.pendingID != 0 {
			if _, err := m.svc.Rename(m.ctx, m.pendingID, val); err != nil {
				m.err = err
			}
			m.reload()
		}
	case modeFilterTag:
		m.tagFilter = val
		m.cursor = 0
		m.reload()
	}
	m.exitInput()
}

func (m *model) focusTask(id int) {
	for i, t := range m.tasks {
		if t.ID == id {
			m.cursor = i
			return
		}
	}
}

func (m *model) cycleStatusFilter() {
	sequence := []string{"", string(task.StatusTodo), string(task.StatusInProgress), string(task.StatusBlocked), string(task.StatusDone)}
	cur := ""
	if m.statusFilter != nil {
		cur = string(*m.statusFilter)
	}
	for i, s := range sequence {
		if s == cur {
			next := sequence[(i+1)%len(sequence)]
			if next == "" {
				m.statusFilter = nil
			} else {
				st := task.Status(next)
				m.statusFilter = &st
			}
			m.cursor = 0
			return
		}
	}
	m.statusFilter = nil
}

func (m *model) currentTask() (task.Task, bool) {
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return task.Task{}, false
	}
	return m.tasks[m.cursor], true
}

func (m *model) currentStatus() task.Status {
	if t, ok := m.currentTask(); ok {
		return t.Status
	}
	return task.StatusTodo
}

func (m *model) setStatus(st task.Status) {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	if _, err := m.svc.ChangeStatus(m.ctx, t.ID, st); err != nil {
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

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(" komlist") + m.headerSuffix() + "\n\n")
	if m.err != nil {
		b.WriteString("  error: " + m.err.Error() + "\n\n")
	}
	if len(m.tasks) == 0 {
		b.WriteString("  (no tasks)\n")
	} else {
		b.WriteString(m.renderTasks())
	}
	if sb := m.statusBar(); sb != "" {
		b.WriteString("\n" + sb)
	}
	b.WriteString("\n" + m.bottomBar() + "\n")
	return b.String()
}

func (m model) headerSuffix() string {
	var parts []string
	if m.statusFilter != nil {
		parts = append(parts, "· "+string(*m.statusFilter))
	}
	if m.tagFilter != "" {
		parts = append(parts, "· #"+m.tagFilter)
	}
	if len(parts) == 0 {
		return ""
	}
	return helpStyle.Render(" " + strings.Join(parts, " "))
}

func (m model) renderTasks() string {
	var b strings.Builder
	prev := ""
	for i, t := range m.tasks {
		if m.grouped {
			bucket := render.UntaggedGroup
			if len(t.Tags) > 0 {
				bucket = t.Tags[0]
			}
			if bucket != prev {
				if prev != "" {
					b.WriteString("\n")
				}
				b.WriteString(" " + render.GroupStyle.Render(bucket) + "\n")
				prev = bucket
			}
		}
		if i == m.cursor {
			plain := "› " + render.TaskLinePlain(t, m.blocked[t.ID])
			b.WriteString(selectedRowStyle.Width(m.rowWidth()).Render(plain) + "\n")
		} else {
			b.WriteString("  " + render.TaskLine(t, m.blocked[t.ID]) + "\n")
		}
	}
	return b.String()
}

func (m model) statusBar() string {
	var parts []string
	if m.statusFilter != nil {
		parts = append(parts, "status="+string(*m.statusFilter))
	}
	if m.tagFilter != "" {
		parts = append(parts, "tag="+m.tagFilter)
	}
	if m.grouped {
		parts = append(parts, "grouped")
	}
	if len(parts) == 0 {
		return ""
	}
	return helpStyle.Render(" filter: " + strings.Join(parts, " · "))
}

func (m model) bottomBar() string {
	switch m.mode {
	case modeAdd:
		return helpStyle.Render(" add: ") + m.input.View()
	case modeEdit:
		return helpStyle.Render(" edit: ") + m.input.View()
	case modeFilterTag:
		return helpStyle.Render(" tag filter: ") + m.input.View()
	case modeConfirmDelete:
		return helpStyle.Render(fmt.Sprintf(" Delete #%d? (y/n)", m.pendingID))
	default:
		return helpStyle.Render(" ↑/↓ move · space cycle · d done · a add · e edit · x delete · g group · s status · f tag · c clear · r reload · q quit")
	}
}

func (m model) rowWidth() int {
	if m.width > 0 {
		return m.width
	}
	return defaultRowWidth
}

func reorderByFirstTag(tasks []task.Task) []task.Task {
	buckets := map[string][]task.Task{}
	var names []string
	for _, t := range tasks {
		name := render.UntaggedGroup
		if len(t.Tags) > 0 {
			name = t.Tags[0]
		}
		if _, ok := buckets[name]; !ok {
			names = append(names, name)
		}
		buckets[name] = append(buckets[name], t)
	}
	sort.Slice(names, func(i, j int) bool {
		if names[i] == render.UntaggedGroup {
			return false
		}
		if names[j] == render.UntaggedGroup {
			return true
		}
		return names[i] < names[j]
	})
	out := make([]task.Task, 0, len(tasks))
	for _, n := range names {
		out = append(out, buckets[n]...)
	}
	return out
}

// Run starts the interactive program and blocks until the user quits.
func Run(svc *service.TaskService) error {
	_, err := tea.NewProgram(newModel(svc)).Run()
	return err
}
