# Quick File Sender (QFS)

Quick File Sender is a high-performance CLI tool written in Go that allows you to beam files and folders directly from your terminal to your mobile phone — or receive files from your phone back to your terminal — using a QR code.

No apps, no accounts, and no cloud uploads—just pure local Wi-Fi speed.

---

## ✨ Features

- **Zero Configuration:** Automatically detects your local IP and sets up a temporary server.
- **Zero Mobile Setup:** No app required on the receiving end; uses the native mobile browser.
- **Two-Way Transfer:** Send files to your phone *or* receive files from your phone — with a single flag.
- **High Speed:** Uses Go's `http.ServeContent` for zero-copy data transfer and optimized throughput.
- **Smart Streaming:**
  - *Single Files:* Served raw for immediate use (e.g., PDFs, APKs, Images).
  - *Multiple Files/Folders:* Zipped on-the-fly and streamed—no waiting for compression to finish.
- **Browser-as-Remote:** In receive mode, QFS serves a lightweight HTML upload UI to your phone's browser — turning it into a temporary file uploader with no app install.
- **Anti-Cache:** Randomized ports and cache-busting headers ensure you never download an old version of a file.
- **Progress Tracking:** Real-time transfer percentage and speed shown directly in your CLI.

---

## 🚀 Installation

### From Source

Ensure you have Go installed:

```bash
git clone https://github.com/Abhishek-Krishna-A-M/quick_file_sender.git
cd quick-file-sender
go mod tidy
go build -ldflags="-s -w" -o qfs main.go
```
---

## 🛠 Usage

### Send a file to your phone

```bash
qfs document.pdf
```

### Send multiple files or entire folders

```bash
qfs photo1.jpg photo2.png ./my_folder
```

### Receive files from your phone

```bash
qfs -r /path/to/save/
```

1. The CLI generates a **QR Code** in your terminal.
2. Scan it with your phone's camera.
3. In **send mode** — tap the link to download at maximum Wi-Fi speed.
4. In **receive mode** — your phone's browser opens an upload UI; select files and send them straight to your terminal.
5. The server automatically shuts down once the transfer is complete.

---

## 🏗 How it Works

Quick File Sender creates a temporary HTTP server on your local machine.

- **Discovery:** Finds your local network IP (e.g., `192.168.1.70`) automatically.
- **Security:** Generates a unique, one-time URL on a random port to prevent caching and port conflicts.
- **Send Mode (GET):** Your phone makes a `GET` request. The terminal pushes bytes out — single files served raw, multiple files/folders streamed as a live ZIP archive via `archive/zip`.
- **Receive Mode (POST):** The `-r` flag switches the server to accept `POST` requests. The terminal pulls bytes from the socket using `r.ParseMultipartForm`, which untangles a multipart stream of files sent from the browser. Files are saved using `os.MkdirAll` and `os.Create` into the specified directory.
- **Bridge UI:** In receive mode, QFS serves a small embedded HTML page (`uploadHTML`) to the phone's browser — turning it into a temporary upload remote with no install required.
- **Cleanup:** Server shuts itself down cleanly after the transfer completes.

---

## 📊 One-Way vs Two-Way

| Feature         | One-Way (Send)         | Two-Way (Receive)              |
|-----------------|------------------------|-------------------------------|
| Command         | `qfs file.txt`         | `qfs -r /path/`               |
| Terminal Role   | File Provider          | File Sink / Storage            |
| Phone Role      | Downloader             | Uploader                       |
| Data Flow       | Disk → Zip → Network   | Network → Buffer → Disk        |
| HTML UI Needed? | No (Direct Download)   | Yes (Upload button in browser) |

---

## 📝 License
MIT License. Free to use, modify, and share.
