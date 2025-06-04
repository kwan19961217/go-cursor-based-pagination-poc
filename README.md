# Go Cursor-based Pagination

This is a POC implementation of how a cursor-based pagination works.
This demonstration will be paginated on the `User` model by `created_at`, and using the `id` as unique key for unambiguously identify which row were discovered.

## Setup

The `User` model is defined as follow.

| User       |            |
| ---------- | ---------- |
| id         | Unique Key |
| created_at | Timestamp  |

You will need to setup a MongoDB in your local environment.

Or alternatively, comment out the following lines in `cmd/httpserver/main.go`

```
	mongoClient, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return
	}
	defer mongoClient.Disconnect(context.Background())
	userStore := user.NewMongoUserRepository(mongoClient)
```

And uncomment the following

```
//userStore := user.NewInMemoryUserRepository()
```

to use a in-memory storage.

Setup your mock data according to your storage option, you can find the mock data for in-memory storage in `internal/domain/user/repository.go`

## Usage

For first request, the following parameters are required.

| Query Param | Description                     |
| ----------- | ------------------------------- |
| start       | RFC3339 Timestamp (inclusive)   |
| end         | RFC3339 Timestamp (inclusive)   |
| order       | Sort direction, `asc` or `desc` |

Example:

`GET localhost:8080/users?start=2025-01-01T00:00:00Z&end=2025-01-05T00:00:00Z&order=asc`

For subsequent request, only the following parameter is required.

| Query Param | Description    |
| ----------- | -------------- |
| next_cursor | Encoded string |

Example:

`GET localhost:8080/users?next_cursor=eyJlbmQiOiIyMDI1LTAxLTA1VDAwOjAwOjAwWiIsIm9yZGVyIjoiYXNjIiwic3RhcnQiOiIyMDI1LTAxLTAxVDAwOjAwOjAwWiIsInVzZXJfaWQiOiI2ODNkN2NiYzRhY2JmOTE5ZDdjZDdiNzUifQ==`
