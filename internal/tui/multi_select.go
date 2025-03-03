package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// multiSelectItem wraps Service with a Selected flag.
type multiSelectItem struct {
	Service  Service
	Selected bool
}

func (m multiSelectItem) Title() string {
	checkbox := "[ ]"
	if m.Selected {
		checkbox = "[x]"
	}
	return fmt.Sprintf("%s %s", checkbox, m.Service.Name)
}

func (m multiSelectItem) Description() string {
	return m.Service.Command
}

func (m multiSelectItem) FilterValue() string {
	return m.Service.Name
}

// multiSelectDelegate is a custom delegate for rendering items.
type multiSelectDelegate struct{}

func (d multiSelectDelegate) Height() int                               { return 1 }
func (d multiSelectDelegate) Spacing() int                              { return 0 }
func (d multiSelectDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d multiSelectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	cursor := "  "
	if index == m.Cursor() {
		cursor = "> "
	}
	if mi, ok := item.(multiSelectItem); ok {
		_, _ = fmt.Fprint(w, cursor+mi.Title())
	}
}

// multiSelectModel defines our multi-select UI model.
type multiSelectModel struct {
	list         list.Model
	selected     []Service
	instructions string
}

func NewMultiSelectModel(services []Service) *multiSelectModel {
	items := make([]list.Item, len(services))
	for i, s := range services {
		items[i] = multiSelectItem{Service: s, Selected: false}
	}
	height := len(items) + 2
	if height < 20 {
		height = 20
	}
	delegate := multiSelectDelegate{}
	l := list.New(items, delegate, 40, height)
	l.Title = "Select Services to Run"
	return &multiSelectModel{
		list:         l,
		instructions: "Use ↑/↓ to navigate, SPACE to toggle selection, and ENTER to confirm.",
	}
}

func (m *multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m *multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "down", "j":
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case " ":
			i := m.list.Cursor()
			if item, ok := m.list.Items()[i].(multiSelectItem); ok {
				item.Selected = !item.Selected
				m.list.SetItem(i, item)
			}
			return m, nil
		case "enter":
			var selected []Service
			for _, item := range m.list.Items() {
				if mi, ok := item.(multiSelectItem); ok && mi.Selected {
					selected = append(selected, mi.Service)
				}
			}
			if len(selected) == 0 {
				if item, ok := m.list.Items()[m.list.Cursor()].(multiSelectItem); ok {
					selected = append(selected, item.Service)
				}
			}
			m.selected = selected
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *multiSelectModel) View() string {
	var b strings.Builder
	b.WriteString(m.instructions + "\n\n")
	b.WriteString(m.list.View())
	b.WriteString("\nPress q to quit.\n")
	return b.String()
}

func RunMultiSelect(services []Service) ([]Service, error) {
	model := NewMultiSelectModel(services)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	m, ok := finalModel.(*multiSelectModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}
	return m.selected, nil
}
