package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/devfullcycle/otel/comunicacao-ms/internal/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	if os.Getenv("WEATHER_API_KEY") == "" {
		log.Fatal("WEATHER_API_KEY environment variable is required")
	}

	exporter, err := zipkin.New(
		"http://zipkin:9411/api/v2/spans",
		zipkin.WithLogger(log.New(os.Stderr, "zipkin: ", log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		log.Fatal(err)
	}

	batcher := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(batcher),
	)
	otel.SetTracerProvider(tp)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	server := web.NewWebServer()
	server.Serve()
}
