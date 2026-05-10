package contentfetch

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/hkstm/fccentrummap/internal/articletext"
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

	maxPage := 0

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

	if maxPage <= 0 {
		return nil, fmt.Errorf("missing or invalid pagination metadata: data-max-page")
	}

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

	c.OnError(func(r *colly.Response, err error) {
		if fetchErr != nil {
			return
		}
		if r != nil && r.Request != nil {
			fetchErr = fmt.Errorf("fetching article %s: %w", r.Request.URL, err)
			return
		}
		fetchErr = fmt.Errorf("fetching article: %w", err)
	})

	c.OnResponse(func(r *colly.Response) {
		html := string(r.Body)
		videoID, ok := ExtractYouTubeVideoID(html)
		var maybeVideoID *string
		if ok {
			maybeVideoID = &videoID
		}

		if err := repo.InsertArticleRaw(r.Request.URL.String(), html, maybeVideoID); err != nil {
			if fetchErr == nil {
				fetchErr = fmt.Errorf("storing article %s: %w", r.Request.URL, err)
			}
			return
		}

		articleRaw, err := repo.GetArticleRawByURL(r.Request.URL.String())
		if err != nil {
			if fetchErr == nil {
				fetchErr = fmt.Errorf("lookup article_raw after insert %s: %w", r.Request.URL.String(), err)
			}
			log.Printf("ERROR: lookup article_raw after insert failed url=%s reason=%v", r.Request.URL.String(), err)
			return
		}
		if articleRaw == nil {
			if fetchErr == nil {
				fetchErr = fmt.Errorf("lookup article_raw after insert %s: missing row", r.Request.URL.String())
			}
			log.Printf("ERROR: article_raw missing after insert url=%s", r.Request.URL.String())
			return
		}

		extractionResult := articletext.ExtractArticleTextContent(html)
		extractionResult.ArticleRawID = articleRaw.ArticleRawID
		if err := repo.ReplaceArticleTextExtraction(extractionResult); err != nil {
			if fetchErr == nil {
				fetchErr = fmt.Errorf("persisting article text extraction article_raw_id=%d url=%s: %w", articleRaw.ArticleRawID, r.Request.URL.String(), err)
			}
			log.Printf("ERROR: persisting article text extraction failed article_raw_id=%d url=%s reason=%v", articleRaw.ArticleRawID, r.Request.URL.String(), err)
			return
		}

		if extractionResult.ErrorMessage != nil {
			log.Printf("INFO: article text extraction article_raw_id=%d url=%s status=%s mode=%s matched_count=%d error=%s", articleRaw.ArticleRawID, r.Request.URL.String(), extractionResult.Status, extractionResult.ExtractionMode, extractionResult.MatchedCount, *extractionResult.ErrorMessage)
			return
		}
		log.Printf("INFO: article text extraction article_raw_id=%d url=%s status=%s mode=%s matched_count=%d", articleRaw.ArticleRawID, r.Request.URL.String(), extractionResult.Status, extractionResult.ExtractionMode, extractionResult.MatchedCount)
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
