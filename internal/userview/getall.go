package userview

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"com.user.com/user/internal/core"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type GetAllUsersEndpoint struct {
	userGetter UserGetter
	validator  *validator.Validate
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Nickname  string    `json:"nickname"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetAllUsersResponse struct {
	Users        []UserResponse `json:"users"`
	PreviousPage string         `json:"previous_page,omitempty"`
	NextPage     string         `json:"next_page,omitempty"`
	Total        int            `json:"total"`
}

type UserGetter interface {
	GetAllUsers(ctx context.Context, filter core.UserFilter) (users []*core.User, previousPage, nextPage string, total int, err error)
}

func NewGetAllUsersEndpoint(userGetter UserGetter) *GetAllUsersEndpoint {
	return &GetAllUsersEndpoint{
		userGetter: userGetter,
		validator:  validator.New(),
	}
}

func (c *GetAllUsersEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()
	logrus.WithContext(ctx).
		WithField("Endoint", "userview.GetAllUsersEndpoint").
		Debug("request started")

	previousPage := r.URL.Query().Get("previous_page")
	nextPage := r.URL.Query().Get("next_page")
	nickname := r.URL.Query().Get("nickname")
	country := r.URL.Query().Get("country")
	firstName := r.URL.Query().Get("first_name")
	lastName := r.URL.Query().Get("last_name")
	filter := core.UserFilter{
		NextPage:     nextPage,
		PreviousPage: previousPage,
		Nickname:     nickname,
		Country:      country,
		FirstName:    firstName,
		LastName:     lastName,
	}
	limit := r.URL.Query().Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid 'limit' query param: %v", limit), http.StatusBadRequest)
			return
		}
		filter.Limit = l
	}

	users, previousPage, nextPage, total, err := c.userGetter.GetAllUsers(ctx, filter)
	if err != nil {
		logrus.WithContext(ctx).
			WithError(err).
			Error("error while getting users")
		http.Error(w, fmt.Sprintf("failed to get users: %v", err), http.StatusInternalServerError)
		return
	}
	sliceOfUsers := make([]UserResponse, 0, len(users))
	for u := range users {
		sliceOfUsers = append(sliceOfUsers, UserResponse{
			ID:        users[u].ID,
			FirstName: users[u].FirstName,
			LastName:  users[u].LastName,
			Nickname:  users[u].Nickname,
			Email:     users[u].Email,
			Password:  users[u].Password,
			CreatedAt: users[u].CreatedAt,
			UpdatedAt: users[u].UpdatedAt,
		})
	}

	response := GetAllUsersResponse{
		Users:        sliceOfUsers,
		PreviousPage: previousPage,
		NextPage:     nextPage,
		Total:        total,
	}

	respondJSON(ctx, w, &response)
	logrus.WithContext(ctx).
		WithField("Endoint", "userview.GetAllUsersEndpoint").
		Debug("request completed")

}

func respondJSON(ctx context.Context, w http.ResponseWriter, resp interface{}) {
	jsonBody, err := json.Marshal(&resp)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed serializing response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonBody)
}
