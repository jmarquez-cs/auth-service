package main

import (
	"context"
	"net/url"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

//go:generate swagger generate spec
type server struct {
	ctx			context.Context
	db			*sql.DB	
	origin		*url.URL	
	router		*mux.Router
}

func main() {
	server := &server{
		router: mux.NewRouter(),
	}
	
	server.ConnectCRDB()
	server.Start()
}