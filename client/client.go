package main 

import (
  "fmt"
  "log"

  "github.com/charmbracelet/bubbles/textinput"
  tea "github.com/charmbracelet/bubbletea"
)

func main() {
  p := tea.NewProgram(initialModel())
  fmt.Println("Hello this go worldoooo")
}
