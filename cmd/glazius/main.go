package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/deliseev/glazius/internal/application/usecase"
	"github.com/deliseev/glazius/internal/infrastructure/repository"
	"github.com/deliseev/glazius/internal/infrastructure/storage"
	"github.com/deliseev/glazius/internal/infrastructure/torrent"
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
	torrentsPath := filepath.Join(configDir, "torrents")

	ctx := context.Background()

	// Инициализация зависимостей (DI контейнер)
	trackerClient := tracker.NewRutrackerClient(http.DefaultClient)
	repo, err := repository.NewJSONRepository(dataPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	storage, err := storage.NewFileTorrentStorage(torrentsPath)
	if err != nil {
		log.Fatalf("torrent cache dir error: %v", err)
	}
	parser := &torrent.BencodeParser{}

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
		err := uc.Execute(ctx, url)
		if err != nil {
			log.Fatalf("error to add: %v", err)
		}
		fmt.Println("Adding:", url)

	case "check":
		uc := usecase.NewCheckUpdatesUseCase(trackerClient, repo, storage, parser)
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
