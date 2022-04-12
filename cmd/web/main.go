package main

import (
	"log"
	"os"

	"github.com/Oloruntobi1/qgdc/internal/cache"
	"github.com/Oloruntobi1/qgdc/internal/database"
	"github.com/Oloruntobi1/qgdc/internal/server"
)

func main() {

	// get a database to use
	db := database.NewRepository(os.Getenv("CURRENT_STORAGE"))
	defer db.Close()

	// get cache to use
	c := cache.GetCurrentCache(os.Getenv("CURRENT_CACHE"))

	// instantiate the server
	server, err := server.NewServer(db, c, os.Getenv("AUTH_SIGNED_SECRET"))
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// start the server
	err = server.Start(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
