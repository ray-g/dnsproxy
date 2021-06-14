package blocker

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	c "github.com/ray-g/dnsproxy/cache"
	r "github.com/ray-g/dnsproxy/cache/record"
	conf "github.com/ray-g/dnsproxy/config"
	"github.com/ray-g/dnsproxy/logger"
	"github.com/ray-g/dnsproxy/stats"
	"github.com/ray-g/dnsproxy/utils"
)

var whitelist = make(map[string]bool)

// Update downloads all of the blocklists and imports them into the database
func update(config *conf.BlockerConfig, cache c.Cache, force bool) error {
	for _, entry := range config.Whitelist {
		whitelist[entry] = true
	}

	for _, entry := range config.Blocklist {
		cache.Set(entry, r.NewBlockedRecord())
		stats.AddBlockedDomain()
	}

	if err := fetchSources(config.SourceURLs, config.SourceDir, force); err != nil {
		return fmt.Errorf("error fetching sources: %s", err)
	}

	return nil
}

func downloadFile(uri string, name string, sourcedir string) error {
	utils.EnsureDirectory(sourcedir)
	filePath := filepath.FromSlash(filepath.Join(sourcedir, name))

	output, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %s", err)
	}
	defer output.Close()

	response, err := http.Get(uri)
	if err != nil {
		return fmt.Errorf("error downloading source: %s", err)
	}
	defer response.Body.Close()

	if _, err := io.Copy(output, response.Body); err != nil {
		return fmt.Errorf("error copying output: %s", err)
	}

	return nil
}

func fetchSources(sources []conf.DNSBlockSource, sourceDir string, force bool) error {
	var wg sync.WaitGroup

	for _, s := range sources {
		filename := fmt.Sprintf("%s.list", s.Name)
		_, err := os.Stat(filepath.Join(sourceDir, filename))
		if err == nil && !force {
			continue
		}

		uri := s.URL

		wg.Add(1)
		go func(uri string, name string) {
			logger.Debugf("fetching source %s", uri)
			if err := downloadFile(uri, name, sourceDir); err != nil {
				logger.Error("failed to download source, err: %v", err)
			}

			wg.Done()
		}(uri, filename)
	}

	wg.Wait()

	return nil
}

// UpdateBlockCache updates the BlockCache
func updateBlockCache(cache c.Cache, sourceDir string) error {
	logger.Debugf("loading blocked domains from %s ...", sourceDir)

	err := filepath.Walk(sourceDir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			fileName := filepath.FromSlash(path)

			if err := parseHostFile(fileName, cache); err != nil {
				return fmt.Errorf("error parsing hostfile %s", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking location %s", err)
	}

	logger.Debugf("%d domains loaded from sources", cache.Length())

	return nil
}

func parseHostFile(fileName string, cache c.Cache) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		if len(line) > 0 {
			fields := strings.Fields(line)

			if len(fields) > 1 {
				line = fields[1]
			} else {
				line = fields[0]
			}

			if !cache.Exists(line) && !whitelist[line] {
				cache.Set(line, r.NewBlockedRecord())
				stats.AddBlockedDomain()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning hostfile: %s", err)
	}

	return nil
}

// PerformUpdate updates the block cache by building a new one and swapping
// it for the old cache.
func PerformUpdate(config *conf.BlockerConfig, cache c.Cache, forceUpdate bool) {
	if err := update(config, cache, forceUpdate); err != nil {
		logger.Fatal(err)
	}

	if err := updateBlockCache(cache, config.SourceDir); err != nil {
		logger.Fatal(err)
	}
}
