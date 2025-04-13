package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_FocusMovement(t *testing.T) {
	tests := []struct {
		name          string
		initialFocus  int
		keyMsg        tea.KeyMsg
		expectedFocus int
	}{
		{
			name:          "Tab from Prefix to Summary",
			initialFocus:  0,
			keyMsg:        tea.KeyMsg{Type: tea.KeyTab},
			expectedFocus: 1,
		},
		{
			name:          "Tab from Summary to Description",
			initialFocus:  1,
			keyMsg:        tea.KeyMsg{Type: tea.KeyTab},
			expectedFocus: 2,
		},
		{
			name:          "Shift+Tab from Description to Summary",
			initialFocus:  2,
			keyMsg:        tea.KeyMsg{Type: tea.KeyShiftTab},
			expectedFocus: 1,
		},
		{
			name:          "Shift+Tab from Prefix to Quit (wrap-around)",
			initialFocus:  0,
			keyMsg:        tea.KeyMsg{Type: tea.KeyShiftTab},
			expectedFocus: 4,
		},
	}

	for _, tt := range tests {
		m := newCommitModel()
		m.focusIndex = tt.initialFocus
		m.summaryEditing = false
		m.descEditing = false

		_, _ = m.Update(tt.keyMsg)
		if m.focusIndex != tt.expectedFocus {
			t.Errorf("%s: expected focusIndex %d, got %d", tt.name, tt.expectedFocus, m.focusIndex)
		}
	}
}

func TestUpdate_QuitOnQ(t *testing.T) {
	tests := []struct {
		name         string
		initialFocus int
		keyMsg       tea.KeyMsg
		expectQuit   bool
	}{
		{
			name:         "Quit from Summary field not editing",
			initialFocus: 1,
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			expectQuit:   true,
		},
		{
			name:         "Quit from Description field not editing",
			initialFocus: 2,
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			expectQuit:   true,
		},
		{
			name:         "Do not quit if Summary is being edited",
			initialFocus: 1,
			keyMsg:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")},
			expectQuit:   false,
		},
	}

	for _, tt := range tests {
		m := newCommitModel()
		m.focusIndex = tt.initialFocus
		if tt.name == "Do not quit if Summary is being edited" {
			m.summaryEditing = true
		} else {
			m.summaryEditing = false
			m.descEditing = false
		}

		_, _ = m.Update(tt.keyMsg)
		if m.quitSelected != tt.expectQuit {
			t.Errorf("%s: expected quitSelected to be %v, got %v", tt.name, tt.expectQuit, m.quitSelected)
		}
	}
}

func TestUpdate_SummaryEditing(t *testing.T) {
	tests := []struct {
		name            string
		key             tea.KeyMsg
		expectedEditing bool
	}{
		{
			name:            "Start editing summary with 'i'",
			key:             tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")},
			expectedEditing: true,
		},
		{
			name:            "Start editing summary with Enter",
			key:             tea.KeyMsg{Type: tea.KeyEnter},
			expectedEditing: true,
		},
	}

	for _, tt := range tests {
		m := newCommitModel()
		m.focusIndex = 1
		m.summaryEditing = false
		_, _ = m.Update(tt.key)
		if m.summaryEditing != tt.expectedEditing {
			t.Errorf("%s: expected summaryEditing to be %v, got %v", tt.name, tt.expectedEditing, m.summaryEditing)
		}
	}
}

func TestUpdate_DescriptionEditing(t *testing.T) {
	tests := []struct {
		name            string
		key             tea.KeyMsg
		expectedEditing bool
	}{
		{
			name:            "Start editing description with 'i'",
			key:             tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")},
			expectedEditing: true,
		},
		{
			name:            "Start editing description with Enter",
			key:             tea.KeyMsg{Type: tea.KeyEnter},
			expectedEditing: true,
		},
	}

	for _, tt := range tests {
		m := newCommitModel()
		m.focusIndex = 2
		m.descEditing = false
		_, _ = m.Update(tt.key)
		if m.descEditing != tt.expectedEditing {
			t.Errorf("%s: expected descEditing to be %v, got %v", tt.name, tt.expectedEditing, m.descEditing)
		}
	}
}

