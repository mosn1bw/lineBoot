package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Print("Server Start...")
	repo = NewDB()
	Migrate()
	app, err := NewNBABotClient(_config.Channel.Secret, _config.Channel.Token, _config.AppBaseURL)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Static("/static", "./static")
	router.Static("/downloaded", "./downloaded")
	router.POST("/callback", app.Callback)
	router.GET("/gamecol/info", app.getGameColumnInfo)
	router.GET("/game/:gameid/:type", app.getGamePlayInfoEN)
	router.GET("/standing/:conference", app.getStandingInfo)

	// admin
	router.GET("/statistic", app.Statistic)

	router.GET("/messages/", app.ListMessages)
	router.GET("/messages/rawdata", app.ListMessagesRawData)

	srv := &http.Server{
		Addr:    ":" + _config.Bind,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
