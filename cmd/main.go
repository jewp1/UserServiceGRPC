package main

import (
	"UsersService/internal/config"
	"UsersService/internal/logger"
	"UsersService/internal/repository"
	"UsersService/internal/service"
	"UsersService/protos/gen"
	"context"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	if err := godotenv.Load(config.EnvPath); err != nil {
		log.Fatal("error loading env file", err)
	}
	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "error processing config"))
	}

	ctx := context.Background()
	newRepository, err := repository.NewRepository(ctx, cfg.PostgreSQL)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error to init repository"))
	}

	logg, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error creating logger"))
	}
	grpcServer := grpc.NewServer()
	authService := service.NewAuthService(cfg, newRepository, logg)
	gen.RegisterAuthServiceServer(grpcServer, authService)

	listen, err := net.Listen("tcp", cfg.Grpc.Port)
	if err != nil {
		logg.Fatal(err, "error creating listener")
	}

	go func() {
		logg.Infof("grpc started listening on port: %s", cfg.Grpc.Port)
		if err := grpcServer.Serve(listen); err != nil {
			logg.Fatal("failed to serve", err)
		}
	}()

	quitSig := make(chan os.Signal, 1)
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	<-quitSig
	logg.Infof("Shutting down grpc server gracefully...")
	grpcServer.GracefulStop()

	logg.Infof("Shutting down postgre gracefully...")
	err = newRepository.ShuttingDownPostgres()
	if err != nil {
		logg.Fatal(errors.Wrap(err, "error shutting down postgres"))
	}
	logg.Infof("shutdown gracefully")

}
