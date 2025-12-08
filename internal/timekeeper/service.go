package timekeeper

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"time-machine/internal/models"
)

// Service provides git operations
type Service struct{}

// NewService creates a new git service
func NewService() *Service {
	return &Service{}
}

// LoadStatus loads the current git repository status
func (s *Service) LoadStatus() tea.Msg {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Open git repository
	repo, err := git.PlainOpen(pwd)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return models.GitNotInitializedMsg{
				Message: "Машина времени не запущена в этой папке",
			}
		}
		return models.ErrMsg{Error: err}
	}

	// Get worktree status
	worktree, err := repo.Worktree()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	status, err := worktree.Status()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Get current branch
	ref, err := repo.Head()
	if err != nil {
		// Handle case where there are no commits yet
		if err == plumbing.ErrReferenceNotFound {
			// Repository is initialized but has no commits
			gitStatus := &models.GitStatus{
				Branch:     "master", // Default branch name
				IsClean:    status.IsClean(),
				LastCommit: "Нет моментов",
			}
			return gitStatus
		}
		return models.ErrMsg{Error: err}
	}

	branchName := ref.Name().Short()
	if ref.Name().IsBranch() {
		branchName = string(ref.Name().Short())
	}

	// Get last commit info
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Build status object
	gitStatus := &models.GitStatus{
		Branch:     branchName,
		IsClean:    status.IsClean(),
		LastCommit: fmt.Sprintf("%s %.7s", commit.Message, commit.Hash.String()[:7]),
	}

	// Categorize files
	for file, entry := range status {
		switch entry.Worktree {
		case git.Modified:
			gitStatus.Modified = append(gitStatus.Modified, file)
		case git.Added:
			gitStatus.Staged = append(gitStatus.Staged, file)
		case git.Untracked:
			gitStatus.Untracked = append(gitStatus.Untracked, file)
		}
	}

	return gitStatus
}

// CreateCheckpoint creates a new checkpoint with the given description
func (s *Service) CreateCheckpoint(description string) tea.Msg {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Open git repository
	repo, err := git.PlainOpen(pwd)
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Add all changes
	_, err = worktree.Add(".")
	if err != nil {
		return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToAddFiles, err)}
	}

	// Create commit with custom message
	commit, err := worktree.Commit(description, &git.CommitOptions{
		Author: &object.Signature{
			Name:  models.CheckpointAuthorName,
			Email: models.CheckpointAuthorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToCreateCheckpoint, err)}
	}

	return models.CheckpointCreatedMsg{
		Success: true,
		Message: fmt.Sprintf("Момент зафиксирован: %.7s", commit.String()),
	}
}

// LoadCheckpoints loads the commit history
func (s *Service) LoadCheckpoints() tea.Msg {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Open git repository
	repo, err := git.PlainOpen(pwd)
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Get current HEAD
	head, err := repo.Head()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Get commit iterator
	commitIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return models.ErrMsg{Error: err}
	}
	defer commitIter.Close()

	var checkpoints []models.Checkpoint
	currentHash := head.Hash().String()

	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Show all commits without filtering
		checkpoint := models.Checkpoint{
			Hash:      commit.Hash.String(),
			Message:   commit.Message,
			Author:    commit.Author.Name,
			Date:      commit.Author.When,
			IsCurrent: commit.Hash.String() == currentHash,
		}
		checkpoints = append(checkpoints, checkpoint)
		return nil
	})

	if err != nil {
		return models.ErrMsg{Error: err}
	}

	return models.CheckpointsLoadedMsg{
		Checkpoints: checkpoints,
	}
}

// RollbackToCheckpoint rolls back to a specific checkpoint
func (s *Service) RollbackToCheckpoint(hash string) tea.Msg {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Open git repository
	repo, err := git.PlainOpen(pwd)
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Parse hash
	commitHash := plumbing.NewHash(hash)

	// Reset to the checkpoint
	err = worktree.Reset(&git.ResetOptions{
		Commit: commitHash,
		Mode:   git.HardReset,
	})
	if err != nil {
		return models.RollbackMsg{
			Success: false,
			Message: fmt.Sprintf("Не удалось перемотать: %v", err),
		}
	}

	return models.RollbackMsg{
		Success: true,
		Message: fmt.Sprintf("Успешно перемотали к моменту: %.7s", hash),
	}
}

