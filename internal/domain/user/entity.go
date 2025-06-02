package user

import "time"

type User struct {
	ID        string    `bson:"_id"`
	CreatedAt time.Time `bson:"created_at"`
}

type UserList struct {
	Users      []User `json:"users"`
	NextCursor string `json:"next_cursor"`
}

type UserCursor struct {
	UserId string    `json:"user_id"`
	Order  string    `json:"order"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}
