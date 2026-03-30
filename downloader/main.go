package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/injoyai/lorca"
)

// ProxyConfig matches the JS object
type ProxyConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Type    string `json:"type"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
}

type DownloadManager struct {
	ui          lorca.UI
	cancelFunc  context.CancelFunc
	mu          sync.Mutex
	downloading bool
}

func (dm *DownloadManager) Start(urlStr string, proxyConfigJSON string) {
	dm.mu.Lock()
	if dm.downloading {
		dm.mu.Unlock()
		dm.ui.Eval(`window.downloader.addLog("Task is already running", "warning")`)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	dm.cancelFunc = cancel
	dm.downloading = true
	dm.mu.Unlock()

	go dm.downloadRoutine(ctx, urlStr, proxyConfigJSON)
}

func (dm *DownloadManager) Stop() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dm.cancelFunc != nil {
		dm.cancelFunc() // This will trigger ctx.Done() in downloadRoutine
		dm.cancelFunc = nil
	}
	// The routine will handle the cleanup and UI update
}

func (dm *DownloadManager) downloadRoutine(ctx context.Context, urlStr string, proxyConfigJSON string) {
	defer func() {
		dm.mu.Lock()
		dm.downloading = false
		dm.cancelFunc = nil
		dm.mu.Unlock()
		// If we finished successfully, completeDownload() is called inside loop.
		// If we panic or return early, we might want to ensure state is reset.
		// But let's handle specific cases below.
	}()

	dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Starting download: %s", "info")`, urlStr))

	// Parse proxy
	var proxyConfig ProxyConfig
	json.Unmarshal([]byte(proxyConfigJSON), &proxyConfig)

	// Setup client
	transport := &http.Transport{}
	if proxyConfig.Enabled {
		proxyURLStr := fmt.Sprintf("%s://%s:%d", proxyConfig.Type, proxyConfig.Host, proxyConfig.Port)
		if proxyConfig.User != "" {
			// Basic handling, might need more specific URL construction for auth
			proxyURLStr = fmt.Sprintf("%s://%s:%s@%s:%d", proxyConfig.Type, proxyConfig.User, proxyConfig.Pass, proxyConfig.Host, proxyConfig.Port)
		}
		proxyURL, err := url.Parse(proxyURLStr)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
			dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Using proxy: %s", "info")`, proxyURLStr))
		} else {
			dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Invalid proxy URL: %s", "error")`, err.Error()))
		}
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   0, // No timeout for large downloads
	}

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Error creating request: %s", "error")`, err.Error()))
		dm.ui.Eval(`window.downloader.updateStatusIndicator("error")`)
		dm.ui.Eval(`window.downloader.stopDownload()`) // Reset UI state
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			dm.ui.Eval(`window.downloader.addLog("Download cancelled", "warning")`)
		} else {
			dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Connection failed: %s", "error")`, err.Error()))
			dm.ui.Eval(`window.downloader.updateStatusIndicator("error")`)
		}
		dm.ui.Eval(`window.downloader.stopDownload()`) // Reset UI state
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Server returned status: %s", "error")`, resp.Status))
		dm.ui.Eval(`window.downloader.updateStatusIndicator("error")`)
		dm.ui.Eval(`window.downloader.stopDownload()`)
		return
	}

	// Get filename
	filename := filepath.Base(resp.Request.URL.Path)
	if filename == "" || filename == "." || filename == "/" {
		filename = "downloaded_file"
	}
	// Check Content-Disposition header for filename
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if fname, ok := params["filename"]; ok {
				filename = fname
			}
		}
	}

	// Ensure unique filename
	filename = ensureUniqueFilename(filename)

	out, err := os.Create(filename)
	if err != nil {
		dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Cannot create file: %s", "error")`, err.Error()))
		dm.ui.Eval(`window.downloader.stopDownload()`)
		return
	}
	defer out.Close()

	dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Saving to: %s", "info")`, filename))
	dm.ui.Eval(`window.downloader.updateStatusIndicator("downloading")`)

	contentLength := resp.ContentLength
	buf := make([]byte, 32*1024)
	var downloaded int64
	startTime := time.Now()
	lastUpdate := time.Now()

	for {
		select {
		case <-ctx.Done():
			// Cancelled
			dm.ui.Eval(`window.downloader.addLog("Download stopped by user", "warning")`)
			// Ideally delete partial file
			out.Close()
			os.Remove(filename)
			return // stopDownload() in JS sets UI state
		default:
			nr, er := resp.Body.Read(buf)
			if nr > 0 {
				nw, ew := out.Write(buf[0:nr])
				if ew != nil {
					dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Write error: %s", "error")`, ew.Error()))
					dm.ui.Eval(`window.downloader.stopDownload()`)
					return
				}
				downloaded += int64(nw)
			}
			if er != nil {
				if er == io.EOF {
					// Success
					dm.ui.Eval(`window.downloader.updateProgress(100)`)
					// completeDownload in JS handles the final state
				} else {
					dm.ui.Eval(fmt.Sprintf(`window.downloader.addLog("Read error: %s", "error")`, er.Error()))
					dm.ui.Eval(`window.downloader.stopDownload()`)
				}
				return
			}

			// Update UI
			if time.Since(lastUpdate) > 200*time.Millisecond {
				progress := 0.0
				if contentLength > 0 {
					progress = float64(downloaded) / float64(contentLength) * 100
				}

				duration := time.Since(startTime).Seconds()
				speed := float64(downloaded) / duration // bytes/s
				speedMB := speed / 1024 / 1024

				remainingStr := "--"
				if speed > 0 && contentLength > 0 {
					remaining := float64(contentLength - downloaded)
					seconds := remaining / speed
					remainingStr = fmt.Sprintf("%.0f s", seconds)
				}

				// Safely format for JS
				jsCode := fmt.Sprintf(`
					window.downloader.updateProgress(%f);
					window.downloader.updateDownloadInfo("%.2f MB", "%.2f MB/s", "%s");
				`, progress, float64(downloaded)/1024/1024, speedMB, remainingStr)

				dm.ui.Eval(jsCode)
				lastUpdate = time.Now()
			}
		}
	}
}

func ensureUniqueFilename(filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	newFilename := filename
	counter := 1
	for {
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			return newFilename
		}
		newFilename = fmt.Sprintf("%s(%d)%s", name, counter, ext)
		counter++
	}
}

func main() {
	// Create UI with data URI or local file
	// Using local file path is easier for development
	cwd, _ := os.Getwd()
	htmlPath := filepath.Join(cwd, "index.html")

	ui, err := lorca.New("file:///"+htmlPath, "", 800, 600)
	if err != nil {
		fmt.Println("Error launching UI:", err)
		return
	}
	defer ui.Close()

	dm := &DownloadManager{ui: ui}

	// Bind Go functions to JS
	ui.Bind("goStartDownload", dm.Start)
	ui.Bind("goStopDownload", dm.Stop)

	// Wait until UI is closed
	<-ui.Done()
}
