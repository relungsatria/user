package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"user/domain"
	"user/usecase"
)

type userService struct {
	usecase usecase.UserUsecase
}

type RegisterParam struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginParam struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserDataResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	UserID   int64  `json:"user_id"`
}

type UserEditRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

var userHandler *userService
var once sync.Once

func GetUserHandler() *userService {
	once.Do(func() {
		userHandler = &userService{
			usecase: usecase.NewUserUsecase(),
		}
	})
	return userHandler
}

func readJsonBody(r *http.Request, target interface{}) (err error) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(reqBody, target)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

func (u *userService) GetSession(ctx context.Context, sessionID string) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 300)
	defer cancel()
	return u.usecase.GetSession(ctx, sessionID)
}

func (u *userService) GetUser(r *http.Request) (data interface{}, err error) {
	ctx := r.Context()
	userIDContext := ctx.Value("user_id")
	userID, err := strconv.ParseInt(fmt.Sprint(userIDContext), 10, 64)
	if err != nil {
		return
	}
	user, err := u.usecase.GetUser(ctx, userID)
	if err != nil {
		return
	}
	return UserDataResponse{
		Username: user.Username,
		Email:    user.Email,
		UserID:   user.UserID,
	}, nil
}

func (u *userService) Register(r *http.Request) (data interface{}, err error) {
	ctx := r.Context()

	var request RegisterParam
	err = readJsonBody(r, &request)
	if err != nil {
		return
	}

	err = u.usecase.Register(ctx, domain.User{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	})
	return
}

func (u *userService) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request LoginParam
	err := readJsonBody(r, &request)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	session, err := u.usecase.Login(ctx, domain.User{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	expire := time.Now().AddDate(0, 0, 7)
	cookie := http.Cookie{
		Name:    domain.SessionKey,
		Value:   session,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, &http.Request{Method: http.MethodGet}, r.Host+"/user", http.StatusPermanentRedirect)
	return
}

func (u *userService) UpdateUser(r *http.Request) (data interface{}, err error) {
	ctx := r.Context()
	userIDContext := ctx.Value("user_id")
	userID, err := strconv.ParseInt(fmt.Sprint(userIDContext), 10, 64)
	if err != nil {
		return
	}

	var request UserEditRequest
	readJsonBody(r, &request)

	err = u.usecase.EditUser(ctx, domain.User{
		UserID:   userID,
		Username: request.Username,
		Email:    request.Email,
	})

	return UserDataResponse{
		Username: request.Username,
		Email:    request.Email,
		UserID:   userID,
	}, nil
}


func (u *userService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDContext := ctx.Value("user_id")
	userID, err := strconv.ParseInt(fmt.Sprint(userIDContext), 10, 64)
	if err != nil {
		return
	}

	session, err := r.Cookie(domain.SessionKey)
	if err != nil {
		return
	}

	err = u.usecase.DeleteUser(ctx, userID, session.Value)

	expire := time.Unix(0,0)
	cookie := http.Cookie{
		Name:    domain.SessionKey,
		Value:   "",
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, &http.Request{Method: http.MethodGet}, r.Host + "/user", http.StatusPermanentRedirect)
	return
}