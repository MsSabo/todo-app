package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/MsSabo/todo-app"
	"github.com/MsSabo/todo-app/pkg/handler"
	"github.com/MsSabo/todo-app/pkg/kafka"
	"github.com/MsSabo/todo-app/pkg/metrics"
	"github.com/MsSabo/todo-app/pkg/repository"
	"github.com/MsSabo/todo-app/pkg/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	//logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatalf("Error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("failed to initialize do: %s", err.Error())
	}

	fmt.Println("Env password = ", os.Getenv("DB_PASSWORD"))
	time.Sleep(10 * time.Second)
	db, err := repository.NewPostgresDB(repository.Config{
		Host:     "db",
		Port:     "5432",
		Username: "postgres",
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   "postgres",
		SSLMode:  "disable",
	})

	if err != nil {
		logrus.Fatalf("failed to initialize do: %s", err.Error())
	}

	producer, err := kafka.ConnectProducer([]string{"broker:29092"})

	if err != nil {
		panic(fmt.Sprintf("Failed to create kafka producer : %s", err.Error()))
	}

	defer func() {
		producer.Close()
	}()

	time.Sleep(time.Duration(10 * time.Second))

	repository := repository.NewRepository(db)
	service := service.NewService(repository)
	handlers := handler.NewHandler(service, &producer)

	srv := new(todo.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("TodoApp Started")

	worker, err := kafka.ConnectConsumer([]string{"broker:29092"})
	if err != nil {
		panic(err)
	}

	consumer, err := worker.ConsumePartition("my-topic", 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	doneCh := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				println("Msg : topic %s value %s ", msg.Topic, string(msg.Value))
			case <-doneCh:
				fmt.Println("End consumer worker")
				break
			}
		}
	}()

	go func() {
		_ = metrics.Listen("todo-app:8082")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	close(doneCh)
	worker.Close()
	logrus.Print("TodoApp Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}
	logrus.Print("Server shutted down")
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
