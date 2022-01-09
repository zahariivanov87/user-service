package userview

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type DeleteUserEndpoint struct {
	userRemover UserRemover
}

type UserRemover interface {
	RemoveUser(ctx context.Context, id uuid.UUID) error
}

func NewDeleteUserEndpoint(userRemover UserRemover) *DeleteUserEndpoint {
	return &DeleteUserEndpoint{
		userRemover: userRemover,
	}
}

func (d *DeleteUserEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()
	logrus.WithContext(ctx).
		WithField("Endoint", "userview.DeleteUserEndpoint").
		Debug("request started")

	params := mux.Vars(r)
	userID := params["userID"]
	id, err := uuid.Parse(userID)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Error("error parsing user's payload")
		http.Error(w, fmt.Sprintf("invalid user id: %v", userID), http.StatusBadRequest)
		return
	}
	err = d.userRemover.RemoveUser(ctx, id)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Error("error while deleting user")
		http.Error(w, fmt.Sprintf("error while deleting user: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.WithContext(ctx).
		WithField("Endoint", "userview.DeleteUserEndpoint").
		Debug("request completed")

}
