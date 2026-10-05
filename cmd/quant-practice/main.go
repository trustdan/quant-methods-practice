package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/assets"
	"github.com/trustdan/quant-methods-practice/internal/httpapi"
)

const version = "0.1.0-dev"

func main() {
	var (
		portFlag      = flag.Int("port", 0, "Loopback port to bind (default 0 for random available port)")
		noBrowserFlag = flag.Bool("no-browser", false, "Do not automatically launch the web browser")
		versionFlag   = flag.Bool("version", false, "Display application version and exit")
		dataDirFlag   = flag.String("data-dir", "", "Path to user data directory for local storage")
		skipIntroFlag = flag.Bool("skip-intro", false, "Skip startup introduction/arcade sequence")
	)

	flag.Parse()

	if *versionFlag {
		fmt.Printf("quant-practice version %s\n", version)
		os.Exit(0)
	}

	distFS, err := assets.FS()
	if err != nil {
		log.Fatalf("failed to load embedded application assets: %v", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *portFlag)
	srv, err := httpapi.NewServer(httpapi.Config{
		Addr:              addr,
		AssetsFS:          distFS,
		Version:           version,
		AllowedDevOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"},
	})
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	bootstrapURL := srv.BootstrapURL()
	log.Printf("=====================================================")
	log.Printf("Quant Methods Practice v%s", version)
	log.Printf("Listening locally on: http://%s", srv.Addr())
	log.Printf("Application URL:      %s", bootstrapURL)
	if *dataDirFlag != "" {
		log.Printf("Data directory:       %s", *dataDirFlag)
	}
	if *skipIntroFlag {
		log.Printf("Intro sequence:       skipped")
	}
	log.Printf("Press Ctrl+C to stop the local server")
	log.Printf("=====================================================")

	if !*noBrowserFlag {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(bootstrapURL); err != nil {
				log.Printf("Note: could not automatically open browser: %v", err)
				log.Printf("Please open %s manually in your browser.", bootstrapURL)
			}
		}()
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down local server...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = srv.Close()
	<-ctx.Done()
	log.Println("Server stopped cleanly.")
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
