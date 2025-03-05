package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/MsSabo/todo-app"
	"github.com/MsSabo/todo-app/pkg/ads"
	"github.com/jmoiron/sqlx"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

type date struct {
	year  int
	month int
	dat   int
}

func getDate(st time.Time) date {
	y, m, d := st.Date()
	return date{y, int(m), d}
}

func (s *AuthPostgres) CreateUser(user todo.User) (int, error) {
	var id int

	query := fmt.Sprintf("INSERT INTO %s (name, username, password_hash) values ($1, $2, $3) RETURNING id", userTable)

	row := s.db.QueryRow(query, user.Name, user.Username, user.Password)

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	ads.IncrementUserAction(id)
	return id, nil
}

func (s *AuthPostgres) GetUser(username, password string) (todo.User, error) {
	var user todo.User
	query := fmt.Sprintf("SELECT id FROM %s WHERE username=$1 AND password_hash=$2", userTable)
	err := s.db.Get(&user, query, username, password)

	return user, err
}

func (s *AuthPostgres) UpdateLastIn(id int) error {

	var last_cmd time.Time
	query := fmt.Sprintf("SELECT last_cmd FROM %s WHERE id=$1", userTable)
	err := s.db.QueryRow(query, id).Scan(&last_cmd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Update last cmd id = ", id)
	currentTime := time.Now()
	query = fmt.Sprintf("UPDATE %s SET last_cmd = $1 WHERE id=$2", userTable)
	_, err = s.db.Exec(query, currentTime, id)

	if getDate(last_cmd) != getDate(currentTime) {
		ads.IncrementUserAction(id)
	}
	return err
}
