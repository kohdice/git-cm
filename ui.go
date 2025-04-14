package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// commitModel is the model that holds the state of the TUI.
// (Note) Focus indexes:
//
//	0: Prefix, 1: Summary, 2: Description, 3: Commit, 4: Quit
//
// Also, summaryEditing and descEditing are used to manage the input mode triggered by the "i" or "enter" key.
type commitModel struct {
	prefixOptions      []string
	currentPrefixIndex int
	dropdownIndex      int
	prefixDropdownOpen bool

	summary        textinput.Model
	desc           textarea.Model
	focusIndex     int // 0: Prefix, 1: Summary, 2: Description, 3: Commit, 4: Quit
	summaryEditing bool
	descEditing    bool

	commitSelected bool
	quitSelected   bool
}

// newCommitModel initializes and returns a new commitModel.
func newCommitModel() *commitModel {
	m := &commitModel{
		prefixOptions:      []string{"feat", "fix", "refactor", "test", "style", "chore", "docs"},
		currentPrefixIndex: 0,
		focusIndex:         0,
		prefixDropdownOpen: false,
		dropdownIndex:      0,
		summaryEditing:     false,
		descEditing:        false,
	}

	// Initialize the text input field for Summary.
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 100
	ti.Width = 50
	m.summary = ti

	// Initialize the text area for Description (allows multi-line input).
	ta := textarea.New()
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	m.desc = ta

	return m
}

// Init is the command that is executed initially (unused in this case).
func (m *commitModel) Init() tea.Cmd {
	return nil
}

