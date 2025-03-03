package multiplexer

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/adammpkins/OmniPath/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// multiplexerModel is our Bubbletea model for managing sessions in fallback mode.
type multiplexerModel struct {
	sessions    []*tui.Session
	activeIndex int
	updateCh    chan struct{}
	mu          sync.Mutex
}

// NewMultiplexerModel creates a new multiplexer model from a slice of session pointers.
func NewMultiplexerModel(sessions []*tui.Session) multiplexerModel {
	m := multiplexerModel{
		sessions:    sessions,
		activeIndex: 0,
		updateCh:    make(chan struct{}, 1),
	}
	// Trigger periodic UI updates.
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			m.triggerUpdate()
		}
	}()
	return m
}

func (m *multiplexerModel) triggerUpdate() {
	select {
	case m.updateCh <- struct{}{}:
	default:
	}
}

// Init implements the tea.Model interface.
func (m multiplexerModel) Init() tea.Cmd {
	return func() tea.Msg {
		<-m.updateCh
		return struct{}{}
	}
}

// Update handles key events and forwards them to sessions.
func (m multiplexerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Send SIGINT to each session's process group.
			for _, sess := range m.sessions {
				if sess.Cmd != nil && sess.Cmd.Process != nil {
					if pgid, err := syscall.Getpgid(sess.Cmd.Process.Pid); err == nil {
						syscall.Kill(-pgid, syscall.SIGINT)
					}
				}
			}
			return m, tea.Quit
		case "left", "h":
			if m.activeIndex > 0 {
				m.activeIndex--
			}
		case "right", "l":
			if m.activeIndex < len(m.sessions)-1 {
				m.activeIndex++
			}
		default:
			// Forward key input to the active session's stdin.
			active := m.sessions[m.activeIndex]
			if active.Stdin != nil {
				_, _ = active.Stdin.Write([]byte(msg.String()))
			}
		}
	}
	return m, func() tea.Msg {
		<-m.updateCh
		return struct{}{}
	}
}

// View renders the multiplexer UI.
func (m multiplexerModel) View() string {
	headerLines := []string{"Sessions:"}
	for i, sess := range m.sessions {
		marker := "  "
		if i == m.activeIndex {
			marker = "> "
		}
		headerLines = append(headerLines, fmt.Sprintf("%s%d: %s", marker, i, sess.Name))
	}
	// Pad header to a fixed height.
	const headerHeight = 6
	for len(headerLines) < headerHeight {
		headerLines = append(headerLines, "")
	}
	header := strings.Join(headerLines, "\n")
	content := header + "\n\n--- Active Session Output ---\n" + m.sessions[m.activeIndex].Output
	return content
}

// RunMultiplexer launches the multiplexer UI.
func RunMultiplexer(sessions []*tui.Session) error {
	m := NewMultiplexerModel(sessions)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
