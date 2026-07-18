package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/deliseev/glazius/internal/application/usecase"
	"github.com/deliseev/glazius/internal/infrastructure/config"
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
	configPath := filepath.Join(configDir, "config.json")
	torrentsPath := filepath.Join(configDir, "torrents")

	ctx := context.Background()

	config, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("loading config error: %v", err)
	}

	// Инициализация зависимостей (DI контейнер)
	trackerClient, err := tracker.NewRutrackerClient()
	if err != nil {
		log.Fatalf("client error: %v", err)
	}
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
	case "login":
		// Инициализируем клиент (который создает новый Jar)
		client, _ := tracker.NewRutrackerClient()
		err := client.Login(ctx, "dan1us", "RanyX")
		if err != nil {
			log.Fatalf("Login failed: %v", err)
		}
		fmt.Println("Login successful! Session cookies saved in Jar.")
		return // Просто выходим после успеха

	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		addCmd.Parse(os.Args[2:])
		url := addCmd.Arg(0)
		if url == "" {
			fmt.Println("Error: URL is required")
			os.Exit(1)
		}
		err = trackerClient.Login(ctx, config.Username, config.Password)
		if err != nil {
			log.Fatal(err)
		}

		// Попробуем скачать торрент напрямую по topic_id (из твоего debug.html это 6880649)
		torrentBytes, err := trackerClient.DownloadTorrent(ctx, url)
		if err != nil {
			log.Fatalf("Download failed: %v", err)
		}

		fmt.Printf("Скачано байт: %d\n", len(torrentBytes))

		uc := usecase.NewAddSeriesUseCase(trackerClient, repo, storage)
		err = uc.Execute(ctx, url)
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
