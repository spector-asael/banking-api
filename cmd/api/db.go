// File: cmd/api/main.go

package main 

import (
	"time"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/spector-asael/banking-api/cmd/api/dependencies"
)

func openDB(settings dependencies.ServerConfig) (*sql.DB, error) {
    // open a connection pool
    db, err := sql.Open("postgres", settings.DB.DSN)
    if err != nil {
        return nil, err
    }
    
    // set a context to ensure DB operations don't take too long
    ctx, cancel := context.WithTimeout(context.Background(),
                                       5 * time.Second)
    defer cancel()
    // let's test if the connection pool was created
    // we trying pinging it with a 5-second timeout
    err = db.PingContext(ctx)
    if err != nil {
        db.Close()
        return nil, err
    }


    // return the connection pool (sql.DB)
    return db, nil
} 
