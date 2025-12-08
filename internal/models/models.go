package models

import (
	"time"

	"github.com/go-git/go-git/v5"
)

// Model represents the application state
type Model struct {
	Repo              *git.Repository
	Status            *GitStatus
	Err               error
	Selected          int
	Quitting          bool
	Checkpoints       []Checkpoint
	HistoryMode       bool
	HistorySelected   int
	Loading           bool
	LoadingText       string
	SyncMessage       string
	ShowSyncMessage   bool
	GitNotInitialized bool
	// Description input mode
	DescriptionMode  bool
	DescriptionInput string
	Suggestions      []string
}

// GitStatus represents git repository status
type GitStatus struct {
	Branch     string
	Staged     []string
	Modified   []string
	Untracked  []string
	Ahead      int
	Behind     int
	IsClean    bool
	LastCommit string
}

// Checkpoint represents a git commit checkpoint
type Checkpoint struct {
	Hash      string
	Message   string
	Author    string
	Date      time.Time
	IsCurrent bool
}

// Message types for Bubble Tea
type (
	StatusMsg struct {
		Text string
	}

	CheckpointCreatedMsg struct {
		Success bool
		Message string
	}

	CheckpointsLoadedMsg struct {
		Checkpoints []Checkpoint
	}

	RollbackMsg struct {
		Success bool
		Message string
	}

	SyncMsg struct {
		Success  bool
		Message  string
		Pulled   bool
		Pushed   bool
		Conflict bool
	}

	DescriptionModeMsg struct {
		Suggestions []string
	}

	GitNotInitializedMsg struct {
		Message string
	}

	GitInitializedMsg struct{}
)

// ErrMsg wraps an error for Bubble Tea
type ErrMsg struct {
	Error error
}

// Menu items constants
const (
	MenuInitGit          = "Начать Vibe-сессию"
	MenuCreateCheckpoint = "Засейвить вайб (Save Vibe)"
	MenuViewHistory      = "История потока (Flow History)"
	MenuRollback         = "Вернуть прошлый вайб"
	MenuSync             = "Синкнуть с облаком"
)

// UI text constants
const (
	TitleMain         = " VibeGit Flow 🌊 "
	TitleDescription  = " VibeGit [Сейвим вайб] "
	PromptDescription = "Опиши этот момент потока:"
	PromptSuggestions = "💡 Или выбери муд:"
	HelpMain          = "↑↓ Навигация | Enter Выбрать | q Выход"
	HelpHotkeys       = "Хоткеи: [C] Сейв [H] История [R] Ресет [S] Синк"
	HelpDescription   = "[Enter Засейвить] [Esc Отмена] [1-9 Быстрый выбор]"
	HelpHistory       = "↑↓ Листать | Enter Вернуть этот вайб | Esc Назад"
	LabelActions      = "Что делаем:"
	LabelHistory      = "Твой флоу:"
	LabelBranch       = "Ветка:"
	LabelLastCommit   = "Последний сейв:"
	LabelStaged       = "Готово к сейву:"
	LabelModified     = "Изменилось:"
	LabelUntracked    = "Новое:"
	TextNoCheckpoints = "Вайбов пока нет, начинай творить"
	TextCurrent       = " (текущий вайб)"
	TextClean         = "✓ Ты в потоке. Всё чисто."
	TextDirty         = "⚡ Есть незасейвленный прогресс"
	TextLoading       = "В процессе: "
)

// Error messages
const (
	ErrFailedToAddFiles         = "не удалось добавить файлы"
	ErrFailedToCreateCheckpoint = "не удалось зафиксировать момент"
	ErrFailedToOpenRepo         = "не удалось открыть проект"
	ErrFailedToGetWorktree      = "не удалось получить рабочую папку"
	ErrFailedToGetStatus        = "не удалось получить статус"
	ErrFailedToGetHead          = "не удалось получить текущий момент"
	ErrFailedToCommit           = "не удалось сохранить решение конфликта"
	ErrFailedToAddChanges       = "не удалось добавить изменения"
	ErrFailedToPush             = "не удалось отправить копию"
	ErrNoRemote                 = "Удаленное хранилище не найдено. Это только локальная версия."
	ErrAlreadyUpToDate          = "Всё актуально"
	ErrConflictsDetected        = "Конфликты решены автоматически"
	ErrForcePushSuccess         = "Копия отправлена принудительно"
	ErrPushSuccess              = "Копия отправлена успешно"
	ErrPullSuccess              = "Копия получена успешно"
)

// Time machine author info
const (
	CheckpointAuthorName  = "Машина Времени"
	CheckpointAuthorEmail = "timemachine@local"
	ConflictAuthorName    = "Time Machine TUI"
	ConflictAuthorEmail   = "timemachine@local"
)

// Default description suggestions
var DefaultSuggestions = []string{
	"Поймал волну 🌊",
	"Фикс на лету 🐛",
	"Новая фича готова ✨",
	"Рефакторинг для души 🧹",
	"Эксперименты с кодом 🧪",
	"Просто сейв на всякий случай 🛡️",
	"Красиво сделал 🎨",
	"Оптимизация 🚀",
	"Тесты прошли ✅",
	"Вайб чек 🤙",
	"Прогресс неостановим 🔥",
	"Магия кода 🪄",
	"Дзен-код 🧘",
	"Ещё один шаг к релизу 🎯",
}

// GetMenuItems returns the list of menu items
func GetMenuItems() []string {
	return []string{
		MenuCreateCheckpoint,
		MenuViewHistory,
		MenuRollback,
		MenuSync,
	}
}

// GetMenuItems returns the list of menu items based on current state
func (m *Model) GetMenuItems() []string {
	if m.GitNotInitialized {
		return []string{MenuInitGit}
	}
	return []string{
		MenuCreateCheckpoint,
		MenuViewHistory,
		MenuRollback,
		MenuSync,
	}
}
