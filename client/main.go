package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

//split model each frame got a model then one UI controller handle switch logic
//Terminal client states
const (
	NotConnected = 1
	Connecting = 2
	Connected = 3
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	state 	  int 
	spinner   spinner.Model
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "localhost 8080"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	sp := spinner.New()
	sp.Spinner = spinner.Line

	return model{
		state: NotConnected,
		spinner: sp,
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.state++
			return m, m.spinner.Tick
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	if(m.state == NotConnected){
		m.textInput, cmd = m.textInput.Update(msg)
	}

	if(m.state == Connecting){
		m.spinner, cmd = m.spinner.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	var str string
	switch s := (m.state); s {
		case NotConnected: 
			str = fmt.Sprintf( "What's host ip address port?\n\n%s\n\n%s",
					m.textInput.View(),
					"(esc to quit)",
					) + "\n"
		case Connecting:
			str = fmt.Sprintf("\n\n %s loading forever now ... press esc to quit\n\n", m.spinner.View())
	}
	return str
}
