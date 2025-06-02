package main

import (
	"github.com/NM711/BioSearchEngine/internal/core"
	"github.com/NM711/BioSearchEngine/internal/db"
)

func main() {
	conn := db.NewDatabaseConn()
	defer conn.Client.Close()
	crawler := core.NewCrawler(conn)
	crawler.Crawl()
}
