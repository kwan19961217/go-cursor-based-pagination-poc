package main

import (
	"context"
	"kwan19961217/cursor-pagination/internal/controller"
	"kwan19961217/cursor-pagination/internal/domain/user"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	app := http.NewServeMux()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return
	}
	defer mongoClient.Disconnect(context.Background())
	userStore := user.NewMongoUserRepository(mongoClient)

	//userStore := user.NewInMemoryUserRepository()
	userService := user.NewUserService(userStore)
	userController := controller.NewUserController(userService)

	app.HandleFunc("GET /users", userController.ListUsers)

	http.ListenAndServe(":8080", app)
}
