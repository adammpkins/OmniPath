package tui

import (
	"fmt"

	"github.com/adammpkins/OmniPath/internal/docs"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// dependencyItem wraps docs.DependencyDocs so it satisfies the list.Item interface.
type dependencyItem docs.DependencyDocs

func (d dependencyItem) Title() string       { return d.Name }
func (d dependencyItem) Description() string { return d.DocURL }
func (d dependencyItem) FilterValue() string { return d.Name }

// selectorModel defines the Bubbletea model for our dependency selector.
type selectorModel struct {
	list list.Model
}

// newSelectorModel creates a new selector model with our dependency items.
// Here, we set the height to 20 to allow at least 10 items to be visible.
func newSelectorModel(deps []docs.DependencyDocs) selectorModel {
	items := make([]list.Item, len(deps))
	for i, dep := range deps {
		items[i] = dependencyItem(dep)
	}
	// Adjust width and height as needed. Here, height is increased to 20.
	l := list.New(items, list.NewDefaultDelegate(), 40, 20)
	l.Title = "Select Dependency"
	return selectorModel{list: l}
}

// Init is the initial command for our model.
func (m selectorModel) Init() tea.Cmd {
	return nil
}

// Update handles key events and other messages.
func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// When the user presses Enter, we quit the TUI.
		case "enter":
			return m, tea.Quit
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the list view.
func (m selectorModel) View() string {
	return m.list.View()
}

// SelectDependency launches the TUI and returns the dependency selected by the user.
func SelectDependency(deps []docs.DependencyDocs) (docs.DependencyDocs, error) {
	model := newSelectorModel(deps)
	// Use Run() to capture the final model.
	finalModel, err := tea.NewProgram(model).Run()
	if err != nil {
		return docs.DependencyDocs{}, err
	}

	// Assert the final model to our selectorModel type.
	m, ok := finalModel.(selectorModel)
	if !ok {
		return docs.DependencyDocs{}, fmt.Errorf("unexpected model type")
	}

	selectedItem := m.list.SelectedItem()
	if dep, ok := selectedItem.(dependencyItem); ok {
		return docs.DependencyDocs{
			Name:   dep.Name,
			DocURL: dep.DocURL,
		}, nil
	}
	return docs.DependencyDocs{}, fmt.Errorf("no dependency selected")
}
