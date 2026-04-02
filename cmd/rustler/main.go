package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-rustler/internal/server";"github.com/stockyard-dev/stockyard-rustler/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9700"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./rustler-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("rustler: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Rustler — Self-hosted web scraper and data collector\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("rustler: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
