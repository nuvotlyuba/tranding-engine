package main

import (
	"context"
	"log"

	"github.com/nuvotlyuba/trading-engine/internal/app"
)

func main() {
	ctx := context.Background()

	err := app.InitAndRun(ctx)
	if err != nil {
		log.Fatalf("Не удалось запустить приложение: %v", err)
	}
}
