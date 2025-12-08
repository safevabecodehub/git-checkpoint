package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"

	"time-machine/internal/models"
	"time-machine/internal/timekeeper"
	"time-machine/internal/ui"
)

func main() {
	// Initialize services
	gitService := timekeeper.NewService()
	renderer := ui.NewRenderer()

	// Initialize model
	m := models.Model{
		Selected: 0,
	}

	// Enable debug logging if DEBUG environment variable is set
	if len(os.Getenv("DEBUG")) > 0 {
		if f, err := tea.LogToFile("debug.log", "debug"); err == nil {
			defer f.Close()
		}
	}

	// Create and run the program
	p := tea.NewProgram(
		NewApp(gitService, renderer, m),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

// App represents the Bubble Tea application
type App struct {
	gitService *timekeeper.Service
	renderer   *ui.Renderer
	model      models.Model
}

// NewApp creates a new application instance
func NewApp(gitService *timekeeper.Service, renderer *ui.Renderer, model models.Model) *App {
	return &App{
		gitService: gitService,
		renderer:   renderer,
		model:      model,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return a.gitService.LoadStatus
}

// Update handles user input and updates the model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *models.GitStatus:
		a.model.Status = msg
		a.model.Loading = false
		return a, nil

	case models.ErrMsg:
		a.model.Err = msg.Error
		a.model.Loading = false
		return a, nil

	case models.StatusMsg:
		a.model.Loading = false
		return a, nil

	case models.GitInitializedMsg:
		a.model.GitNotInitialized = false
		a.model.Err = nil
		a.model.Loading = false
		return a, a.gitService.LoadStatus

	case models.CheckpointCreatedMsg:
		a.model.Loading = false
		if msg.Success {
			return a, a.gitService.LoadStatus
		}
		return a, nil

	case models.CheckpointsLoadedMsg:
		a.model.Checkpoints = msg.Checkpoints
		a.model.HistoryMode = true
		a.model.HistorySelected = 0
		a.model.Loading = false
		return a, nil

	case models.RollbackMsg:
		a.model.Loading = false
		a.model.HistoryMode = false
		if msg.Success {
			return a, a.gitService.LoadStatus
		}
		return a, nil

	case models.DescriptionModeMsg:
		a.model.Loading = false
		a.model.DescriptionMode = true
		a.model.DescriptionInput = ""
		a.model.Suggestions = msg.Suggestions
		return a, nil

	case models.GitNotInitializedMsg:
		a.model.GitNotInitialized = true
		a.model.Err = fmt.Errorf(msg.Message)
		a.model.Loading = false
		return a, nil

	case models.SyncMsg:
		a.model.Loading = false
		if msg.Success {
			return a, a.gitService.LoadStatus
		}
		// Store sync error message to display
		a.model.SyncMessage = msg.Message
		a.model.ShowSyncMessage = true
		return a, nil

	case tea.KeyMsg:
		return a.handleKeyMsg(msg)
	}

	return a, nil
}

// View renders the UI
func (a *App) View() string {
	return a.renderer.View(a.model)
}

// handleKeyMsg handles keyboard input
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.model.Loading {
		return a, nil
	}

	// Clear sync message when user presses any key
	if a.model.ShowSyncMessage {
		a.model.ShowSyncMessage = false
		a.model.SyncMessage = ""
	}

	if a.model.DescriptionMode {
		return a.handleDescriptionInput(msg)
	}

	if a.model.HistoryMode {
		return a.handleHistoryInput(msg)
	}

	// Handle Escape key using Type for better reliability
	switch msg.Type {
	case tea.KeyEscape:
		a.model.Quitting = true
		return a, tea.Quit
	}

	switch msg.String() {
	case "ctrl+c", "q":
		a.model.Quitting = true
		return a, tea.Quit

	case "esc", "escape":
		// Fallback for terminals where Type detection doesn't work
		a.model.Quitting = true
		return a, tea.Quit

	case "up", "k":
		if a.model.Selected > 0 {
			a.model.Selected--
		}

	case "down", "j":
		if a.model.Selected < len(models.GetMenuItems())-1 {
			a.model.Selected++
		}

	case "enter", " ":
		// Handle menu selection
		return a, a.handleMenuSelection()

	// Hotkeys for quick actions
	case "c":
		// Create checkpoint shortcut
		a.model.Selected = 0
		return a, a.handleMenuSelection()

	case "h":
		// View history shortcut
		a.model.Selected = 1
		return a, a.handleMenuSelection()

	case "r":
		// Rollback shortcut
		a.model.Selected = 2
		return a, a.handleMenuSelection()

	case "s":
		// Sync shortcut
		a.model.Selected = 3
		return a, a.handleMenuSelection()
	}

	return a, nil
}

