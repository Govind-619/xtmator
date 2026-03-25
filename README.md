# XTMATOR 🏗️

**A professional, lightning-fast Construction Estimation and Standard Bill of Quantities (BOQ) System.**

XTMATOR bridges the gap between massive construction databases like the Delhi Schedule of Rates (DSR) and modern, fluid project estimation. Built natively with Go and React, it compiles entirely into a single cross-platform executable file running a localized, ultra-fast persistence engine.

## Features ✨

* **Multi-Sheet Workspace Partitioning:** Divide massive multi-million dollar projects cleanly into customizable spreadsheets (e.g., Civil, Mechanical, Plumbing) tightly isolated within a single parent Project constraint.
* **Instant Native Exporters:** Generate professional A4 PDF reports (powered by `maroto`) and exact tabular Excel `.xlsx` spreadsheets (powered by `excelize`) automatically grouping your layouts contextually by category.
* **Universal Cost Indexing:** Handle rampant material inflation directly from the application header by scaling base DSR rates automatically across thousands of items using the Global Cost Index toggles.
* **Client Portal & Share Links:** Generate cryptic AES-encrypted public permalinks routing untrusted clients to secure, read-only headless web views of real-time BOQ values with 1 click.
* **Proprietary Templates & Rate Overrides:** Easily drop down of DSR boundaries, inline edit raw parameters directly within the spreadsheet cell grid, and save those overrides permanently to a personal custom items library.
* **Extremely Portable:** Absolutely zero Node.js setups, NPM run commands, or bulky PostgreSQL containers required. Simply run the system.

## Stack Architecture 🛠️

XTMATOR leverages one of the leanest web architectures available today:
- **Backend:** Go (Golang) standard library HTTP multiplexer securely mapped onto raw REST endpoints.
- **Persistence:** High-concurrency `database/sql` SQLite3 driver.
- **Frontend Engine:** Embedded React SPA mapped syntactically through standalone Babel compilers natively injecting UI directly via `go:embed` binaries.

## Getting Started 🚀

1. Ensure [Go 1.20+](https://go.dev/dl/) is installed. 
2. Clone the repository and boot the server:
```bash
go run ./cmd/xtmator/main.go
```
3. The server natively injects itself over `localhost` at port `3333` and executes your Default OS Browser! 
4. Import your Excel DSR schedules into the embedded storage and begin estimating seamlessly!

## 🚀 Sharing & Deployment

XTMATOR is designed for extreme portability. You can share it with friends or clients who have zero IT knowledge using one of two methods:

### Method 1: The Single-File Binary (Easiest)
Since XTMATOR is written in Go, it compiles into a single executable file that includes everything (including the DSR catalog).

1. **Build for Windows:**
   - If you have Go installed on Windows:
     ```bash
     CGO_ENABLED=1 go build -o xtmator.exe ./cmd/xtmator/main.go
     ```
   - If you are on Linux and want to build for Windows:
     ```bash
     CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o xtmator.exe ./cmd/xtmator/main.go
     ```
2. **Share:** Send the `xtmator.exe` to your friend.
3. **Run:** They just need to double-click it. It will automatically create `xtmator.db` and populate it with the 3,374 DSR items on its own.
4. **Access:** Open `http://localhost:3333` in any browser.

### Method 2: Docker (Best for Reliability)
If they have Docker installed, they can run the entire environment with one command:

```bash
docker-compose up -d
```
The app will be available at `http://localhost:3333`. All data is persisted in a Docker volume.

## 📜 License
MIT License. Created for the construction industry.
