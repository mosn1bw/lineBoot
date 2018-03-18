package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Print("Server Start...")
	app, err := NewNBABotClient(_config.Channel.Secret, _config.Channel.Token, _config.AppBaseURL)
	if err != nil {
		log.Fatal(err)
	}
	// serve /static/** files
	staticFileServer := http.FileServer(http.Dir("static"))
	http.HandleFunc("/static/", http.StripPrefix("/static/", staticFileServer).ServeHTTP)
	// serve /downloaded/** files
	downloadedFileServer := http.FileServer(http.Dir(app.downloadDir))
	http.HandleFunc("/downloaded/", http.StripPrefix("/downloaded/", downloadedFileServer).ServeHTTP)

	http.HandleFunc("/callback", app.Callback)

	http.HandleFunc("/statistic", app.Statistic)

	addr := fmt.Sprintf(":%s", _config.Bind)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
