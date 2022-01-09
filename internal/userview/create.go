package userview

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"com.user.com/user/internal/core"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
)

type CreateUserEndpoint struct {
	userCreator UserCreator
	validator   *validator.Validate
}

type UserCreator interface {
	CreateUser(ctx context.Context, user core.User) error
}

type CreateUserParams struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Nickname  string `json:"nickname" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Email     string `json:"email" validate:"required"`
	Country   string `json:"country" validate:"required"`
}

func NewCreateUserEndpoint(userCreator UserCreator) *CreateUserEndpoint {
	return &CreateUserEndpoint{
		userCreator: userCreator,
		validator:   validator.New(),
	}
}

func (c *CreateUserEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()

	logrus.WithContext(ctx).
		WithField("Endoint", "userview.CreateUserEndpoint").
		Debug("request started")
	var createUserParams CreateUserParams
	if !tryReadingBody(ctx, w, r, &createUserParams, c.validator) {
		return
	}

	//Translate outer layer (view) user into internal (core) user
	user := core.User{
		FirstName: createUserParams.FirstName,
		LastName:  createUserParams.LastName,
		Nickname:  createUserParams.Nickname,
		Password:  createUserParams.Password,
		Email:     createUserParams.Email,
		Country:   createUserParams.Country,
	}

	err := c.userCreator.CreateUser(ctx, user)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Error("error while creating user")
		http.Error(w, fmt.Sprintf("error while creating user: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.WithContext(ctx).
		WithField("Endoint", "userview.CreateUserEndpoint").
		Debug("request completed")

}

func tryReadingBody(
	ctx context.Context, w http.ResponseWriter, r *http.Request,
	into interface{}, validate *validator.Validate,
) bool {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed parsing request body: %v", err), http.StatusBadRequest)
		return false
	}
	_ = r.Body.Close()

	err = json.Unmarshal(data, into)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to deserialize request body: %v", err), http.StatusBadRequest)
		return false
	}

	err = validate.Struct(into)
	if err != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return false
	}
	return true
}
