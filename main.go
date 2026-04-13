package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"math/rand"

	"github.com/skip2/go-qrcode"
)

type ProgressWriter struct {
	w       http.ResponseWriter
	written int64
	total   int64
}

func (pw *ProgressWriter) Header() http.Header         { return pw.w.Header() }
func (pw *ProgressWriter) WriteHeader(statusCode int) { pw.w.WriteHeader(statusCode) }
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.written += int64(n)
	if pw.total > 0 {
		percent := float64(pw.written) / float64(pw.total) * 100
		fmt.Printf("\r       🚀 Transferring: %.2f%% (%d/%d bytes)", percent, pw.written, pw.total)
	}
	return n, err
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./transfer <file1> <file2>")
		return
	}

	targets := os.Args[1:]
	isSingleFile := len(targets) == 1
	var totalSize int64
	var firstFilePath string

	for _, t := range targets {
		info, err := os.Stat(t)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if info.IsDir() {
			isSingleFile = false
			filepath.Walk(t, func(_ string, f os.FileInfo, _ error) error {
				if !f.IsDir() { totalSize += f.Size() }
				return nil
			})
		} else {
			totalSize += info.Size()
			firstFilePath = t
		}
	}

	ip := getLocalIP()
	
	// FIX 1: Randomize Port so the browser doesn't reuse the connection
	rand.Seed(time.Now().UnixNano())
	port := fmt.Sprintf("%d", rand.Intn(1000) + 8000)
	
	// FIX 2: Random string in URL to bypass browser cache
	token := fmt.Sprintf("%d", time.Now().Unix())
	addr := fmt.Sprintf("http://%s:%s/download/%s", ip, port, token)

	q, _ := qrcode.New(addr, qrcode.Low)
	fmt.Println(q.ToSmallString(false))
	fmt.Printf("\n🔗 Scan to download: %s\n", addr)
	
	sizeMode := "KB"
	displaySize := float64(totalSize) / 1024
	if displaySize > 1024 {
		displaySize /= 1024
		sizeMode = "MB"
	}
	fmt.Printf("📊 Total Size: %.2f %s\n\n", displaySize, sizeMode)

	// Use a pattern to match the random token
	http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		// FIX 3: Strict Anti-Cache Headers
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		pw := &ProgressWriter{w: w, total: totalSize}

		if isSingleFile {
			file, _ := os.Open(firstFilePath)
			defer file.Close()
			info, _ := file.Stat()
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(firstFilePath)))
			http.ServeContent(pw, r, info.Name(), info.ModTime(), file)
		} else {
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", "attachment; filename=\"transfer.zip\"")
			zw := zip.NewWriter(pw)
			for _, target := range targets {
				filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
					if err != nil || info.IsDir() { return nil }
					header, _ := zip.FileInfoHeader(info)
					header.Name, _ = filepath.Rel(filepath.Dir(target), path)
					header.Method = zip.Deflate
					writer, _ := zw.CreateHeader(header)
					file, _ := os.Open(path)
					defer file.Close()
					io.Copy(writer, file)
					return nil
				})
			}
			zw.Close()
		}

		fmt.Println("\n\n✅ Transfer Complete!")
		go func() { time.Sleep(1 * time.Second); os.Exit(0) }()
	})

	http.ListenAndServe(":"+port, nil)
}
