package userview

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"com.user.com/user/internal/core"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type UpdateUserEndpoint struct {
	userModifier UserModifier
	validator    *validator.Validate
}

type UserModifier interface {
	ModifyUser(ctx context.Context, user core.User) error
}

type UpdateUserParams struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	FirstName string    `json:"first_name" validate:"required"`
	LastName  string    `json:"last_name" validate:"required"`
	Nickname  string    `json:"nickname" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Country   string    `json:"country" validate:"required"`
}

func NewUpdateUserEndpoint(userManager UserModifier) *UpdateUserEndpoint {
	return &UpdateUserEndpoint{
		userModifier: userManager,
		validator:    validator.New(),
	}
}

func (u *UpdateUserEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()
	logrus.WithContext(ctx).
		WithField("Endoint", "userview.GetAllUsersEndpoint").
		Debug("request started")

	params := mux.Vars(r)
	userID := params["userID"]
	id, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid user id: %v", userID), http.StatusBadRequest)
		return
	}

	var updateUserParams UpdateUserParams
	if !tryReadingBody(ctx, w, r, &updateUserParams, u.validator) {
		return
	}

	//Translate outer layer (view) user into internal (core) user
	user := core.User{
		ID:        id,
		FirstName: updateUserParams.FirstName,
		LastName:  updateUserParams.LastName,
		Nickname:  updateUserParams.Nickname,
		Password:  updateUserParams.Password,
		Email:     updateUserParams.Email,
		Country:   updateUserParams.Country,
	}

	err = u.userModifier.ModifyUser(ctx, user)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Error("error while modifying user")
		http.Error(w, fmt.Sprintf("error while modifying user: %v", err), http.StatusInternalServerError)
		return
	}

	logrus.WithContext(ctx).
		WithField("Endoint", "userview.GetAllUsersEndpoint").
		Debug("request started")

}
