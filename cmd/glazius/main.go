package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/deliseev/glazius/internal/application/usecase"
	"github.com/deliseev/glazius/internal/infrastructure/repository"
	"github.com/deliseev/glazius/internal/infrastructure/tracker"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "glazius")
	dataPath := filepath.Join(configDir, "data.json")

	ctx := context.Background()

	// Инициализация зависимостей (DI контейнер)
	trackerClient, err := tracker.NewRutrackerClient()
	if err != nil {
		log.Fatalf("client error: %v", err)
	}
	repo, err := repository.NewJSONRepository(dataPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Парсинг подкоманд
	switch os.Args[1] {

	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		addCmd.Parse(os.Args[2:])
		url := addCmd.Arg(0)
		if url == "" {
			fmt.Println("Error: URL is required")
			os.Exit(1)
		}

		uc := usecase.NewAddSeriesUseCase(trackerClient, repo)
		err = uc.Execute(ctx, url)
		if err != nil {
			log.Fatalf("error to add: %v", err)
		}
		fmt.Println("Adding:", url)

	case "check":
		uc := usecase.NewCheckUpdatesUseCase(trackerClient, repo)
		uc.Execute(ctx)
		fmt.Println("Checking updates...")

	case "ack":
		id := os.Args[2]
		fmt.Println("Acknowledging:", id)

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: glazius <command> [arguments]")
	fmt.Println("Commands: add, check, ack, list, remove")
}
