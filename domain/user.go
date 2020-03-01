package domain

const SessionKey = "sid"

type User struct {
	UserID   int64
	Username string
	Email    string
	Password string
}

type Session struct {
	SessionID string
	UserID    string
}
