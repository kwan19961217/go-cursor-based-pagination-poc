package user

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type UserService struct {
	userRepository UserRepository
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) ListUsers(start time.Time, end time.Time, order string, userId string) (*UserList, error) {
	users := s.userRepository.ListUsers(start, end, order, userId)

	if len(users) == 0 {
		return &UserList{Users: users, NextCursor: ""}, nil
	}

	b, err := json.Marshal(map[string]any{
		"start":   start,
		"end":     end,
		"order":   order,
		"user_id": users[len(users)-1].ID,
	})
	if err != nil {
		return nil, err
	}

	nextCursor := base64.StdEncoding.EncodeToString(b)
	return &UserList{Users: users, NextCursor: nextCursor}, nil
}
