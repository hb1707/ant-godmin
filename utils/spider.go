package utils

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type SiteMapCache struct {
	LastUpdate time.Time
	Sitemap
}

var SiteMapCacheMap = make(map[string]SiteMapCache)

// Sitemap represents the structure of a sitemap.xml
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

// URL represents a single URL entry in a sitemap.xml
type URL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

var lockCheckUpdateFromSitemapUpdate sync.Mutex

// CheckUpdateFromSitemapUpdate checks if url has been updated
func CheckUpdateFromSitemapUpdate(path string) (time.Time, error) {
	// 获取网站域名
	sitemapUrl := fmt.Sprintf("%s/sitemap.xml", GetDomain(path))
	var sitemap Sitemap
	if _, ok := SiteMapCacheMap[sitemapUrl]; !ok || time.Since(SiteMapCacheMap[sitemapUrl].LastUpdate) > 5*time.Minute {
		lockCheckUpdateFromSitemapUpdate.Lock()
		defer lockCheckUpdateFromSitemapUpdate.Unlock()
		SiteMapCacheMap[sitemapUrl] = SiteMapCache{
			LastUpdate: time.Now(),
			Sitemap:    sitemap,
		}
		resp, err := http.Get(sitemapUrl)
		if err != nil {
			return time.Time{}, fmt.Errorf("error fetching sitemap: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return time.Time{}, fmt.Errorf("error reading response body: %v", err)
		}
		err = xml.Unmarshal(body, &sitemap)
		if err != nil {
			return time.Time{}, fmt.Errorf("error parsing XML: %v", err)
		}
		SiteMapCacheMap[sitemapUrl] = SiteMapCache{
			LastUpdate: time.Now(),
			Sitemap:    sitemap,
		}
		//log.Warning("SiteMapCacheMap", sitemapUrl)
	} else {
		sitemap = SiteMapCacheMap[sitemapUrl].Sitemap
	}
	var modTime time.Time
	for _, v := range sitemap.URLs {
		if strings.Contains(v.Loc, path) {
			modTime, err := time.Parse(time.RFC3339, v.LastMod)
			if err != nil {
				continue // Skip entries with parsing errors
			}
			return modTime, nil
		}
	}
	//
	//if modTime.IsZero() {
	//	return time.Time{}, fmt.Errorf("no valid lastmod dates found")
	//}
	return modTime, nil
}
func GetDomain(urlStr string) string {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
}
