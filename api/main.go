package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	_ "modernc.org/sqlite"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := parseConfig()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if cfg.Production {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	environment := "development"
	if cfg.Production {
		environment = "production"
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Environment:      environment,
	})
	if err != nil {
		return err
	}
	defer sentry.Flush(2 * time.Second)

	db, err := sql.Open("sqlite", cfg.DatabaseURI)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	mux := chi.NewMux()
	mux.Use(corsMiddleware)
	mux.Get("/good", goodHandler(db))
	mux.Get("/bad", badHandler(db))

	sentryHandler := sentryhttp.New(sentryhttp.Options{})
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: sentryHandler.Handle(mux),
	}

	var g errgroup.Group
	g.Go(func() error {
		<-ctx.Done()
		log.Debug().Msg("Shutting down server gracefully")
		return srv.Shutdown(context.TODO())
	})
	g.Go(func() error {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})
	return g.Wait()
}

type config struct {
	Production  bool
	Addr        string
	DatabaseURI string
	SentryDSN   string
}

func parseConfig() config {
	cfg := config{
		Production:  false,
		Addr:        ":8080",
		DatabaseURI: ":memory:",
		SentryDSN:   "",
	}
	godotenv.Load()

	if val, ok := os.LookupEnv("PRODUCTION"); ok {
		cfg.Production = val == "true"
	}
	if val, ok := os.LookupEnv("ADDR"); ok {
		cfg.Addr = val
	}
	if val, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DatabaseURI = val
	}
	if val, ok := os.LookupEnv("SENTRY_DSN"); ok {
		cfg.SentryDSN = val
	}
	return cfg
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "baggage, sentry-trace")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func goodHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var epoch int64
		err := db.QueryRowContext(r.Context(), "SELECT unixepoch()").Scan(&epoch)
		if err != nil {
			httpError(r.Context(), w, err, http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Current epoch: %d", epoch)
	}
}

func badHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var epoch int64
		err := db.QueryRowContext(r.Context(), "SELECT unixepochwrongfunction()").Scan(&epoch)
		if err != nil {
			httpError(r.Context(), w, err, http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Current epoch: %d", epoch)
	}
}

func httpError(ctx context.Context, w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	log.Error().Err(err).Send()

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		return
	}
	hub.CaptureException(err)
}
