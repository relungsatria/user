package service

import (
	"context"
	"encoding/json"
	"github.com/pressly/chi"
	"log"
	"net/http"
	"user/service/handler"
)

type Service interface {
	Run()
}

type Option struct {
	ListenPort string
}

type service struct {
	option Option
	router *chi.Mux
}

type StandardResponse struct {
	Status int         `json:"status"`
	Error  string      `json:"error"`
	Data   interface{} `json:"data,omitempty"`
}

func (s *service) Run() {
	log.Println("running service on", s.option.ListenPort)
	log.Fatal(http.ListenAndServe(s.option.ListenPort, s.router))
}

func New(option Option) Service {
	r := chi.NewRouter()

	r.Route("/", registrationService())
	r.Route("/user", userService())

	return &service{
		option: option,
		router: r,
	}
}

func userService() func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(authenticatedOnly)
		r.Get("/", serviceWrapper(handler.GetUserHandler().GetUser))
		r.Put("/", serviceWrapper(handler.GetUserHandler().UpdateUser))
		r.Delete("/", handler.GetUserHandler().DeleteUser)
	}
}

func registrationService() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/register", serviceWrapper(handler.GetUserHandler().Register))
		r.Post("/login", handler.GetUserHandler().Login)
	}
}

func authenticatedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sid")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		session, err := handler.GetUserHandler().GetSession(ctx, cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, "user_id", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func serviceWrapper(f func(r *http.Request) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(r)
		response := StandardResponse{
			Status: 1,
			Error:  "",
			Data:   data,
		}
		if err != nil {
			response.Status = 0
			response.Error = "server error"
			w.WriteHeader(http.StatusInternalServerError)
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		log.Println(string(jsonResponse))
		w.Write(jsonResponse)
	}
}
