package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"time-machine/internal/models"
)

// Styles for the UI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EE6FF8"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C")).
			Bold(true)
)

// Renderer handles UI rendering
type Renderer struct{}

// NewRenderer creates a new UI renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// View renders the complete UI
func (r *Renderer) View(m models.Model) string {
	if m.Quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(models.TitleMain))
	b.WriteString("\n\n")

	// Show loading state
	if m.Loading {
		b.WriteString(normalStyle.Render(models.TextLoading + m.LoadingText))
		b.WriteString("\n\n")
		return b.String()
	}

	// Show error if any
	if m.Err != nil {
		b.WriteString(errorStyle.Render(" ⚠️ Ошибка: " + m.Err.Error()))
		b.WriteString("\n\n")
	}

	// Show sync message if needed
	if m.ShowSyncMessage {
		if m.SyncMessage != "" {
			b.WriteString(warningStyle.Render(" ⚠ " + m.SyncMessage))
		} else {
			b.WriteString(successStyle.Render(" ✓ Копия создана"))
		}
		b.WriteString("\n\n")
	}

	// Show description input mode
	if m.DescriptionMode {
		b.WriteString(r.renderDescriptionInput(m))
	} else if m.HistoryMode {
		b.WriteString(r.renderHistory(m))
	} else {
		// Show git status
		if m.Status != nil {
			b.WriteString(r.renderGitStatus(m.Status))
			b.WriteString("\n\n")
		}

		// Menu
		b.WriteString(r.renderMenu(m))
	}

	return b.String()
}

// renderDescriptionInput displays the description input interface
func (r *Renderer) renderDescriptionInput(m models.Model) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(models.TitleDescription))
	b.WriteString("\n\n")

	b.WriteString(normalStyle.Render(models.PromptDescription))
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("> " + m.DescriptionInput + "_"))
	b.WriteString("\n\n")

	b.WriteString(normalStyle.Render(models.PromptSuggestions))
	b.WriteString("\n")

	for i, suggestion := range m.Suggestions {
		if i < 10 { // Show only first 10 suggestions
			prefix := "   "
			if i < 9 {
				prefix = fmt.Sprintf(" [%d] ", i+1)
			}
			b.WriteString(normalStyle.Render(prefix + suggestion))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(normalStyle.Render(models.HelpDescription))

	return b.String()
}

// renderGitStatus displays the current git repository status
func (r *Renderer) renderGitStatus(status *models.GitStatus) string {
	var b strings.Builder

	// Branch info
	branchText := fmt.Sprintf("%s %s", models.LabelBranch, status.Branch)
	if status.Ahead > 0 || status.Behind > 0 {
		branchText += fmt.Sprintf(" (↑%d ↓%d)", status.Ahead, status.Behind)
	}
	b.WriteString(normalStyle.Render(branchText))
	b.WriteString("\n")

	// Last commit
	if status.LastCommit != "" {
		b.WriteString(normalStyle.Render(models.LabelLastCommit + " " + status.LastCommit))
		b.WriteString("\n")
	}

	// Status
	if status.IsClean {
		b.WriteString(successStyle.Render(models.TextClean))
	} else {
		b.WriteString(warningStyle.Render(models.TextDirty))
	}
	b.WriteString("\n\n")

	// File changes
	if len(status.Staged) > 0 {
		b.WriteString(successStyle.Render(models.LabelStaged))
		b.WriteString("\n")
		for _, file := range status.Staged {
			b.WriteString(normalStyle.Render("  ✓ " + file))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(status.Modified) > 0 {
		b.WriteString(warningStyle.Render(models.LabelModified))
		b.WriteString("\n")
		for _, file := range status.Modified {
			b.WriteString(normalStyle.Render("  • " + file))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(status.Untracked) > 0 {
		b.WriteString(normalStyle.Render(models.LabelUntracked))
		b.WriteString("\n")
		for _, file := range status.Untracked {
			b.WriteString(normalStyle.Render("  ? " + file))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderMenu displays the action menu
func (r *Renderer) renderMenu(m models.Model) string {
	menuItems := m.GetMenuItems()
	var b strings.Builder

	b.WriteString(normalStyle.Render(models.LabelActions))
	b.WriteString("\n")

	for i, item := range menuItems {
		if i == m.Selected {
			b.WriteString(selectedStyle.Render("▶ " + item))
		} else {
			b.WriteString(normalStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(normalStyle.Render(models.HelpMain))
	b.WriteString("\n")
	b.WriteString(normalStyle.Render(models.HelpHotkeys))

	return b.String()
}

// renderHistory displays the checkpoint history
func (r *Renderer) renderHistory(m models.Model) string {
	var b strings.Builder

	b.WriteString(normalStyle.Render(models.LabelHistory))
	b.WriteString("\n\n")

	if len(m.Checkpoints) == 0 {
		b.WriteString(normalStyle.Render(models.TextNoCheckpoints))
		b.WriteString("\n\n")
	} else {
		for i, checkpoint := range m.Checkpoints {
			prefix := "  "
			if i == m.HistorySelected {
				prefix = "▶ "
			}

			indicator := ""
			if checkpoint.IsCurrent {
				indicator = models.TextCurrent
			}

			line := fmt.Sprintf("%s%s %.7s - %s%s",
				prefix,
				checkpoint.Date.Format("2006-01-02 15:04"),
				checkpoint.Hash,
				checkpoint.Message,
				indicator,
			)

			if i == m.HistorySelected {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(normalStyle.Render(line))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(normalStyle.Render(models.HelpHistory))

	return b.String()
}
