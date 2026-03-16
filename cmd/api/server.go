// Filename: cmd/api/server.go
package main

import (
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "time"
	"os/signal"
	"syscall"
	"context"
	"errors"
	"github.com/spector-asael/banking-api/cmd/api/dependencies"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/handlers"
)
func Serve(settings *dependencies.ServerConfig, appInstance *dependencies.ApplicationDependencies) error {


		HandlerDependencies := &handlers.HandlerDependencies{
			Logger: appInstance.Logger,
			Config: appInstance.Config,
		}

		apiServer := &http.Server{
		Addr: fmt.Sprintf(":%d", settings.Port),
		Handler: HandlerDependencies.Routes(), 
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second, 
		ErrorLog: slog.NewLogLogger(appInstance.Logger.Handler(), slog.LevelError),
	}
  // create a channel to keep track of any errors during the shutdown process
   shutdownError := make(chan error)
 // create a goroutine that runs in the background listening
 // for the shutdown signals
   go func() {
      quit := make(chan os.Signal, 1)  // receive the shutdown signal
      signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // signal occurred
      s := <-quit   // blocks until a signal is received
      // message about shutdown in process
      appInstance.Logger.Info("shutting down server", "signal", s.String())
     // create a context
     ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
     defer cancel()


   // initiate the shutdown. If all okay this returns nil
   	shutdownError <- apiServer.Shutdown(ctx)
    }()

	appInstance.Logger.Info("starting server", "address", apiServer.Addr,
		"environment", settings.Environment)
	
   err := apiServer.ListenAndServe()
   if !errors.Is(err, http.ErrServerClosed) {
        return err
    }

	   // check the error channel to see if there were shutdown errors
   err = <-shutdownError
	if err != nil {
			return err
	}
	// graceful shutdown was successful
	appInstance.Logger.Info("stopped server", "address", apiServer.Addr)

	return nil
}