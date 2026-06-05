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
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/semos1204/komlist/internal/render"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(render.Accent)
	helpStyle  = lipgloss.NewStyle().Faint(true)

	// selectedRowStyle tints the whole cursor row with a soft blue
	// background keyed to the accent colour, so the row stands out at
	// a glance.
	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "153", Dark: "25"})
)

const (
	// defaultRowWidth is the selection-bar width used before the first
	// WindowSizeMsg arrives. Replaced once the real terminal width is known.
	defaultRowWidth = 80
	// inputRightMargin keeps the textinput field a few columns away from the
	// right edge so the caret never touches the border.
	inputRightMargin = 10
)

type mode int

const (
	modeNormal mode = iota
	modeAdd
	modeEdit
	modeEditTags
	modeEditDue
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

	mode     mode
	input    textinput.Model
	targetID int // task being edited or queued for deletion (0 when none)

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
		m.input.Width = max(20, msg.Width-inputRightMargin)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		switch m.mode {
		case modeAdd, modeEdit, modeEditTags, modeEditDue, modeFilterTag:
			return m.handleInputKey(msg)
		case modeConfirmDelete:
			return m.handleConfirmKey(msg)
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
	"p":     (*model).cyclePriority,
	"t":     (*model).enterEditTags,
	"u":     (*model).enterEditDue,
	"R":     (*model).cycleRecurrence,
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

func (m model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if err := m.svc.Delete(m.ctx, m.targetID); err != nil {
			m.err = err
		}
		m.targetID = 0
		m.mode = modeNormal
		m.reload()
	case "n", "esc":
		m.targetID = 0
		m.mode = modeNormal
	}
	return m, nil
}

// enterInput prepares the textinput field for one of the input modes. The
// targetID identifies the task being acted on (0 when irrelevant, e.g. add
// or filter).
func (m *model) enterInput(md mode, value, placeholder string, targetID int) {
	m.input.SetValue(value)
	m.input.Placeholder = placeholder
	m.input.Focus()
	m.targetID = targetID
	m.mode = md
}

func (m *model) enterAdd() {
	m.enterInput(modeAdd, "", "new task title", 0)
}

func (m *model) enterEdit() {
	if t, ok := m.currentTask(); ok {
		m.enterInput(modeEdit, t.Title, "", t.ID)
	}
}

func (m *model) enterFilterTag() {
	m.enterInput(modeFilterTag, m.tagFilter, "tag (empty clears)", 0)
}

func (m *model) enterEditTags() {
	if t, ok := m.currentTask(); ok {
		m.enterInput(modeEditTags, strings.Join(t.Tags, ","), "tag1,tag2 (empty clears)", t.ID)
	}
}

func (m *model) enterEditDue() {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	value := ""
	if t.DueAt != nil {
		value = t.DueAt.Format(time.DateOnly)
	}
	m.enterInput(modeEditDue, value, "YYYY-MM-DD (empty clears)", t.ID)
}

func (m *model) enterConfirmDelete() {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	m.targetID = t.ID
	m.mode = modeConfirmDelete
}

func (m *model) exitInput() {
	m.input.Blur()
	m.input.SetValue("")
	m.input.Placeholder = ""
	m.targetID = 0
	m.mode = modeNormal
}

func (m *model) commitInput() {
	val := strings.TrimSpace(m.input.Value())
	switch m.mode {
	case modeAdd:
		m.commitAdd(val)
	case modeEdit:
		m.commitRename(val)
	case modeEditTags:
		m.commitTags(val)
	case modeEditDue:
		m.commitDue(val)
	case modeFilterTag:
		m.tagFilter = val
		m.cursor = 0
		m.reload()
	}
	m.exitInput()
}

func (m *model) commitAdd(val string) {
	if val == "" {
		return
	}
	t, err := m.svc.Add(m.ctx, val)
	if err != nil {
		m.err = err
		return
	}
	m.reload()
	m.focusTask(t.ID)
}

func (m *model) commitRename(val string) {
	if val == "" || m.targetID == 0 {
		return
	}
	if _, err := m.svc.Rename(m.ctx, m.targetID, val); err != nil {
		m.err = err
	}
	m.reload()
}

func (m *model) commitTags(val string) {
	if m.targetID == 0 {
		return
	}
	var tags []string
	if val != "" {
		tags = strings.Split(val, ",")
	}
	if _, err := m.svc.SetTags(m.ctx, m.targetID, tags); err != nil {
		m.err = err
	}
	m.reload()
}

func (m *model) commitDue(val string) {
	if m.targetID == 0 {
		return
	}
	var due *time.Time
	if val != "" && val != "none" {
		parsed, err := time.Parse(time.DateOnly, val)
		if err != nil {
			m.err = fmt.Errorf("invalid date %q (expected YYYY-MM-DD)", val)
			return
		}
		due = &parsed
	}
	if _, err := m.svc.SetDueAt(m.ctx, m.targetID, due); err != nil {
		m.err = err
	}
	m.reload()
}

func (m *model) focusTask(id int) {
	for i, t := range m.tasks {
		if t.ID == id {
			m.cursor = i
			return
		}
	}
}

// statusCycle is the rotation order for the `s` keybind. The empty status
// stands for "no filter" (show all).
var statusCycle = []task.Status{
	"",
	task.StatusTodo,
	task.StatusInProgress,
	task.StatusBlocked,
	task.StatusDone,
}

// priorityCycle is the rotation order for the `p` keybind. The empty
// priority stands for "unset".
var priorityCycle = []task.Priority{
	"",
	task.PriorityLow,
	task.PriorityMedium,
	task.PriorityHigh,
}

// recurrenceCycle is the rotation order for the `R` keybind. Interval forms
// like "2w" are reachable from the CLI; the TUI cycles keywords only.
var recurrenceCycle = []task.Recurrence{
	task.RecurNone,
	task.RecurDaily,
	task.RecurWeekly,
	task.RecurMonthly,
	task.RecurWeekdays,
	task.RecurWeekends,
}

// nextInCycle returns the value immediately after cur in cycle, wrapping
// around. cur not found in cycle yields the first element.
func nextInCycle[T comparable](cycle []T, cur T) T {
	for i, v := range cycle {
		if v == cur {
			return cycle[(i+1)%len(cycle)]
		}
	}
	return cycle[0]
}

func (m *model) cycleStatusFilter() {
	var cur task.Status
	if m.statusFilter != nil {
		cur = *m.statusFilter
	}
	next := nextInCycle(statusCycle, cur)
	if next == "" {
		m.statusFilter = nil
	} else {
		m.statusFilter = &next
	}
	m.cursor = 0
}

func (m *model) cyclePriority() {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	next := nextInCycle(priorityCycle, t.Priority)
	if _, err := m.svc.SetPriority(m.ctx, t.ID, next); err != nil {
		m.err = err
		return
	}
	m.reload()
}

func (m *model) cycleRecurrence() {
	t, ok := m.currentTask()
	if !ok {
		return
	}
	next := nextInCycle(recurrenceCycle, t.Recur)
	if _, err := m.svc.SetRecurrence(m.ctx, t.ID, next); err != nil {
		m.err = err
		return
	}
	m.reload()
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

// nextStatus advances through the three active states (todo → in-progress →
// done → todo) when the user hits space or enter. "blocked" is left out of
// this rotation because it usually marks a state the user discovers rather
// than chooses — it is reachable via the regular `kl status` command.
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
	b.WriteString(titleStyle.Render(" komlist") + "\n\n")
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
			b.WriteString(selectedRowStyle.Width(m.rowWidth()).
				Render(" "+render.TaskLinePlain(t, m.blocked[t.ID])) + "\n")
		} else {
			b.WriteString(m.renderRow(t) + "\n")
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
		return helpStyle.Render(" rename: ") + m.input.View()
	case modeEditTags:
		return helpStyle.Render(" tags: ") + m.input.View()
	case modeEditDue:
		return helpStyle.Render(" due: ") + m.input.View()
	case modeFilterTag:
		return helpStyle.Render(" tag filter: ") + m.input.View()
	case modeConfirmDelete:
		return helpStyle.Render(fmt.Sprintf(" Delete #%d? (y/n)", m.targetID))
	default:
		line1 := " ↑↓ nav · ⏎ cycle · d done · a add · e rename · p prio · t tags · u due · R recur · x del"
		line2 := " g group · s status · f tag · c clear · r reload · q quit"
		return helpStyle.Render(line1) + "\n" + helpStyle.Render(line2)
	}
}

// rowWidth returns the width used to pad the selection bar to full-width,
// falling back to defaultRowWidth before the first WindowSizeMsg arrives.
func (m model) rowWidth() int {
	if m.width > 0 {
		return m.width
	}
	return defaultRowWidth
}

// renderRow lays out a non-cursor row as "  <left> ... <right>" with the
// right half (tag pills + recur + due) flushed to the terminal's right
// edge. When the row has nothing right-aligned, it falls back to just the
// left content.
func (m model) renderRow(t task.Task) string {
	left := "  " + render.TaskLineLeft(t, m.blocked[t.ID])
	right := render.TaskLineRight(t)
	if right == "" {
		return left
	}
	filler := m.rowWidth() - lipgloss.Width(left) - lipgloss.Width(right) - 1
	if filler < 1 {
		filler = 1
	}
	return left + strings.Repeat(" ", filler) + right
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
