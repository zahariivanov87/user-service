package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"com.user.com/user/internal/core"

	"github.com/google/uuid"
)

const (
	storeUserStmt  = `INSERT INTO users (id, first_name, last_name, nickname, password, email, country, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())`
	updateUserStms = `UPDATE users SET first_name=$1, last_name=$2, nickanme=$3, password=$4, email=$5, country=$6 last_updated=now() WHERE id=&7 and updated_at<&8`
	deleteUserStmt = `DELETE FROM users WHERE id=$1`

	DEFAULT_LIMIT = 100
)

// Store - represents abstraction over db
type Store struct {
	db sql.DB
}

func NewStore(db sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// SaveUser - stores user entity in db
func (s *Store) SaveUser(ctx context.Context, user core.User) error {
	_, err := s.db.ExecContext(ctx,
		storeUserStmt,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Password,
		user.Email,
		user.Country,
	)
	return err
}

// UpdateUser - updates user entity in db
// updated_at<&8 condition takes care of concurrent modifications.
func (s *Store) UpdateUser(ctx context.Context, user core.User) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET first_name=$1, last_name=$2, nickanme=$3, password=$4, email=$5, country=$6 last_updated=now() WHERE id=&7 and updated_at<&8",
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Password,
		user.Email,
		user.Country,
		user.ID,
		user.UpdatedAt,
	)
	return err
}

// DeleteUser - deletes user from db
func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		deleteUserStmt,
		id,
	)
	return err
}

func (s *Store) GetAllUsers(ctx context.Context, filter core.UserFilter) (users []*core.User, previousPage, nextPage string, total int, err error) {
	var (
		createdAt        string
		id               string
		offset           int
		whereOrAndClause string
	)
	getAllUsersBaseStmt := `SELECT id, first_name, last_name, nickname, password, email, country, created_at, updated_at FROM users`
	args, nextPlaceholder := s.buildArgumentsAndExtendQueryFromFilter(filter, getAllUsersBaseStmt)
	whereOrAndClause = "AND"
	if len(args) == 0 {
		whereOrAndClause = "WHERE"
	}
	limit := filter.Limit
	if limit == 0 {
		limit = DEFAULT_LIMIT
	}
	// Newly created users first
	sortOrder := "DESC"
	if filter.NextPage != "" {
		createdAt, id, offset, err = s.decodeCursor(filter.NextPage)
		if err != nil {
			return nil, "", "", 0, err
		}
		getAllUsersBaseStmt += fmt.Sprintf(" %s (created_at, id) < ($%d, $%d)", whereOrAndClause, nextPlaceholder, nextPlaceholder+1)
		args = append(args, createdAt, id)
		nextPlaceholder += 2
		offset += limit
	} else if filter.PreviousPage != "" {
		createdAt, id, offset, err = s.decodeCursor(filter.PreviousPage)
		if err != nil {
			return nil, "", "", 0, err
		}
		// if offset - limit = 0, then reload first page again. In that case if new users have appeared
		// they will be part of new pagination starting from the first page.
		if offset-limit != 0 {
			getAllUsersBaseStmt += fmt.Sprintf(" %s (created_at, id) > ($%d, $%d)", whereOrAndClause, nextPlaceholder, nextPlaceholder+1)
			args = append(args, createdAt, id)
			nextPlaceholder += 2
			// Reverse the set in order to move backwards
			sortOrder = "ASC"
		}
		offset -= limit
	} else {
		offset = 0
	}

	getAllUsersBaseStmt += fmt.Sprintf(" ORDER BY created_at %s, id %s", sortOrder, sortOrder)
	getAllUsersBaseStmt += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := s.db.QueryContext(ctx, getAllUsersBaseStmt, args...)
	defer func() {
		_ = rows.Close()
	}()
	if err != nil {
		return nil, "", "", 0, err
	}

	results := make([]*core.User, 0)
	for rows.Next() {
		var u core.User
		userRowErr := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.Nickname,
			&u.Password,
			&u.Email,
			&u.Country,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if userRowErr != nil {
			return nil, "", "", 0, userRowErr
		}
		results = append(results, &u)
	}

	total, err = s.totalCount(ctx, filter)
	if err != nil {
		return nil, "", "", 0, err
	}

	previousPage = s.setPreviousPage(results, offset, limit)
	nextPage = s.setNextPage(results, offset, total, limit)

	return results, previousPage, nextPage, total, nil
}