// Update processes the user input events.
func (m *commitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Force quit on Ctrl+C.
		if msg.Type == tea.KeyCtrlC {
			m.quitSelected = true
			return m, tea.Quit
		}

		// Quit when "q" is pressed and not in input mode.
		if !m.summaryEditing && !m.descEditing && msg.String() == "q" {
			m.quitSelected = true
			return m, tea.Quit
		}

		// Global focus movement: Tab / Shift+Tab.
		switch msg.String() {
		case "tab":
			// Exit input mode for Summary and Description.
			if m.focusIndex == 1 && m.summaryEditing {
				m.summaryEditing = false
				m.summary.Blur()
			}
			if m.focusIndex == 2 && m.descEditing {
				m.descEditing = false
				m.desc.Blur()
			}
			if m.focusIndex == 0 && m.prefixDropdownOpen {
				m.prefixDropdownOpen = false
			}
			m.focusIndex = (m.focusIndex + 1) % 5
			return m, nil

		case "shift+tab":
			if m.focusIndex == 1 && m.summaryEditing {
				m.summaryEditing = false
				m.summary.Blur()
			}
			if m.focusIndex == 2 && m.descEditing {
				m.descEditing = false
				m.desc.Blur()
			}
			if m.focusIndex == 0 && m.prefixDropdownOpen {
				m.prefixDropdownOpen = false
			}
			m.focusIndex = (m.focusIndex - 1 + 5) % 5
			return m, nil
		}

		switch m.focusIndex {
		case 0: // Operations for the Prefix.
			if m.prefixDropdownOpen {
				switch msg.String() {
				case "up", "k":
					if m.dropdownIndex > 0 {
						m.dropdownIndex--
					}
				case "down", "j":
					if m.dropdownIndex < len(m.prefixOptions)-1 {
						m.dropdownIndex++
					}
				case "enter":
					m.currentPrefixIndex = m.dropdownIndex
					m.prefixDropdownOpen = false
				case "esc":
					m.prefixDropdownOpen = false
				}
				return m, nil
			} else {
				if msg.String() == "enter" {
					m.prefixDropdownOpen = true
					m.dropdownIndex = m.currentPrefixIndex
					return m, nil
				}
			}
		case 1: // For the Summary input field.
			// Start input mode when "i" or "enter" is pressed and not already editing.
			if (msg.String() == "i" || msg.String() == "enter") && !m.summaryEditing {
				m.summaryEditing = true
				m.summary.Focus()
				return m, nil
			}
			// Exit input mode when "esc" is pressed.
			if msg.String() == "esc" && m.summaryEditing {
				m.summaryEditing = false
				m.summary.Blur()
				return m, nil
			}
			if m.summaryEditing {
				var cmd tea.Cmd
				m.summary, cmd = m.summary.Update(msg)
				return m, cmd
			}
		case 2: // For the Description input area.
			if (msg.String() == "i" || msg.String() == "enter") && !m.descEditing {
				m.descEditing = true
				m.desc.Focus()
				return m, nil
			}
			if msg.String() == "esc" && m.descEditing {
				m.descEditing = false
				m.desc.Blur()
				return m, nil
			}
			if m.descEditing {
				var cmd tea.Cmd
				m.desc, cmd = m.desc.Update(msg)
				return m, cmd
			}
		case 3: // Commit button selected.
			if msg.String() == "enter" {
				m.commitSelected = true
				return m, tea.Quit
			}
		case 4: // Quit button selected.
			if msg.String() == "enter" {
				m.quitSelected = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View returns a string that represents the current state of the model for rendering.
func (m *commitModel) View() string {
	// Styles for labels.
	focusLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f7b977")).Bold(true)
	noFocusLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	// Style for displaying input values set to white (here "#e6eae6").
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e6eae6"))

	var s string

	// Display Prefix.
	var prefixLabel string
	if m.focusIndex == 0 {
		prefixLabel = focusLabelStyle.Render("Prefix")
	} else {
		prefixLabel = noFocusLabelStyle.Render("Prefix")
	}
	s += prefixLabel + ": " + inputStyle.Render(m.prefixOptions[m.currentPrefixIndex]) + "\n"

	// If the dropdown is open, display the candidate list.
	if m.focusIndex == 0 && m.prefixDropdownOpen {
		for i, option := range m.prefixOptions {
			var line string
			if i == m.dropdownIndex {
				line = focusLabelStyle.Render("> " + option)
			} else {
				line = noFocusLabelStyle.Render("  " + option)
			}
			s += line + "\n"
		}
	}
	s += "\n"

	// Display Summary.
	var summaryLabel string
	if m.focusIndex == 1 {
		summaryLabel = focusLabelStyle.Render("Summary")
	} else {
		summaryLabel = noFocusLabelStyle.Render("Summary")
	}
	s += summaryLabel + ": " + inputStyle.Render(m.summary.View()) + "\n\n"

	// Display Description.
	var descriptionLabel string
	if m.focusIndex == 2 {
		descriptionLabel = focusLabelStyle.Render("Description")
	} else {
		descriptionLabel = noFocusLabelStyle.Render("Description")
	}
	s += descriptionLabel + ":\n" + inputStyle.Render(m.desc.View()) + "\n\n"

	// Display Commit and Quit buttons.
	var commitButton, quitButton string
	if m.focusIndex == 3 {
		commitButton = focusLabelStyle.Render("[ Commit ]")
	} else {
		commitButton = noFocusLabelStyle.Render("[ Commit ]")
	}
	if m.focusIndex == 4 {
		quitButton = focusLabelStyle.Render("[ Quit ]")
	} else {
		quitButton = noFocusLabelStyle.Render("[ Quit ]")
	}
	s += commitButton + "    " + quitButton + "\n"
	return s
}

// runTUI starts the TUI and returns a CommitMessage constructed
// from the final state of the TUI, or an error if something goes wrong.
// If the user chooses to quit, it returns errQuit.
func runTUI() (*commitMessage, error) {
	m := newCommitModel()
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error starting program: %w", err)
	}

	model, ok := final.(*commitModel)
	if !ok {
		return nil, fmt.Errorf("type assertion failed")
	}

	if model.quitSelected {
		return nil, errQuit
	}

	return &commitMessage{
		Prefix:      model.prefixOptions[model.currentPrefixIndex],
		Summary:     model.summary.Value(),
		Description: model.desc.Value(),
	}, nil
}
