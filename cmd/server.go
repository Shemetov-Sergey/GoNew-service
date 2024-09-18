package main

import (
	"context"
	"fmt"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/cache"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/client"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/config"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/middleware"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/pb/gonews"

	"log"
	"net"
	"time"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/rss"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/services"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/storage"
	"github.com/Shemetov-Sergey/GoNew-service/pkg/storage/postgres"

	"google.golang.org/grpc"
)

func main() {
	log.Println("Start GoNews")

	c, err := config.LoadConfig()

	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	connString := c.DBUrl

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	postsChan := make(chan storage.Post)
	errChan := make(chan error)
	runErrorsCheck(errChan)
	db, err := postgres.New(ctx, connString, postsChan, errChan)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Db.Close()

	configFile := "config.json"
	sourceRss, err := rss.NewSourceRss(configFile, postsChan, errChan)

	if err != nil {
		log.Fatal(err)
	}

	// Получаем список последних публикаций из базы данных (если они есть). Чтобы не дублировать данные в бд.
	lastPubTimeMap, _ := db.GetLastPubDateForSources(sourceRss.SourceRssList)
	sourceRss.LastPubTimeFromDB = lastPubTimeMap

	sourceRss.RunGetSourcesInfo()
	db.RunInsertPosts()

	pc := cache.NewPaginationCache()
	commentCli := client.InitServiceClient(&c)

	s := services.Server{
		H: db,
		P: pc,
		C: commentCli,
	}

	lis, err := net.Listen("tcp", c.Port)

	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	fmt.Println("Auth Svc on", c.Port)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor,
		),
	)

	gonews.RegisterGoNewsServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}

// runErrorsCheck читает ошибки из канала ошибок
func runErrorsCheck(errChan chan error) {
	go func() {
		for {
			select {
			case err := <-errChan:
				log.Println(err)
			}
		}
	}()
}
