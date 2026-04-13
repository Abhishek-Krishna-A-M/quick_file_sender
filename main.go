package main

import (
    _ "embed"
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/skip2/go-qrcode"
)

// ProgressWriter tracks bytes written to the ResponseWriter for the CLI progress bar
type ProgressWriter struct {
	w       http.ResponseWriter
	written int64
	total   int64
}

//go:embed upload.html
var uploadHTML string

func (pw *ProgressWriter) Header() http.Header         { return pw.w.Header() }
func (pw *ProgressWriter) WriteHeader(statusCode int) { pw.w.WriteHeader(statusCode) }
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.written += int64(n)
	if pw.total > 0 {
		percent := float64(pw.written) / float64(pw.total) * 100
		fmt.Printf("\r        Transferring: %.2f%% (%d/%d bytes)", percent, pw.written, pw.total)
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

var uploadDir string

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprint(w, uploadHTML)
		return
	}

	// Max 32MB in RAM, then it uses temp files on disk
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Parse Error", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["myFiles"]
	fmt.Printf("\nReceiving %d file(s) into: %s\n", len(files), uploadDir)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		fullPath := filepath.Join(uploadDir, fileHeader.Filename)
		dst, err := os.Create(fullPath)
		if err != nil {
			fmt.Printf("Error creating %s: %v\n", fullPath, err)
			continue
		}
		defer dst.Close()

		io.Copy(dst, file)
		fmt.Printf("  ✅ Received: %s\n", fullPath)
	}

	fmt.Fprintf(w, "Successfully uploaded %d file(s) to %s", len(files), uploadDir)

	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
}

func main() {
	receiveMode := flag.Bool("r", false, "Receive mode (receive file from phone)")
	flag.Parse()

	ip := getLocalIP()
	rand.Seed(time.Now().UnixNano())
	port := fmt.Sprintf("%d", rand.Intn(1000)+8000)
	token := fmt.Sprintf("%d", time.Now().Unix())

	if *receiveMode {
		// --- RECEIVE MODE ---
		uploadDir = "."
		if len(flag.Args()) > 0 {
			uploadDir = flag.Args()[0]
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		addr := fmt.Sprintf("http://%s:%s/upload", ip, port)
		q, _ := qrcode.New(addr, qrcode.Low)
		fmt.Println(q.ToSmallString(false))
		fmt.Printf("\nScan to UPLOAD to: %s\nURL: %s\n", uploadDir, addr)

		http.HandleFunc("/upload", handleUpload)
		http.ListenAndServe(":"+port, nil)

	} else {
		// --- SEND MODE ---
		if len(flag.Args()) < 1 {
			fmt.Println("Usage:\n  Send:    qfs <file1> <dir1>\n  Receive: qfs -r [target_dir]")
			return
		}

		targets := flag.Args()
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
					if !f.IsDir() {
						totalSize += f.Size()
					}
					return nil
				})
			} else {
				totalSize += info.Size()
				firstFilePath = t
			}
		}

		addr := fmt.Sprintf("http://%s:%s/download/%s", ip, port, token)
		q, _ := qrcode.New(addr, qrcode.Low)
		fmt.Println(q.ToSmallString(false))
		fmt.Printf("\nScan to DOWNLOAD from terminal: %s\n", addr)

		sizeMode := "KB"
		displaySize := float64(totalSize) / 1024
		if displaySize > 1024 {
			displaySize /= 1024
			sizeMode = "MB"
		}
		fmt.Printf("Total Size: %.2f %s\n\n", displaySize, sizeMode)

		http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
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
						if err != nil || info.IsDir() {
							return nil
						}
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

			fmt.Println("\n\nTransfer Complete!")
			go func() { time.Sleep(1 * time.Second); os.Exit(0) }()
		})

		http.ListenAndServe(":"+port, nil)
	}
}
