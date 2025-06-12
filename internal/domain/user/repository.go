package user

import (
	"bytes"
	"context"
	"sort"
	"time"

	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const limit = 2

type ByCreatedAt []User

func (a ByCreatedAt) Len() int      { return len(a) }
func (a ByCreatedAt) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool {
	// need a unique key to handle the case when createdAt is the same
	if a[i].CreatedAt.Equal(a[j].CreatedAt) {
		return a[i].ID < a[j].ID
	}
	return a[i].CreatedAt.Before(a[j].CreatedAt)
}

type UserRepository interface {
	ListUsers(start time.Time, end time.Time, order string, userId string) []User
}

type InMemoryUserRepository struct {
	users []User
}

var _ UserRepository = (*InMemoryUserRepository)(nil)

func NewInMemoryUserRepository() UserRepository {
	return &InMemoryUserRepository{
		users: []User{
			{ID: "1", CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ID: "2", CreatedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
			{ID: "3", CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ID: "4", CreatedAt: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)},
			{ID: "5", CreatedAt: time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)},
			{ID: "6", CreatedAt: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
		},
	}
}

func (r *InMemoryUserRepository) ListUsers(start time.Time, end time.Time, order string, lastUserId string) []User {
	if order == "asc" {
		sort.Sort(ByCreatedAt(r.users))
	} else {
		sort.Sort(sort.Reverse(ByCreatedAt(r.users)))
	}

	count := 0
	filteredUsers := lo.Filter(r.users, func(user User, _ int) bool {
		skipUser := false
		if lastUserId != "" {
			// for handling the case when the last request's last user has the same createdAt as the current request's first user
			if order == "asc" && user.CreatedAt.Equal(start) {
				skipUser = user.ID <= lastUserId
			} else if order == "desc" && user.CreatedAt.Equal(end) {
				skipUser = user.ID >= lastUserId
			}
		}
		// createdAt >= start && createdAt <= end
		if (user.CreatedAt.Equal(start) || user.CreatedAt.After(start)) && (user.CreatedAt.Equal(end) || user.CreatedAt.Before(end)) && !skipUser && count < limit {
			count++
			return true
		}
		return false
	})

	return filteredUsers
}

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(client *mongo.Client) UserRepository {
	collection := client.Database("dev").Collection("users")
	return &MongoUserRepository{collection: collection}
}

func (r *MongoUserRepository) ListUsers(start time.Time, end time.Time, order string, lastUserId string) []User {
	filter := bson.M{
		"created_at": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	sort := bson.D{}

	if order == "asc" {
		sort = bson.D{
			{Key: "created_at", Value: 1},
			{Key: "_id", Value: 1},
		}
		if lastUserId != "" {
			objectId, err := bson.ObjectIDFromHex(lastUserId)
			if err != nil {
				return []User{}
			}
			filter = bson.M{
				"$or": bson.A{
					// for handling the case when the last request's last user has the same createdAt as the current request's first user
					bson.M{
						"_id": bson.M{
							"$gt": objectId,
						},
						"created_at": start,
					},
					bson.M{
						"created_at": bson.M{
							"$gt":  start,
							"$lte": end,
						},
					},
				},
			}
		}
	} else {
		sort = bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}
		if lastUserId != "" {
			objectId, err := bson.ObjectIDFromHex(lastUserId)
			if err != nil {
				return []User{}
			}
			filter = bson.M{
				"$or": bson.A{
					// for handling the case when the last request's last user has the same createdAt as the current request's first user
					bson.M{
						"_id": bson.M{
							"$lt": objectId,
						},
						"created_at": end,
					},
					bson.M{
						"created_at": bson.M{
							"$gte": start,
							"$lt":  end,
						},
					},
				},
			}
		}
	}

	opts := options.Find().SetSort(sort).SetLimit(limit)
	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		return []User{}
	}
	defer cursor.Close(context.Background())

	users := []User{}
	for cursor.Next(context.Background()) {
		var raw bson.Raw
		_ = cursor.Decode(&raw)
		decoder := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(raw)))
		decoder.ObjectIDAsHexString()

		var user User
		if err := decoder.Decode(&user); err != nil {
			return []User{}
		}

		users = append(users, user)
	}
	return users
}
