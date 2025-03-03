package multiplexer

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/adammpkins/OmniPath/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

type multiplexerModel struct {
	sessions    []*tui.Session
	activeIndex int
	updateCh    chan struct{}
	mu          sync.Mutex
}

func NewMultiplexerModel(sessions []*tui.Session) multiplexerModel {
	m := multiplexerModel{
		sessions:    sessions,
		activeIndex: 0,
		updateCh:    make(chan struct{}, 1),
	}
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

func (m multiplexerModel) Init() tea.Cmd {
	return func() tea.Msg {
		<-m.updateCh
		return struct{}{}
	}
}

func (m multiplexerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			for _, sess := range m.sessions {
				if strings.Contains(strings.ToLower(sess.Name), "sail") {
					log.Println("Detected Laravel Sail; running './vendor/bin/sail down'")
					cmd := exec.Command("./vendor/bin/sail", "down")
					if err := cmd.Run(); err != nil {
						log.Printf("Error shutting down Laravel Sail: %v", err)
					}
				}
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

func (m multiplexerModel) View() string {
	headerLines := []string{"Sessions:"}
	for i, sess := range m.sessions {
		marker := "  "
		if i == m.activeIndex {
			marker = "> "
		}
		headerLines = append(headerLines, fmt.Sprintf("%s%d: %s", marker, i, sess.Name))
	}
	const headerHeight = 6
	for len(headerLines) < headerHeight {
		headerLines = append(headerLines, "")
	}
	header := strings.Join(headerLines, "\n")
	content := header + "\n\n--- Active Session Output ---\n" + m.sessions[m.activeIndex].Output
	return content
}

func RunMultiplexer(sessions []*tui.Session) error {
	m := NewMultiplexerModel(sessions)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
