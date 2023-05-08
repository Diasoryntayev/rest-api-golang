package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"rest/internal/config"
	"rest/internal/user"
	"rest/internal/user/db"
	"rest/pkg/client/mongodb"
	"rest/pkg/logging"

	"github.com/julienschmidt/httprouter"
)

func main() {
	logger := logging.GetLogger()
	logger.Info("create router")
	router := httprouter.New()

	cfg := config.GetConfig()
	cfgMongo := cfg.MongoDB
	mongoDBClient, err := mongodb.NewClient(context.Background(), cfgMongo.Host, cfgMongo.Port,
		cfgMongo.Username, cfgMongo.Password, cfgMongo.Database, cfgMongo.AuthDB)
	if err != nil {
		panic(err)
	}
	storage := db.NewStorage(mongoDBClient, "users", logger)

	// users, err := storage.FindAll(context.Background())
	// fmt.Println(users)
	// user1 := user.User{
	// 	ID:           "",
	// 	Email:        "dias@gmail.com",
	// 	UserName:     "dias",
	// 	PasswordHash: "12345",
	// }

	// user1ID, err := storage.Create(context.Background(), user1)
	// if err != nil {
	// 	panic(err)
	// }
	// logger.Info(user1ID)

	// user1Found, err := storage.FindOne(context.Background(), user1ID)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(user1Found)

	// user1Found.Email = "newEmail@test.ok"
	// if err = storage.Update(context.Background(), user1Found); err != nil {
	// 	panic(err)
	// }

	// if err = storage.Delete(context.Background(), user1ID); err != nil {
	// 	panic(err)
	// }

	// _, err = storage.FindOne(context.Background(), user1ID)
	// if err != nil {
	// 	panic(err)
	// }

	logger.Info("register user handler")
	handler := user.NewHandler(logger)
	handler.Register(router)

	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger := logging.GetLogger()
	logger.Info("start application")

	var listener net.Listener
	var listenErr error

	if cfg.Listen.Type == "sock" {
		logger.Info("detect app path")
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("creat socket")
		sockethPath := path.Join(appDir, "app.sock")

		logger.Info("listen unix socket")
		listener, listenErr = net.Listen("unix", sockethPath)
		logger.Infof("server is litening unix socket: %s", sockethPath)

	} else {
		logger.Info("listen tcp")
		listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port))
		logger.Infof("server is litening port %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)
	}

	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	server := &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}
