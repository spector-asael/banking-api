package main

import (
	// "log/slog"
	// "github.com/spector-asael/banking-api/cmd/api/dependencies"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spector-asael/banking-api/cmd/api/dependencies"
	"github.com/spector-asael/banking-api/internal/data"
	"github.com/spector-asael/banking-api/internal/mailer"
)

func main() {

	var settings dependencies.ServerConfig
	const appVersion = "1.0.0"

	fmt.Println("Starting API server...")

	flag.IntVar(&settings.Port, "port", 4000, "Server port")
	flag.StringVar(&settings.Environment, "env", "development", "Environment(development|staging|production)")
	flag.StringVar(&settings.DB.DSN, "db-dsn", "", "PostgreSQL DSN")
	flag.Float64Var(&settings.Limiter.RPS, "limiter-rps", 2, "Rate Limiter maximum requests per second")
	flag.StringVar(&settings.SMTP.Host,
		"smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	// We have port 25, 465, 587, 2525. If 25 doesn't work choose another
	flag.IntVar(&settings.SMTP.Port, "smtp-port", 2525, "SMTP port")
	// Use your Username value provided by Mailtrap
	flag.StringVar(&settings.SMTP.Username, "smtp-username",
		"", "SMTP username")

	// Use your Password value provided by Mailtrap
	flag.StringVar(&settings.SMTP.Password, "smtp-password",
		"", "SMTP password")

	flag.StringVar(&settings.SMTP.Sender, "smtp-sender",
		"Craboo Bank <no-reply@craboobank.spector.net>",
		"SMTP sender")

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
	// The number of active goroutines
	// is a useful metric to monitor in a production application
	// as it can help you identify potential performance issues or bottlenecks.
	// By publishing this information using expvar,
	// you can easily track the number of active goroutines over time
	// and identify any spikes or trends that may indicate a problem.
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	// The database connection pool metrics
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	// THe current timestamp
	expvar.Publish("timestamp", expvar.Func(func() any {
		return fmt.Sprintf("%d", time.Now().Unix())
	}))

	models := data.Models{}.NewModels(db)
	appInstance := &dependencies.ApplicationDependencies{
		Config: settings,
		Logger: logger,
		Models: models,
		Mailer: mailer.New(settings.SMTP.Host, settings.SMTP.Port,
			settings.SMTP.Username, settings.SMTP.Password, settings.SMTP.Sender),
	}

	err = Serve(&settings, appInstance)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

}
