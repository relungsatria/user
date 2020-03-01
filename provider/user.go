package provider

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis"
	"log"
	"sync"
	"time"
	"user/config"
	"user/domain"
)

type userRepo struct {
	db    *sql.DB
	redis *redis.Client
	mutex sync.Mutex
}

type UserProvider interface {
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByUserID(ctx context.Context, userID int64) (domain.User, error)
	AddUser(ctx context.Context, user domain.User) error
	EditUser(ctx context.Context, user domain.User) error
	DeleteUser(ctx context.Context, userID int64) error
	GetSession(ctx context.Context, sessionID string) (domain.Session, error)
	SetSession(ctx context.Context, sessionID string, userID int64) error
	DeleteSession(ctx context.Context, sessionID string) error
}

func NewUserProvider() UserProvider {
	repo := &userRepo{}
	repo.getDB()

	return repo
}

func (u *userRepo) getDB() *sql.DB {
	if u.db == nil {
		u.mutex.Lock()
		u.db = initDB(config.GetConfig().Database[config.UserDB].Master)
		u.mutex.Unlock()
	}
	return u.db
}

func (u *userRepo) getRedis() *redis.Client {
	if u.redis == nil {
		u.mutex.Lock()
		u.redis = initRedis(config.GetConfig().Redis[config.UserRedis].Address)
		u.mutex.Unlock()
	}
	return u.redis
}

func (u *userRepo) GetUserByUserID(ctx context.Context, userID int64) (user domain.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second * 1)
	defer cancel()

	rows, err := queryTemplate(ctx, u.getDB(), `select user_id, username, email, password from user where user_id = ?`, userID)
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}

func (u *userRepo) GetUserByUsername(ctx context.Context, username string) (user domain.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second * 1)
	defer cancel()

	rows, err := queryTemplate(ctx, u.getDB(), `select user_id, username, email, password from user where username = ?`, username)
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}

func (u *userRepo) GetUserByEmail(ctx context.Context, email string) (user domain.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	rows, err := queryTemplate(ctx, u.getDB(), `select user_id, username, email, password from user where email = ?`, email)
	if err != nil {
		return
	}
	for rows.Next() {
		err = rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			log.Print(err)
			return
		}
	}
	return
}

func (u *userRepo) GetSession(ctx context.Context, sessionID string) (session domain.Session, err error) {
	ctx, cancel := context.WithTimeout(ctx, 300)
	defer cancel()

	result, err := u.getRedis().Get(sessionID).Result()
	if err != nil {
		log.Println(err)
		return
	}
	session.UserID = result
	session.SessionID = sessionID
	return
}

func (u *userRepo) SetSession(ctx context.Context, sessionID string, userID int64) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 300)
	defer cancel()

	err = u.getRedis().Set(sessionID, userID, time.Hour*24*2).Err()
	if err != nil {
		log.Println(err)
	}
	return
}

func (u *userRepo) DeleteSession(ctx context.Context, sessionID string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 300)
	defer cancel()

	err = u.getRedis().Del(sessionID).Err()
	if err != nil {
		log.Println(err)
	}
	return
}

func (u *userRepo) AddUser(ctx context.Context, user domain.User) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return mutationTemplate(ctx, u.getDB(), `insert into user (username, email, password) values (?, ?, ?)`, user.Username, user.Email, user.Password)
}

func (u *userRepo) EditUser(ctx context.Context, user domain.User) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return mutationTemplate(ctx, u.getDB(), `update user set username = ?, email = ? where user_id = ?`, user.Username, user.Email, user.UserID)
}

func (u *userRepo) DeleteUser(ctx context.Context, userID int64) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return mutationTemplate(ctx, u.getDB(), `delete from user where user_id = ?`, userID)
}