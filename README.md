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