// handleDescriptionInput handles input when in description mode
func (a *App) handleDescriptionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle number keys for quick selection first
	if len(msg.Runes) == 1 {
		r := msg.Runes[0]
		if r >= '1' && r <= '9' {
			index := int(r - '1')
			if index < len(a.model.Suggestions) {
				a.model.DescriptionInput = a.model.Suggestions[index]
				return a, nil
			}
		}
	}

	switch msg.Type {
	case tea.KeyEscape:
		// Exit description mode
		a.model.DescriptionMode = false
		a.model.DescriptionInput = ""
		return a, nil

	case tea.KeyEnter:
		// Create checkpoint with description
		description := a.model.DescriptionInput
		if description == "" {
			// Use default if empty
			description = "Сейв без описания"
		}
		a.model.DescriptionMode = false
		a.model.Loading = true
		a.model.LoadingText = "Сейвлю вайб..."
		return a, func() tea.Msg {
			return a.gitService.CreateCheckpoint(description)
		}

	case tea.KeyBackspace:
		if len(a.model.DescriptionInput) > 0 {
			a.model.DescriptionInput = a.model.DescriptionInput[:len(a.model.DescriptionInput)-1]
		}
		return a, nil

	case tea.KeyRunes:
		// Add typed characters (but not numbers since we handled them above)
		r := msg.Runes[0]
		if r < '1' || r > '9' {
			a.model.DescriptionInput += string(msg.Runes)
		}
		return a, nil
	}

	switch msg.String() {
	case "ctrl+c":
		a.model.Quitting = true
		return a, tea.Quit
	}

	return a, nil
}

// handleHistoryInput handles input when in history mode
func (a *App) handleHistoryInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Escape key using Type for better reliability
	switch msg.Type {
	case tea.KeyEscape:
		// Escape goes back to main menu
		a.model.HistoryMode = false
		return a, nil

	case tea.KeyBackspace:
		// Backspace also goes back to main menu (more intuitive)
		a.model.HistoryMode = false
		return a, nil
	}

	switch msg.String() {
	case "ctrl+c":
		a.model.Quitting = true
		return a, tea.Quit

	case "q":
		// 'q' now quits from history mode too for consistency
		a.model.Quitting = true
		return a, tea.Quit

	case "esc", "escape":
		// Fallback for terminals where Type detection doesn't work
		a.model.HistoryMode = false
		return a, nil

	case "up", "k":
		if a.model.HistorySelected > 0 {
			a.model.HistorySelected--
		}

	case "down", "j":
		if a.model.HistorySelected < len(a.model.Checkpoints)-1 {
			a.model.HistorySelected++
		}

	case "enter", " ":
		if a.model.HistorySelected < len(a.model.Checkpoints) {
			checkpoint := a.model.Checkpoints[a.model.HistorySelected]
			a.model.Loading = true
			a.model.LoadingText = "Возвращаю старый вайб..."
			return a, func() tea.Msg {
				return a.gitService.RollbackToCheckpoint(checkpoint.Hash)
			}
		}
	}

	return a, nil
}

// handleMenuSelection processes the selected menu item
func (a *App) handleMenuSelection() tea.Cmd {
	menuItems := a.model.GetMenuItems()
	if a.model.Selected >= len(menuItems) {
		return nil
	}

	selectedItem := menuItems[a.model.Selected]

	switch selectedItem {
	case models.MenuInitGit:
		a.model.Loading = true
		a.model.LoadingText = "Настраиваю пространство..."
		return a.gitService.InitGit

	case models.MenuCreateCheckpoint:
		// Enter description mode via async message (like history)
		a.model.Loading = true
		a.model.LoadingText = "Ловлю вдохновение..."
		return func() tea.Msg {
			return models.DescriptionModeMsg{
				Suggestions: models.DefaultSuggestions,
			}
		}

	case models.MenuViewHistory:
		a.model.Loading = true
		a.model.LoadingText = "Вспоминаем былое..."
		return a.gitService.LoadCheckpoints

	case models.MenuRollback:
		a.model.Loading = true
		a.model.LoadingText = "Вспоминаем былое..."
		return a.gitService.LoadCheckpoints

	case models.MenuSync:
		a.model.Loading = true
		a.model.LoadingText = "Синхронизирую потоки..."
		return a.gitService.SyncWithRemote
	}

	return nil
}