// SyncWithRemote performs pull and push operations with simple conflict handling
func (s *Service) SyncWithRemote() tea.Msg {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Open git repository
	repo, err := git.PlainOpen(pwd)
	if err != nil {
		return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToOpenRepo, err)}
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToGetWorktree, err)}
	}

	// Get remote
	remote, err := repo.Remote("origin")
	if err != nil {
		// Return a user-friendly message instead of an error
		return models.SyncMsg{
			Success: false,
			Message: models.ErrNoRemote,
			Pulled:  false,
			Pushed:  false,
		}
	}

	syncMsg := models.SyncMsg{Success: true}

	// First, try to pull from remote
	fmt.Println("Pulling from remote...")
	pullErr := worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
	})

	if pullErr != nil {
		if pullErr == git.NoErrAlreadyUpToDate {
			syncMsg.Message = models.ErrAlreadyUpToDate
			syncMsg.Pulled = false
		} else {
			// Handle conflicts by forcing our changes (simple approach for vibecoders)
			fmt.Println("Conflicts detected, forcing local changes...")

			// Add all changes and commit if there are any
			status, err := worktree.Status()
			if err != nil {
				return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToGetStatus, err)}
			}

			if !status.IsClean() {
				// Add all changes
				_, err = worktree.Add(".")
				if err != nil {
					return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToAddChanges, err)}
				}

				// Create a conflict resolution commit
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				commitMsg := fmt.Sprintf("Auto-resolve conflicts: %s", timestamp)

				_, err = worktree.Commit(commitMsg, &git.CommitOptions{
					Author: &object.Signature{
						Name:  models.ConflictAuthorName,
						Email: models.ConflictAuthorEmail,
						When:  time.Now(),
					},
				})
				if err != nil {
					return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToCommit, err)}
				}
			}

			syncMsg.Conflict = true
			syncMsg.Message = models.ErrConflictsDetected
		}
	} else {
		syncMsg.Pulled = true
		syncMsg.Message = models.ErrPullSuccess
	}

	// Then, push to remote
	fmt.Println("Pushing to remote...")
	pushErr := remote.Push(&git.PushOptions{
		RemoteName: "origin",
	})

	if pushErr != nil {
		if pushErr == git.NoErrAlreadyUpToDate {
			if syncMsg.Message == models.ErrAlreadyUpToDate {
				syncMsg.Message = models.ErrAlreadyUpToDate
			} else {
				syncMsg.Message += ", already up to date on push"
			}
			syncMsg.Pushed = false
		} else {
			// Try force push for simplicity (acceptable for vibecoders)
			fmt.Println("Normal push failed, trying force push...")
			forceErr := remote.Push(&git.PushOptions{
				RemoteName: "origin",
				Force:      true,
			})

			if forceErr != nil {
				return models.ErrMsg{Error: fmt.Errorf("%s: %w", models.ErrFailedToPush, forceErr)}
			}

			syncMsg.Pushed = true
			if syncMsg.Message == models.ErrAlreadyUpToDate {
				syncMsg.Message = models.ErrForcePushSuccess
			} else {
				syncMsg.Message += ", force pushed successfully"
			}
		}
	} else {
		syncMsg.Pushed = true
		if syncMsg.Message == models.ErrAlreadyUpToDate {
			syncMsg.Message = models.ErrPushSuccess
		} else {
			syncMsg.Message += ", pushed successfully"
		}
	}

	return syncMsg
}

// InitGit initializes a new git repository
func (s *Service) InitGit() tea.Msg {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return models.ErrMsg{Error: err}
	}

	// Initialize git repository
	_, err = git.PlainInit(pwd, false)
	if err != nil {
		return models.ErrMsg{Error: fmt.Errorf("не удалось запустить машину времени: %w", err)}
	}

	return models.GitInitializedMsg{}
}
