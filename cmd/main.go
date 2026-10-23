package main

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/tylorkolbeck/k8s-dashboard-tui/internal/logger"
)

type SystemStats struct {
	prevCpu cpu.Stats
	memory  memory.Stats
}

type model struct {
	choices          []string
	cursor           int
	selected         map[int]struct{}
	log              *logger.Logger
	tickerDone       chan struct{}
	systemStats      SystemStats
	cpuUsageProgress progress.Model
	memUsageProgress progress.Model
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, theres been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel() model {
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Println("fatal: ", err)
		os.Exit(1)
	}

	prevCpu, err := cpu.Get()
	if err != nil {
		fmt.Printf("could not fetch CPU stats")
		os.Exit(1)
	}

	mem, err := memory.Get()
	if err != nil {
		fmt.Printf("could not fetch CPU stats")
		os.Exit(1)
	}

	return model{
		choices:    []string{"Carrots", "celery", "bread"},
		selected:   make(map[int]struct{}),
		log:        log,
		tickerDone: make(chan struct{}),
		systemStats: SystemStats{
			prevCpu: *prevCpu,
			memory:  *mem,
		},
		cpuUsageProgress: progress.New(progress.WithDefaultBlend()),
		memUsageProgress: progress.New(progress.WithDefaultBlend()),
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		currentCpu, err := cpu.Get()
		if err != nil {
			m.log.Write(fmt.Sprintf("cpu fetch error: %v", err))
			return m, tickCmd()
		}

		var memProgressCmd tea.Cmd
		mem, err := memory.Get()
		if err != nil {
			m.log.Write(fmt.Sprintf("memory fetch error: %v", err))
		} else {
			// Convert to 0.0 - 1.0 scale for the progress bar
			memPercent := float64(mem.Used) / float64(mem.Total)
			memProgressCmd = m.memUsageProgress.SetPercent(memPercent)
		}

		// Calculate Usage
		totalDelta := float64(currentCpu.Total - m.systemStats.prevCpu.Total)
		idlePercent := (float64(currentCpu.Idle-m.systemStats.prevCpu.Idle) / totalDelta) * 100
		totalUsage := 100.0 - idlePercent

		// Update state for the next tick
		m.systemStats.prevCpu = *currentCpu
		progressCmd := m.cpuUsageProgress.SetPercent(totalUsage / 100.0)

		// Return a Batch command: animate the progress bar AND queue up the next tick
		return m, tea.Batch(progressCmd, memProgressCmd, tickCmd())
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.cpuUsageProgress, cmd = m.cpuUsageProgress.Update(msg)
		m.memUsageProgress, cmd = m.memUsageProgress.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			close(m.tickerDone)
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
	s += fmt.Sprintf("CPU: %v \n", m.cpuUsageProgress.View())
	s += fmt.Sprintf("Memory: %v \n", m.memUsageProgress.View())

	v := tea.NewView(s)
	v.AltScreen = true

	return v
}
