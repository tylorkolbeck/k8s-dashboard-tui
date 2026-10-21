package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tylorkolbeck/k8s-dashboard-tui/internal/logger"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func main() {
	log, err := logger.NewLogger()
	// f, err := tea.LogToFile("logs/log.txt", "debug")
	if err != nil {
		fmt.Println("fatal: ", err)
		os.Exit(1)
	}

	log.Info("test info log")
	log.Warn("test warning log")
	log.Err("test error log")
	log.Write("some info log 1")
	log.Write("some info log 2")

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, theres been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel() model {
	return model{
		choices:  []string{"Carrots", "celery", "bread"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", "space":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	s := "What should we buy at the market?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\n Press q to quit.\n"

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
