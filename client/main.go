package main

import (
	"fmt"
  "net"
  "os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

/* client connection to the tcp server */
var conn net.Conn

/* client tui states */
const (
	NotConnected = 1 
	Connecting = 2
	Connected = 3
)

type (
	errMsg error
	connectionMsg struct{} 
)

/* model */
type model struct {
	state 	  int 
	spinner   spinner.Model
	textInput textinput.Model
	err       error
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
    fmt.Printf("Got an error : %v\n", err);
    os.Exit(1)
	}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "localhost:8080"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	sp := spinner.New()
	sp.Spinner = spinner.Line

	return model{
		state:      NotConnected,
		spinner:    sp,
		textInput:  ti,
		err:        nil,
	}
}

/* Connect : tries to establish a tcp connection with argument inside textInput */
func Connect(m model) tea.Cmd {
	return func() tea.Msg {
    var err error
    conn, err = net.Dial("tcp", m.textInput.Value())
    if err != nil {
      return errMsg(err)
    }
		return connectionMsg{} 
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
      if(m.state == NotConnected){
			  return m, Connect(m) 
      }
		}

  // Connection succeed
  case connectionMsg:
    m.state++
    return m, m.spinner.Tick  

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, tea.Quit 
	}

	if(m.state == NotConnected){
		m.textInput, cmd = m.textInput.Update(msg)
	}

	if(m.state == Connecting){
		m.spinner, cmd = m.spinner.Update(msg)
	}
	return m, cmd
}

func (m model) View() string{
  if m.err != nil {
    return fmt.Sprintf("\n\nerror -> %v\n\n", m.err) + "\n"
  }

	if(m.state == Connecting){
		return ConnectingView(m)
	}
	return NotConnectedView(m)
}

func NotConnectedView(m model) string {
	return fmt.Sprintf("What's host ip address port?\n\n%s\n\n%s",
					m.textInput.View(),
					"press esc to quit") + "\n"
}

func ConnectingView(m model) string {
  conn.Write([]byte("FuckOff"))
  return fmt.Sprintf("\n\n %s wait till connect to \n\n", m.spinner.View()) + "\n"
}
