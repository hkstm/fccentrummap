package scraper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/hkstm/fccentrummap/internal/repository"
)

const baseURL = "https://fccentrum.nl/categorie/spots/"

func CrawlArticleURLs() ([]string, error) {
	var urls []string
	seen := make(map[string]bool)

	c := colly.NewCollector(
		colly.AllowedDomains("fccentrum.nl"),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		Delay:      500 * time.Millisecond,
	})

	var maxPage int

	c.OnHTML("[data-max-page]", func(e *colly.HTMLElement) {
		if mp, err := strconv.Atoi(e.Attr("data-max-page")); err == nil {
			maxPage = mp
		}
	})

	c.OnHTML(".e-loop-item a.e-con", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if href != "" && !seen[href] {
			seen[href] = true
			urls = append(urls, href)
		}
	})

	if err := c.Visit(baseURL); err != nil {
		return nil, fmt.Errorf("visiting base URL: %w", err)
	}
	c.Wait()

	for page := 2; page <= maxPage; page++ {
		pageURL := fmt.Sprintf("%spage/%d/", baseURL, page)
		if err := c.Visit(pageURL); err != nil {
			return nil, fmt.Errorf("visiting page %d: %w", page, err)
		}
		c.Wait()
	}

	return urls, nil
}

func FetchAndStoreArticles(urls []string, repo *repository.Repository) error {
	c := colly.NewCollector(
		colly.AllowedDomains("fccentrum.nl"),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		Delay:      500 * time.Millisecond,
	})

	var fetchErr error

	c.OnResponse(func(r *colly.Response) {
		if err := repo.InsertArticleRaw(r.Request.URL.String(), string(r.Body)); err != nil {
			fetchErr = fmt.Errorf("storing article %s: %w", r.Request.URL, err)
		}
	})

	for _, url := range urls {
		if fetchErr != nil {
			break
		}
		if err := c.Visit(url); err != nil {
			return fmt.Errorf("fetching article %s: %w", url, err)
		}
	}
	c.Wait()

	return fetchErr
}
