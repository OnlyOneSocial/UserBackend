package apiserver

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/katelinlis/UserBackend/internal/app/store/sqlstore"
)

//Start ...
func Start(config *Config) {
	fmt.Println("Hello world")

	db, err := newDB(config.DatabaseURL)
	if err != nil {
		fmt.Print(err)
	}

	defer db.Close()

	store := sqlstore.New(db)
	srv := newServer(store, config)

	fmt.Println("Start webserver on", config.BindAddr)
	err = http.ListenAndServe(config.BindAddr, srv)

}

func newDB(DatabaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