// returns slice of arguments to be used in the query as well as the last placeholder.
func (s *Store) buildArgumentsAndExtendQueryFromFilter(filter core.UserFilter, getUsersBaseStmt string) ([]interface{}, int) {
	var args []interface{}
	nextPlaceholder := 0
	whereOrAndStmt := "WHERE"
	if filter.Country != "" {
		getUsersBaseStmt += fmt.Sprintf(" %s country=$%d", whereOrAndStmt, nextPlaceholder)
		args = append(args, filter.Country)
		whereOrAndStmt = "AND"
		nextPlaceholder++
	}
	if filter.FirstName != "" {
		getUsersBaseStmt += fmt.Sprintf(" %s first_name=$%d", whereOrAndStmt, nextPlaceholder)
		args = append(args, filter.FirstName)
		whereOrAndStmt = "AND"
		nextPlaceholder++
	}
	if filter.LastName != "" {
		getUsersBaseStmt += fmt.Sprintf(" %s last_name=$%d", whereOrAndStmt, nextPlaceholder)
		args = append(args, filter.LastName)
		whereOrAndStmt = "AND"
		nextPlaceholder++
	}
	if filter.Nickname != "" {
		getUsersBaseStmt += fmt.Sprintf(" %s nickname=$%d", whereOrAndStmt, nextPlaceholder)
		args = append(args, filter.Nickname)
	}
	return args, nextPlaceholder + 1
}

func (s *Store) setPreviousPage(results []*core.User, offset, limit int) string {
	// If offset - limit = 0, it means that we are on the first page, hence it makes no sense to set previous_page
	if offset-limit >= 0 && len(results) > 0 {
		return s.encodeCursor(results[len(results)-1].CreatedAt.UTC(), results[len(results)-1].ID.String(), offset)
	}
	return ""
}

func (s *Store) setNextPage(results []*core.User, offset, total, limit int) string {
	// Only in this case it makes sense to put next_page as part of response. Otherwise it means that we are on
	// the last page, hence next_page is obsolete.
	if total-offset > 0 && len(results) >= limit && (total-offset != limit) {
		return s.encodeCursor(results[len(results)-1].CreatedAt.UTC(), results[len(results)-1].ID.String(), offset)
	}
	return ""
}

func (s *Store) decodeCursor(encodedCursor string) (createdAt, id string, offset int, err error) {
	byt, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return
	}

	arrStr := strings.Split(string(byt), ",")
	if len(arrStr) != 3 {
		err = errors.New("cursor is invalid")
		return
	}

	createdAt = arrStr[0]
	id = arrStr[1]
	offset, err = strconv.Atoi(arrStr[2])
	if err != nil {
		err = errors.New("cursor is invalid")
		return
	}

	return createdAt, id, offset, nil
}

func (s *Store) encodeCursor(createdAt time.Time, id string, offset int) string {
	createdAtAsStr := fmt.Sprintf("%d-%d-%d %d:%d:%d", createdAt.Year(), int(createdAt.Month()),
		createdAt.Day(), createdAt.Hour(), createdAt.Minute(), createdAt.Second())
	key := fmt.Sprintf("%s,%s,%d", createdAtAsStr, id, offset)
	return base64.StdEncoding.EncodeToString([]byte(key))
}

func (s *Store) totalCount(ctx context.Context, filter core.UserFilter) (int, error) {
	getAllUsersBaseStmt := `SELECT count(*) FROM users`
	args, _ := s.buildArgumentsAndExtendQueryFromFilter(filter, getAllUsersBaseStmt)
	rows, err := s.db.QueryContext(ctx, getAllUsersBaseStmt, args...)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var total int
	for rows.Next() {
		err = rows.Scan(&total)
		if err != nil {
			return 0, err
		}
	}
	if rows.Err() != nil {
		return 0, err
	}
	return total, nil
}

func (s *Store) InitDBTables(ctx context.Context) {

	path := filepath.Join("./migrations", "001_create_users_db.up.sql")

	c, ioErr := ioutil.ReadFile(path)
	if ioErr != nil {
		// handle error.
		fmt.Println("XXX error reading file")
		panic(ioErr)
	}
	sql := string(c)

	_, err := s.db.ExecContext(ctx, sql)
	if err != nil {
		fmt.Println("XXX error initiating users table")
		panic(err)
	}

}
