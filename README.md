# Quick File Sender (QFS)

Quick File Sender is a high-performance CLI tool written in Go that allows you to beam files and folders directly from your terminal to your mobile phone (or any device with a camera and a browser) using a QR code.

No apps, no accounts, and no cloud uploads—just pure local Wi-Fi speed.

---

## ✨ Features

- **Zero Configuration:** Automatically detects your local IP and sets up a temporary server.
- **Zero Mobile Setup:** No app required on the receiving end; uses the native mobile browser.
- **High Speed:** Uses Go's `http.ServeContent` for zero-copy data transfer and optimized throughput.
- **Smart Streaming:**
  - *Single Files:* Served raw for immediate use (e.g., PDFs, APKs, Images).
  - *Multiple Files/Folders:* Zipped on-the-fly and streamed—no waiting for compression to finish.
- **Anti-Cache:** Randomized ports and cache-busting headers ensure you never download an old version of a file.
- **Progress Tracking:** Real-time transfer percentage and speed shown directly in your CLI.

---

## 🚀 Installation

### From Source

Ensure you have Go installed:

```bash
git clone https://github.com/yourusername/quick_file_sender.git
cd quick-file-sender
go mod tidy
go build -ldflags="-s -w" -o qfs main.go

---

## 🛠 Usage

### Send a single file

```bash
qfs document.pdf
```

### Send multiple files or entire folders

```bash
qfs photo1.jpg photo2.png ./my_folder
```

1. The CLI will generate a **QR Code**.
2. Scan it with your phone's camera.
3. Tap the link to download at maximum Wi-Fi speed.
4. The server automatically shuts down once the transfer is complete.

---

## 🏗 How it Works

Quick File Sender creates a temporary HTTP server on your local machine.

- **Discovery:** It finds your local network IP (e.g., `192.168.1.70`).
- **Security:** It generates a unique, one-time URL and opens a random port to prevent browser caching and port conflicts.
- **Efficiency:** For multiple files, it uses `archive/zip` to stream a ZIP archive directly into the network socket. This means the phone starts downloading the first byte before the tool even knows the total size of the last file.

---

## 📝 License

MIT License. Free to use, modify, and share.
