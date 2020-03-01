package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"user/config"
	"user/domain"
	"user/provider"
)

type UserUsecase interface {
	Register(ctx context.Context, user domain.User) error
	Login(ctx context.Context, login domain.User) (string, error)
	GetSession(ctx context.Context, sessionID string) (domain.Session, error)
	GetUser(ctx context.Context, sessionID int64) (domain.User, error)
	EditUser(ctx context.Context, editedUser domain.User) error
	DeleteUser(ctx context.Context, userID int64, sessionID string) error
}

type userUsecase struct {
	config       *config.Config
	userProvider provider.UserProvider
}

func NewUserUsecase() UserUsecase {
	return &userUsecase{
		config:       config.GetConfig(),
		userProvider: provider.NewUserProvider(),
	}
}

func (u *userUsecase) GetUser(ctx context.Context, userID int64) (user domain.User, err error) {
	if userID <= 0 {
		err = errors.New("invalid session")
		log.Println(err.Error())
		return
	}
	return u.userProvider.GetUserByUserID(ctx, userID)
}

func (u *userUsecase) GetSession(ctx context.Context, sessionID string) (user domain.Session, err error) {
	if sessionID == ""{
		err = errors.New("invalid session")
		log.Println(err.Error())
		return
	}
	return u.userProvider.GetSession(ctx, sessionID)
}

func (u *userUsecase) Register(ctx context.Context, user domain.User) (err error) {
	if user.Email == "" || user.Username == "" || user.Password == ""{
		err = errors.New("invalid param")
		log.Println(err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return u.userProvider.AddUser(ctx, user)
}

func (u *userUsecase) Login(ctx context.Context, login domain.User) (session string, err error) {
	if login.Email == "" && login.Username == "" {
		err = errors.New("invalid param")
		log.Println(err.Error())
		return
	}
	if login.Password == "" {
		err = errors.New("invalid param")
		log.Println(err.Error())
		return
	}

	user, err := u.getUser(ctx, login.Username, login.Email)
	if err != nil {
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		return
	}

	sessionID, err := u.generateSession()
	if err != nil {
		return
	}
	log.Println(sessionID)

	err = u.userProvider.SetSession(ctx, sessionID, user.UserID)
	return sessionID, err
}

func (u *userUsecase) generateSession() (session string, err error) {
	b := make([]byte, 64)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	return base64.URLEncoding.EncodeToString(b), err
}

func (u *userUsecase) getUser(ctx context.Context, username string, email string) (user domain.User, err error) {
	if username != "" {
		return u.userProvider.GetUserByUsername(ctx, username)
	} else if email != "" {
		return u.userProvider.GetUserByEmail(ctx, email)
	}
	err = errors.New("empty param")
	return
}

func (u *userUsecase) EditUser(ctx context.Context, editedUser domain.User) (err error) {
	if editedUser.Email == "" || editedUser.Username == "" {
		err = errors.New("invalid param")
		log.Println(err.Error())
		return
	}
	return u.userProvider.EditUser(ctx, editedUser)
}

func (u *userUsecase) DeleteUser(ctx context.Context, userID int64, sessionID string) (err error) {
	if userID <= 0 || sessionID == "" {
		err = errors.New("invalid user")
		log.Println(err.Error())
		return
	}
	err = u.userProvider.DeleteUser(ctx, userID)
	if err != nil {
		return
	}

	err = u.userProvider.DeleteSession(ctx, sessionID)
	if err != nil {
		return
	}
	return
}