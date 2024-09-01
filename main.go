package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

const (
	cacheDir = "./cache"
	port     = 8080
)

func init() {
	image.RegisterFormat("webp", "RIFF????WEBPVP8", webp.Decode, webp.DecodeConfig)
}

func main() {
	// Parse CLI arguments
	urlFlag := flag.String("url", "", "URL of the image to resize")
	formatFlag := flag.String("f", "jpg", "Output format (jpg, png, webp)")
	widthFlag := flag.Int("w", 0, "Desired width")
	heightFlag := flag.Int("h", 0, "Desired height")
	serverFlag := flag.Bool("server", false, "Run as HTTP server")
	flag.Parse()

	if *serverFlag || *urlFlag == "" {
		log.Println("Running as HTTP server")
		http.HandleFunc("/", handleRequest)
		log.Printf("Server starting on port %d\n", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	} else if *urlFlag != "" && *widthFlag > 0 && *heightFlag > 0 {
		// Run as CLI tool
		err := resizeAndSaveImage(*urlFlag, *formatFlag, *widthFlag, *heightFlag)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		log.Println("Image resized and saved successfully")
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) == 3 && parts[1] == "cache" {
		fn := path.Clean(parts[2])
		log.Printf("cache: \033[30m%s\033[0m", fn)
		if _, err := os.Stat(filepath.Join(cacheDir, fn)); err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, filepath.Join(cacheDir, fn))
		return
	}

	if len(parts) < 5 {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	format := strings.TrimPrefix(parts[1], "format:")
	resizeParams := strings.Split(strings.TrimPrefix(parts[2], "resize:"), ":")
	if len(resizeParams) != 3 {
		http.Error(w, "Invalid resize parameters", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))

	width, _ := strconv.Atoi(resizeParams[1])
	height, _ := strconv.Atoi(resizeParams[2])
	url, err := url.QueryUnescape(strings.Join(parts[4:], "/"))
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	err = resizeAndSaveImage(url, format, width, height)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	cacheKey := generateCacheKey(url, format, width, height)
	cachedPath := filepath.Join(cacheDir, cacheKey)

	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))
	http.ServeFile(w, r, cachedPath)
}

func resizeAndSaveImage(imageURL, format string, width, height int) error {
	cacheKey := generateCacheKey(imageURL, format, width, height)
	cachedPath := filepath.Join(cacheDir, cacheKey)

	if _, err := os.Stat(cachedPath); err == nil {
		// Image already cached
		log.Printf("Cache hit \033[32m%s\033[0m\n", cacheKey)
		return nil
	}

	// Download image
	log.Printf("Downloading \033[33m%s\033[0m\n", imageURL)
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	// Decode image
	img, err := imaging.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize image
	resizedImg := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	// Ensure cache directory exists
	os.MkdirAll(cacheDir, os.ModePerm)

	// Save resized image to cache
	out, err := os.Create(cachedPath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %v", err)
	}
	defer out.Close()

	// Encode and save the image
	switch format {
	case "jpg", "jpeg":
		err = imaging.Encode(out, resizedImg, imaging.JPEG, imaging.JPEGQuality(85))
	case "png":
		err = imaging.Encode(out, resizedImg, imaging.PNG)
	case "webp":
		err = webp.Encode(out, resizedImg, &webp.Options{Lossless: false, Quality: 80})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}

func generateCacheKey(url, format string, width, height int) string {
	key := url
	url_hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x_%dx%d.%s", url_hash, width, height, format)
}
