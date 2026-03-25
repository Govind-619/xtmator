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

How to Download, Set Up, and Run XTMATOR (Windows)
--------------------------------------------------

Follow these simple steps to get the application running on your computer.

> **Prerequisite:** Make sure you have [Go](https://go.dev/dl/ "null") installed on your system before starting.

### Step 1: Download the Code

1.  Go to the XTMATOR GitHub page.

2.  Click on the green **"<> Code"** button near the top right.

3.  Click **"Download ZIP"** from the dropdown menu.

### Step 2: Extract the Folder

1.  Locate the downloaded `.zip` file on your computer (usually in your `Downloads` folder).

2.  Right-click the file and select **"Extract All..."**.

3.  Follow the prompts to extract it. This will create a normal, unzipped folder containing the code.

### Step 3: Open the Command Window

1.  Open that newly extracted folder so you can see all the files inside.

2.  Click once in the **address bar** at the very top of the folder window (where it shows the folder path, like `C:\Users\Downloads\xtmator-main`).

3.  Delete the text in the address bar, type `cmd`, and press **Enter**.

4.  A black command prompt window will pop up.

### Step 4: Create the Application

1.  Copy the following command exactly as it is:

    ```
    CGO_ENABLED=1 go build -o xtmator.exe ./cmd/xtmator/main.go

    ```

2.  Paste it into the black command window and press **Enter**.

3.  Wait a few seconds. It might look like nothing is happening, but it is building the app behind the scenes.

### Step 5: Run the App

1.  Look back at your open folder. You should now see a brand new file named **`xtmator.exe`**.

2.  Double-click `xtmator.exe` to launch it.

3.  **⚠️ CRITICAL:** A black window will open and stay open. **Do not close this window!** As long as you are using the app, this window needs to remain open in the background.

### Step 6: Access the App

1.  Open your favorite web browser (like Chrome, Edge, or Firefox).

2.  Type the following into the web address bar at the top:

    ```
    http://localhost:3333

    ```

3.  Press **Enter**. The app is now ready to use! It has automatically set up its database and loaded all the necessary items for you.

### Method 2: Docker (Best for Reliability)
If they have Docker installed, they can run the entire environment with one command:

```bash
docker-compose up -d
```
The app will be available at `http://localhost:3333`. All data is persisted in a Docker volume.

## 📜 License
MIT License. Created for the construction industry.
