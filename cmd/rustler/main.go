package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/stockyard-dev/stockyard-rustler/internal/license"
	"github.com/stockyard-dev/stockyard-rustler/internal/server"
	"github.com/stockyard-dev/stockyard-rustler/internal/store"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("rustler %s\n", version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--health" {
		fmt.Println("ok")
		os.Exit(0)
	}
	log.SetFlags(log.Ltime | log.Lshortfile)
	port := 8980
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil { port = n }
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" { dataDir = "./data" }
	licenseKey := os.Getenv("RUSTLER_LICENSE_KEY")
	licInfo, licErr := license.Validate(licenseKey, "rustler")
	if licenseKey != "" && licErr != nil {
		log.Printf("[license] WARNING: %v — running in free tier", licErr)
		licInfo = nil
	}
	limits := server.LimitsFor(licInfo)
	if licInfo != nil && licInfo.IsPro() {
		log.Printf("  License: Pro (%s)", licInfo.CustomerID)
	} else {
		log.Printf("  License: Free tier")
	}
	db, err := store.Open(dataDir)
	if err != nil { log.Fatalf("database: %v", err) }
	defer db.Close()
	log.Printf("  Stockyard Rustler %s — dead link checker", version)
	log.Printf("  API: http://localhost:%d/api/scans", port)
	log.Printf("  Dashboard: http://localhost:%d/ui", port)
	srv := server.New(db, port, limits)
	if err := srv.Start(); err != nil { log.Fatalf("server: %v", err) }
}