func TestUpdate_ButtonSelection(t *testing.T) {
	tests := []struct {
		name         string
		initialFocus int
		keyMsg       tea.KeyMsg
		expectCommit bool
		expectQuit   bool
	}{
		{
			name:         "Commit button selected",
			initialFocus: 3,
			keyMsg:       tea.KeyMsg{Type: tea.KeyEnter},
			expectCommit: true,
			expectQuit:   false,
		},
		{
			name:         "Quit button selected",
			initialFocus: 4,
			keyMsg:       tea.KeyMsg{Type: tea.KeyEnter},
			expectCommit: false,
			expectQuit:   true,
		},
	}

	for _, tt := range tests {
		m := newCommitModel()
		m.focusIndex = tt.initialFocus
		_, _ = m.Update(tt.keyMsg)
		if m.commitSelected != tt.expectCommit {
			t.Errorf("%s: expected commitSelected to be %v, got %v", tt.name, tt.expectCommit, m.commitSelected)
		}
		if m.quitSelected != tt.expectQuit {
			t.Errorf("%s: expected quitSelected to be %v, got %v", tt.name, tt.expectQuit, m.quitSelected)
		}
	}
}

func TestUpdate_PrefixDropdown(t *testing.T) {
	tests := []struct {
		name                       string
		initialDropdownOpen        bool
		keyMsg                     tea.KeyMsg
		expectedDropdownOpen       bool
		expectedCurrentPrefixIndex int
	}{
		{
			name:                       "Open dropdown with Enter",
			initialDropdownOpen:        false,
			keyMsg:                     tea.KeyMsg{Type: tea.KeyEnter},
			expectedDropdownOpen:       true,
			expectedCurrentPrefixIndex: 0,
		},
		{
			name:                       "Close dropdown with Esc",
			initialDropdownOpen:        true,
			keyMsg:                     tea.KeyMsg{Type: tea.KeyEsc},
			expectedDropdownOpen:       false,
			expectedCurrentPrefixIndex: 0,
		},
		{
			name:                       "Select option in dropdown and close with Enter",
			initialDropdownOpen:        true,
			keyMsg:                     tea.KeyMsg{Type: tea.KeyEnter},
			expectedDropdownOpen:       false,
			expectedCurrentPrefixIndex: 0,
		},
	}

	for _, tt := range tests {
		m := newCommitModel()
		m.focusIndex = 0
		m.prefixDropdownOpen = tt.initialDropdownOpen
		m.dropdownIndex = 0
		_, _ = m.Update(tt.keyMsg)
		if m.prefixDropdownOpen != tt.expectedDropdownOpen {
			t.Errorf("%s: expected prefixDropdownOpen to be %v, got %v", tt.name, tt.expectedDropdownOpen, m.prefixDropdownOpen)
		}
		if m.currentPrefixIndex != tt.expectedCurrentPrefixIndex {
			t.Errorf("%s: expected currentPrefixIndex %d, got %d", tt.name, tt.expectedCurrentPrefixIndex, m.currentPrefixIndex)
		}
	}
}

func TestView_Output(t *testing.T) {
	m := newCommitModel()
	m.focusIndex = 1

	view := m.View()
	if !strings.Contains(view, "Summary:") {
		t.Error("View output should contain 'Summary:'")
	}
	if !strings.Contains(view, "Prefix:") {
		t.Error("View output should contain 'Prefix:'")
	}
	if !strings.Contains(view, "Description:") {
		t.Error("View output should contain 'Description:'")
	}
	if !strings.Contains(view, "[ Commit ]") {
		t.Error("View output should contain '[ Commit ]'")
	}
	if !strings.Contains(view, "[ Quit ]") {
		t.Error("View output should contain '[ Quit ]'")
	}
}
