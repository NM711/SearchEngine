package core

import (
	"log"
	"net/url"
	"os/signal"

	"github.com/NM711/BioSearchEngine/internal/db"
	"github.com/gocolly/colly"
)

type DomainLevel string

const (
	TopDomain  DomainLevel = "TOP"
	RootDomain DomainLevel = "ROOT"
	SubDomain  DomainLevel = "SUB"
)

type Crawler struct {
	collector *colly.Collector
	queries   *db.CrawlerResourceQueries
}

/*
	Parses page meta data or generates its own off of <h1> and <p> tags.
*/

func (c *Crawler) parsePageMeta() {
}

/*
	Parses page anchors <a> and saves their hypertext reference to the database
*/

func (c *Crawler) parsePageLinks() {
	c.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// 1. Check if resource has been saved, if it has compare domain path with resource domain path
		// 2. If path matches, simply yoink the route and save it in a ResourceRoute table linked to the resource in question

		href := e.Attr("href")

		link, err := url.Parse(href)

		if err != nil {
			log.Printf("Failed to parse url \"%s\" because: %s\n", href, err.Error())
		}

		domainPath := link.Hostname()

		var resourceID int64

		resourceID, err = c.queries.CreateResource(domainPath)

		var errcode, ok = db.GetQueryErrorCode(err)

		if ok && errcode == db.QueryErrorCode(db.UniqueViolation) {
			resourceID = c.queries.GetResourceID(domainPath)

			// Strip the route and push it to ResourceRoute, (include query params in route path.) ---> NOTE IGNORING QUERY PARAMS AS OF NOW, WILL WORK
			// THROUGH A PREPARED WAY TO INDEX AND USE THEM IN SEARCH. FOR NOW WE WILL ONLY WORK WITH ROUTES.

			// NOTE ROUTES SHOULD BE SAVED AS NORMAL IF ASSOCIATED WITH THE PAGE.
			// IF DURING SEARCH CYCLE ROUTE RESULTS IN 5xx or 4xx ERRORS FLAG IT AND SET A GRACE PERIOD IN DATABASE BEFORE SCHEDULING A REMOVAL
			// SAY A PERIOD OF 3 DAYS SINCE ITS A SMALL ENGINE.
			route := link.Path

			c.queries.CreateResourceRoute(resourceID, route)
			return
		}

	})
}

// This configures the colllector, its better than making multiple calls over and over per iteration.
// We just need to define the callbacks once and colly will run them when needed.
func (c *Crawler) initScanConfig() {
	c.parsePageMeta()
	c.parsePageLinks()
}

func (c *Crawler) scan(resource *db.Resource) {
	// Start by scanning resource, then its routes. Then move onto the next resource.
	log.Printf("Scanning resource ID \"%d\" - FullDomain \"%s\"", resource.ID, resource.FullDomain)
	url := "https://" + resource.FullDomain
	err := c.collector.Visit(url)

	// If visit error just skip, it will most likely be caused by invalid https addresses. Our crawler will not index non https pages.
	if err != nil {
		log.Printf("Failed to visit url \"%s\" because: %s\n", url, err.Error())
	}
}

/*
	Algorithm here should be pretty simple.
	1. Per page scrape all anchors and meta tags.
	2. Per url that has not been logged, add to database (create domain and resource, create a page route)
	3. To move on to the next page to scrape, I suppose we will have some sort of priority level per resource and we will start from highest to lowest.
*/

func (c *Crawler) Crawl() {
	log.Println("Crawler has been started...")
	resourceBatch := c.queries.SelectResources()
	// The collector kinda works on something akin to an event or hook/callback system so its parse functions only need to be called once
	// To register the actual callbacks to be executed
	c.initScanConfig()
	for _, resource := range *resourceBatch {
		c.scan(&resource)
	}
}

func NewCrawler(conn *db.DatabaseConnection) *Crawler {
	return &Crawler{colly.NewCollector(), db.NewCrawlerResourceQueries(conn)}
}
