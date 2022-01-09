package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"com.user.com/user/internal/core"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

//go:generate ~/go/bin/counterfeiter  . UserStore

type UserStore interface {
	SaveUser(ctx context.Context, user core.User) error
	UpdateUser(ctx context.Context, user core.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetAllUsers(ctx context.Context, filter core.UserFilter) (users []*core.User, previousPage, nextPage string, total int, err error)
}

type Notifier interface {
	NotifySubscriber(ctx context.Context, msg string) error
}

type Manager struct {
	userStore    UserStore
	notifier     Notifier
	shouldNotify bool
}

func NewManager(userStore UserStore, notifier Notifier, shouldNotify bool) *Manager {
	return &Manager{
		userStore:    userStore,
		notifier:     notifier,
		shouldNotify: shouldNotify,
	}
}

func (m *Manager) CreateUser(ctx context.Context, user core.User) error {
	user.ID = uuid.New()
	if !m.isEmailValid(user.Email) {
		return errors.New("invalid email")
	}
	err := m.userStore.SaveUser(ctx, user)
	if err != nil {
		return err
	}
	if m.shouldNotify {
		err = m.notifier.NotifySubscriber(ctx, fmt.Sprintf("User has been created: %v", user))
		if err != nil {
			logrus.WithContext(ctx).
				WithError(err).
				Error("user.manager: error while notifying subscribers for user creation")
		}
	}
	return nil
}

func (m *Manager) ModifyUser(ctx context.Context, user core.User) error {
	if !m.isEmailValid(user.Email) {
		return errors.New("invalid email")
	}
	err := m.userStore.UpdateUser(ctx, user)
	if err != nil {
		return err
	}
	if m.shouldNotify {
		err = m.notifier.NotifySubscriber(ctx, fmt.Sprintf("User has been updated: %v", user))
		if err != nil {
			logrus.WithContext(ctx).
				WithError(err).
				Error("user.manager: error while notifying subscribers for user update")
		}
	}
	return nil
}

func (m *Manager) RemoveUser(ctx context.Context, id uuid.UUID) error {
	err := m.userStore.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	if m.shouldNotify {
		err = m.notifier.NotifySubscriber(ctx, fmt.Sprintf("User has been deleted: %s", id))
		if err != nil {
			logrus.WithContext(ctx).
				WithError(err).
				Error("user.manager: error while notifying subscribers for user deletion")
		}
	}
	return nil
}

func (m *Manager) GetAllUsers(ctx context.Context, filter core.UserFilter) (users []*core.User, previousPage, nextPage string, total int, err error) {
	if filter.PreviousPage != "" && filter.NextPage != "" {
		return nil, "", "", 0, errors.New("either next or previous page should be provided")
	}
	return m.userStore.GetAllUsers(ctx, filter)
}

func (m *Manager) isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}
