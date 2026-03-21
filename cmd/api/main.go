package main 

import (
	// "log/slog"
	// "github.com/spector-asael/banking-api/cmd/api/dependencies"
	"fmt"
	"flag"
	"log/slog"
	"os"
	"github.com/spector-asael/banking-api/cmd/api/dependencies"
	"github.com/spector-asael/banking-api/internal/data"
	"strings"
	"expvar"
)

func main() {

	var settings dependencies.ServerConfig 
	const appVersion = "1.0.0"

	fmt.Println("Starting API server...")

	flag.IntVar(&settings.Port, "port", 4000, "Server port")
	flag.StringVar(&settings.Environment, "env", "development", "Environment(development|staging|production)")
	flag.StringVar(&settings.DB.DSN, "db-dsn", "", "PostgreSQL DSN")
    flag.Float64Var(&settings.Limiter.RPS, "limiter-rps", 2, "Rate Limiter maximum requests per second")

	flag.IntVar(&settings.Limiter.Burst, "limiter-burst", 5, "Rate Limiter maximum burst")

    flag.BoolVar(&settings.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", 
		func(val string) error {
		settings.Cors.TrustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// release the database resources before exiting
	// defer db.Close()
	defer db.Close()

	logger.Info("database connection pool established")

	expvar.NewString("version").Set(appVersion)

	appInstance := &dependencies.ApplicationDependencies {
		Config: settings,
		Logger: logger,
		Models: data.Models{},
	}

	err = Serve(&settings, appInstance)
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }

}