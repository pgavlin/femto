package femto

import (
	"bytes"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Messenger is an object that makes it easy to send messages to the user
// and get input from the user
type Messenger struct {
	log *Buffer
	// Are we currently prompting the user?
	hasPrompt bool
	// Is there a message to print
	hasMessage bool

	// Message to print
	message string
	// The user's response to a prompt
	response string
	// style to use when drawing the message
	style tcell.Style

	// We have to keep track of the cursor for prompting
	cursorx int

	// This map stores the history for all the different kinds of uses Prompt has
	// It's a map of history type -> history array
	history    map[string][]string
	historyNum int

	// Is the current message a message from the gutter
	gutterMessage bool
}

func (m *Messenger) Reset() {
	m.cursorx = 0
	m.message = ""
	m.response = ""
}

// Clear clears the line at the bottom of the editor
func (m *Messenger) Clear() {
}

// AddLog sends a message to the log view
func (m *Messenger) AddLog(msg ...interface{}) {
	logMessage := fmt.Sprint(msg...)
	buffer := m.getBuffer()
	buffer.insert(buffer.End(), []byte(logMessage+"\n"))
	buffer.Cursor.Loc = buffer.End()
	buffer.Cursor.Relocate()
}

func (m *Messenger) getBuffer() *Buffer {
	if m.log == nil {
		m.log = NewBufferFromString("", "")
		m.log.name = "Log"
	}
	return m.log
}

// Message sends a message to the user
func (m *Messenger) Message(msg ...interface{}) {
	displayMessage := fmt.Sprint(msg...)
	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if m.hasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		m.message = displayMessage

		m.style = defStyle

		if _, ok := colorscheme["message"]; ok {
			m.style = colorscheme["message"]
		}

		m.hasMessage = true
	}
	// add the message to the log regardless of active prompts
	m.AddLog(displayMessage)
}

// Error sends an error message to the user
func (m *Messenger) Error(msg ...interface{}) {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, msg...)

	// only display a new message if there isn't an active prompt
	// this is to prevent overwriting an existing prompt to the user
	if m.hasPrompt == false {
		// if there is no active prompt then style and display the message as normal
		m.message = buf.String()
		m.style = defStyle.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorMaroon)

		if _, ok := colorscheme["error-message"]; ok {
			m.style = colorscheme["error-message"]
		}
		m.hasMessage = true
	}
	// add the message to the log regardless of active prompts
	m.AddLog(buf.String())
}
// A GutterMessage is a message displayed on the side of the editor
type GutterMessage struct {
	lineNum int
	msg     string
	kind    int
}

// These are the different types of messages
const (
	// GutterInfo represents a simple info message
	GutterInfo = iota
	// GutterWarning represents a compiler warning
	GutterWarning
	// GutterError represents a compiler error
	GutterError
)

