package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// The seeder will only build the queries for us and return a sql file that we can hook up to compose.

func main() {
	data, err := os.ReadFile("data/urls.txt")

	if err != nil {
		log.Fatalf("Could not read seed file because: %s\n", err.Error())
	}

	urlArray := strings.Split(string(data), "\n")

	var insertString = "INSERT INTO Resource (full_domain) VALUES\n"

	for i, rawURL := range urlArray {

		if len(rawURL) == 0 {
			continue
		}

		parsedURL, err := url.Parse(rawURL)

		if err != nil {
			log.Fatalf(`Failed to parse raw  url "%s" because: \n`, rawURL, err.Error())
		}

		domainPath := parsedURL.Hostname()

		if i < len(urlArray)-2 {
			insertString += fmt.Sprintf("\t('%s'),\n", domainPath)
		} else {
			insertString += fmt.Sprintf("\t('%s');\n", domainPath)
		}

	}

	path := "docker/mysql/seed.sql"

	err = os.WriteFile(path, []byte(insertString), 0744)

	if err != nil {
		log.Fatalf("Failed to write seeder sql file because: %s\n", err.Error())
	}

	log.Println(`SQL seed file written!`)
}
