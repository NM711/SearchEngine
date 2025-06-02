package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
)

type CrawlerResourceQueries struct {
	dbc *DatabaseConnection
}

type Resource struct {
	ID         int
	FullDomain string
}

type QueryErrorCode uint

const (
	UniqueViolation QueryErrorCode = iota
)

type QueryError struct {
	Code QueryErrorCode
	Mssg string
}

func GetQueryErrorCode(err error) (QueryErrorCode, bool) {
	if queryErr, ok := err.(*QueryError); ok {
		return queryErr.Code, true
	}
	return 0, false
}

func (q *QueryError) Error() string {
	return fmt.Sprintf("(%d) %s", q.Code, q.Mssg)
}

func errorToQueryErrorCode(err error) *QueryError {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case 1062:
			return &QueryError{QueryErrorCode(UniqueViolation), "unique violation : " + err.Error()}
		default:
			break
		}
	}

	return nil
}

func scanResourceRows(rows *sql.Rows) *[]Resource {
	var resources []Resource
	for rows.Next() {
		var resource Resource

		if err := rows.Scan(&resource.ID, &resource.FullDomain); err != nil {
			log.Fatalf("Failed to scan row into Resource because:%s\n", err.Error())
		}

		resources = append(resources, resource)
	}

	return &resources
}

func (r *CrawlerResourceQueries) GetResourceID(path string) int64 {
	var id int64
	err := squirrel.Select("id").From("Resource").Where(squirrel.Eq{"full_domain": path}).QueryRow().Scan(&id)

	if err != nil {
		log.Fatalf("Failed to retrieve resource with id of \"%d\" because: %s\n", id, err.Error())
	}

	return id
}

/**
Simple algorithm to return priority resource batches, for the crawlere to crawl.

Returns resources based on priority, returned in batches of 500.
1. NULL last_crawled
2. Whatever is oldest

In the future i could extend this to make it so more total resource routes visit means a greater importance within my web.
Likewise, similar to how the ranking algo will be built, more external references to resources means a higher resource priority.
Will factor this in some other time though, for now im going to keep it last_crawled time based.
*/

func (r *CrawlerResourceQueries) SelectResources() *[]Resource {
	var selectResources squirrel.SelectBuilder

	selectResources = squirrel.
		Select("id", "full_domain").
		From("Resource").
		Limit(500)

	rows, err := selectResources.Where(squirrel.Eq{"last_crawled": nil}).RunWith(r.dbc.Client).Query()

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Could not select any rows that contained uncrawled resources!")
			log.Println("Falling back to filter by earliest last crawled date...")
		} else {
			log.Fatalf("Could not query resources because: %s\n", err.Error())
		}
	} else {
		return scanResourceRows(rows)
	}

	rows, err = selectResources.OrderBy("last_crawled ASC").RunWith(r.dbc.Client).Query()

	if err != nil {
		log.Fatalf("Could not query earliest last crawled resources because: %s\n", err.Error())
	}

	return scanResourceRows(rows)
}

func (r *CrawlerResourceQueries) CreateResource(fullDomainPath string) (int64, error) {
	resourceInsert := squirrel.Insert("Resource").Columns("full_domain").Values(fullDomainPath)
	result, err := resourceInsert.RunWith(r.dbc.Client).Exec()
	// Instead of simply logging return an error code in the form of an enum i think.
	// This allows the crawler to handle an appropriate fallback based on some kind of signal
	if err != nil {
		queryErr := errorToQueryErrorCode(err)
		// If its not some query error it could be a network error, invalid service error, idk, we terminate program in that case.
		if queryErr == nil {
			log.Fatalf("Failed to insert resource with domain path \"%s\" because: %s\n", fullDomainPath, err.Error())
		}

		return 0, queryErr
	}

	// Get the ID of the inserted resource
	id, err := result.LastInsertId()

	if err != nil {
		log.Fatalf("Failed to get last insert ID for resource with domain path \"%s\" because: %s\n", fullDomainPath, err.Error())
	}

	log.Printf("Successfully inserted resource with domain path \"%s\" and ID %d!\n", fullDomainPath, id)
	return id, nil
}

/*
	Updates the last crawled date of a given reasource to NOW()
*/

func (r *CrawlerResourceQueries) UpdateResourceCrawlDate(id int) {
	resourceUpdate := squirrel.Update("Resource").Set("last_crawled", "NOW()").Where(squirrel.Eq{"id": id})

	_, err := resourceUpdate.RunWith(r.dbc.Client).Exec()

	if err != nil {
		log.Fatalf("Something occurred when attempting to update resource last_crawled date to NOW()! %s\n", err.Error())
	}

	log.Printf("Updated last_crawled date of resource with id \"%d\"!\n", id)
}

func (r *CrawlerResourceQueries) CreateResourceRoute(id int64, routePath string) {
	resourceRouteInsert := squirrel.Insert("ResourceRoute").Columns("resource_id", "route_path").Values(id, routePath)
	_, err := resourceRouteInsert.RunWith(r.dbc.Client).Exec()

	if err != nil {
		log.Fatalf("Something occurred when attempting to insert route \"%s\" for resource id \"%d\" because: %s\n", routePath, id, err.Error())
	}

	log.Printf("Successfully created route \"%s\" for resource id \"%d\"\n", routePath, id)
}

func (r *CrawlerResourceQueries) CreateResourceReference() {

}

func NewCrawlerResourceQueries(conn *DatabaseConnection) *CrawlerResourceQueries {
	return &CrawlerResourceQueries{conn}
}
