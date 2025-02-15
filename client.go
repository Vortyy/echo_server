package main

import (
	"fmt"
  "net"
  "os"
  "time"
  "strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

/* Styling var */
var (
  keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
  foregroundStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
  spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
  errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

//TODO array to handle multiple connection with tabs
/* global client connection to the tcp server */
var conn net.Conn

/* client tui states */
const (
	NotConnected = 1 
	Connecting = 2
	Connected = 3
  Closed = -1
)

type (
	errMsg error
	connectionMsg struct{} 
  recvMsg struct{msg string}
  closeMsg struct{}
)

/* model */
type model struct {
	state 	  int 
	spinner   spinner.Model
	textInput textinput.Model
  messages  []string
  viewport  viewport.Model
  textArea  textarea.Model
	err       error
}

func main() { 
  defer func() { /* if a connection has been establish close it */ 
    if conn != nil {
      conn.Close()
    }
  } ()

  p := tea.NewProgram(initialModel()) /* start bubbletea program */
	if _, err := p.Run(); err != nil {
    fmt.Printf("Got an error : %v\n", err);
    os.Exit(1)
	}
}

func initialModel() model {
  /* NotConnected view */
	ti := textinput.New()
	ti.Placeholder = "localhost:8080"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

  /* Connecting view */
	sp := spinner.New()
	sp.Spinner = spinner.Points
  sp.Style = spinnerStyle

  /* Connected view */
  ta := textarea.New()
  ta.Placeholder = "Send a message..."
  ta.Focus()

  ta.Prompt = "┃ "
  ta.CharLimit = 280
  
  ta.SetWidth(50)
  ta.SetHeight(3)
  ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
  ta.ShowLineNumbers = false
  ta.KeyMap.InsertNewline.SetEnabled(false)

  vp := viewport.New(50, 5)

	return model{
		state:      NotConnected,
    messages:   []string{},
		spinner:    sp,
		textInput:  ti,
    viewport:   vp,
    textArea:   ta,
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
    
    //Wait just to see the beautifull spinner :)
    time.Sleep(2 * time.Second)

		return connectionMsg{} 
	}
}

/* Send : tries to send a msg to the server and wait/read for the response */
func Send(m model, send string) tea.Cmd {
  return func() tea.Msg {
    //Send the messages
    _, err := conn.Write([]byte(send))
    if err != nil {
      return errMsg(err)
    }
    
    //Buffer for read
    reply := make([]byte, 280) //TODO check BUFSIZE ??? i think 8192
    //Wait and read the server reply
    var n int
    n, err = conn.Read(reply)
    if err != nil {
      return errMsg(err)
    }

    return recvMsg{msg: foregroundStyle.Render("server: ") + string(reply[0:n])}  
  }
}

/* Close : tries to close the current opened conn */
func Close() tea.Msg {
  err := conn.Close()
  if err != nil {
    return errMsg(err)
  }
  return closeMsg{}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

/* Update func */
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  /* this key always quit */
  if msg, ok := msg.(tea.KeyMsg); ok {
    k := msg.String()
    if k == "esc" {
      m.state = Closed
      return m, tea.Quit
    }
  }

  if(m.state == Connecting){
    return updateConnecting(msg, m)
  }

  if(m.state == Connected){
    return updateConnected(msg, m)
  }

  return updateNotConnected(msg, m)
}

func updateNotConnected(msg tea.Msg, m model) (tea.Model, tea.Cmd){
  switch msg := msg.(type) {
    case tea.KeyMsg:
    switch msg.Type {
      case tea.KeyEnter:
        m.state++
        return m, tea.Batch(Connect(m), m.spinner.Tick) 
    }
  }

  var cmd tea.Cmd
  m.textInput, cmd = m.textInput.Update(msg)

  return m, cmd
}

func updateConnecting(msg tea.Msg, m model) (tea.Model, tea.Cmd){
  switch msg := msg.(type) {
    case connectionMsg:
      m.messages = nil /* reset previous messages connection */ 
      m.state++
      return m, tea.Batch(textarea.Blink, tea.WindowSize())  

    // We handle errors just like any other message
    case errMsg:
      m.state--
      m.err = msg
      return m, nil  
  }

  var cmd tea.Cmd
  m.spinner, cmd = m.spinner.Update(msg)

  return m, cmd
}

func updateConnected(msg tea.Msg, m model) (tea.Model, tea.Cmd){
  var (
    tiCmd tea.Cmd
    vpCmd tea.Cmd
  )

  m.textArea, tiCmd = m.textArea.Update(msg)
  m.viewport, vpCmd = m.viewport.Update(msg)

  switch msg := msg.(type) {
  case tea.WindowSizeMsg:
    m.viewport.Width = msg.Width 
    m.textArea.SetWidth(msg.Width)
    m.viewport.Height = msg.Height - m.textArea.Height() - lipgloss.Height("\n\n\n\n") //gap + helpline + connectedline
    if len(m.messages) > 0 {
      m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
    } else {
      m.viewport.SetContent("send something...")
    }
    m.viewport.GotoBottom()

  case tea.KeyMsg:
    switch msg.Type {
    case tea.KeyEnter:
      str := m.textArea.Value()
      m.messages = append(m.messages, foregroundStyle.Render("you: ") + str) 
      m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
      m.textArea.Reset()
      m.viewport.GotoBottom()
      return m, Send(m, str) 
    case tea.KeyCtrlC:
      m.state=NotConnected
      return m, tea.Batch(Close, textinput.Blink)
    }

  case recvMsg:
    m.messages = append(m.messages, msg.msg) 
    m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
    m.viewport.GotoBottom()

  case errMsg:
    m.state = NotConnected
    return m, tea.Batch(Close, textinput.Blink)
  }

  return m, tea.Batch(tiCmd, vpCmd) 
}

/* Main view */
func (m model) View() string{
	if(m.state == Connecting){
		return ConnectingView(m)
	}

  if(m.state == Connected){
    return ConnectedView(m)
  }

  if(m.state == Closed){
    return ClosingView() 
  }

	return NotConnectedView(m)
}

func NotConnectedView(m model) string {
  var str string
  str += "What's host "
  str += foregroundStyle.Render("ip:port") 
  str += " ? " 

  if m.err != nil {
    str += errStyle.Render(fmt.Sprintf("%v", m.err))
  }

  str += "\n\n"
  str += m.textInput.View()
  str += "\n\n"

  str += helpStyle.Render(" 󰌑 : to start connection • esc: exit")
  
  return str + "\n"
}

func ConnectingView(m model) string {
  return fmt.Sprintf("\n\n %s wait till connect to %s\n\n", 
          m.spinner.View(),
          foregroundStyle.Render(m.textInput.Value())) + "\n"
}

func ConnectedView(m model) string {
  var str string
  str += fmt.Sprintf("Connected on %s\n", conn.RemoteAddr().String())
  str += m.viewport.View()
  str += "\n\n"
  str += m.textArea.View()
  str += "\n\n" + helpStyle.Render(" 󰌑 : send a msg •  , : move along viewport • esc: exit • ctrl+c: close connection")
  return str
}

func ClosingView() string {
  return fmt.Sprintf("Bye Bye ;)\n")
}
