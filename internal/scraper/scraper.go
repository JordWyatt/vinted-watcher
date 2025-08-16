package scraper

type Scraper interface {
	Scrape(url string) (interface{}, error)
}
