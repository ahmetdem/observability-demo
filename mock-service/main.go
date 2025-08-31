package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestCounter, requestDuration)
}

func main() {
	// structured logging
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// RabbitMQ connect with retry
	amqpURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	var amqpConn *amqp.Connection
	var err error
	for i := 0; i < 10; i++ {
		amqpConn, err = amqp.Dial(amqpURL)
		if err == nil {
			break
		}
		log.Warn().Err(err).Msg("RabbitMQ not ready, retrying")
		time.Sleep(2 * time.Second)
	}
	if amqpConn == nil {
		log.Error().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	var amqpChannel *amqp.Channel
	if amqpConn != nil {
		amqpChannel, _ = amqpConn.Channel()
		amqpChannel.ExchangeDeclare("tasks", "fanout", true, false, false, false, nil)
		defer amqpConn.Close()
	}

	mux := http.NewServeMux()

	// main handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		msg := fmt.Sprintf("hello from mock-service at %s", time.Now().Format(time.RFC3339))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg))

		status := "200"
		duration := time.Since(start).Seconds()

		// record metrics
		requestCounter.WithLabelValues(r.URL.Path, r.Method, status).Inc()
		requestDuration.WithLabelValues(r.URL.Path, r.Method, status).Observe(duration)

		// publish to RabbitMQ if available
		if amqpChannel != nil {
			_ = amqpChannel.Publish("tasks", "", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(msg),
			})
		}

		// log structured
		log.Info().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Str("status", status).
			Float64("duration_seconds", duration).
			Msg("handled request")
	})

	// health handler
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// prometheus metrics handler
	mux.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	log.Info().Str("addr", addr).Msg("starting mock-service")

	srv := &http.Server{Addr: addr, Handler: mux}

	// run server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
