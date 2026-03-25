package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/joho/godotenv"

	"github.com/Govind-619/xtmator/handler"
	"github.com/Govind-619/xtmator/repository"
	"github.com/Govind-619/xtmator/usecase"
)

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}

func main() {
	// ── 0. Load environment ───────────────────────────────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using system environment variables")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "xtmator.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	// ── 1. Database ───────────────────────────────────────────────────────────────
	db, err := repository.NewSQLiteDB(dbPath)
	if err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer db.Close()

	// ── 2. Repositories ───────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	dsrRepo := repository.NewDSRRepository(db)
	boqRepo := repository.NewBOQRepository(db)
	projectSheetRepo := repository.NewProjectSheetRepository(db)
	customItemRepo := repository.NewCustomItemRepository(db)

	// ── 3. Usecases ───────────────────────────────────────────────────────────────
	authUC := usecase.NewAuthUsecase(userRepo)
	googleAuthUC := usecase.NewGoogleOAuthUsecase(userRepo, authUC)
	projectUC := usecase.NewProjectUsecase(projectRepo)
	boqUC := usecase.NewBOQUsecase(boqRepo, dsrRepo, projectRepo, projectSheetRepo)

	customItemUC := usecase.NewCustomItemUsecase(customItemRepo)

	// ── 4. Handlers ───────────────────────────────────────────────────────────────
	webH := handler.NewWebHandler()
	authH := handler.NewAuthHandler(authUC)
	googleH := handler.NewGoogleAuthHandler(googleAuthUC)
	projH := handler.NewProjectHandler(projectUC, authUC)
	projSheetH := handler.NewProjectSheetHandler(projectSheetRepo)
	dsrH := handler.NewDSRHandler(dsrRepo)
	boqH := handler.NewBOQHandler(boqUC, authUC)
	exportH := handler.NewExportHandler(boqUC, authUC)
	shareH := handler.NewShareHandler(boqUC)

	customItemH := handler.NewCustomItemHandler(customItemUC, authUC)

	// ── 5. Rate limiters ──────────────────────────────────────────────────────────
	authLimiter := handler.NewRateLimiter(10, time.Minute) // 10 req/min on auth
	apiLimiter := handler.NewRateLimiter(120, time.Minute) // 120 req/min on API

	// ── 6. Router ─────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// SPA shell
	mux.HandleFunc("/", webH.Home)

	// Auth (rate limited, public)
	mux.HandleFunc("/api/auth/register", authLimiter.Middleware(authH.Register))
	mux.HandleFunc("/api/auth/login", authLimiter.Middleware(authH.Login))
	mux.HandleFunc("/api/auth/google", googleH.InitiateLogin)
	mux.HandleFunc("/api/auth/google/callback", googleH.Callback)

	// DSR catalogue (JWT + rate limited)
	mux.HandleFunc("/api/dsr/categories", apiLimiter.Middleware(handler.JWTAuth(authUC, dsrH.Categories)))
	mux.HandleFunc("/api/dsr/items", apiLimiter.Middleware(handler.JWTAuth(authUC, dsrH.Items)))

	// Custom Items (JWT + rate limited)
	mux.HandleFunc("/api/custom-items", apiLimiter.Middleware(handler.JWTAuth(authUC, customItemH.HandleItems)))
	mux.HandleFunc("/api/custom-items/", apiLimiter.Middleware(handler.JWTAuth(authUC, customItemH.HandleItem)))

	// Public share endpoint
	mux.HandleFunc("/api/share/", shareH.HandleSharedSheet)

	// Public assets
	mux.HandleFunc("/assets/logo_web.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(handler.LogoWebData)
	})

	// Projects + BOQ (JWT + rate limited)
	mux.HandleFunc("/api/projects", apiLimiter.Middleware(handler.JWTAuth(authUC, projH.HandleProjects)))
	mux.HandleFunc("/api/projects/", apiLimiter.Middleware(handler.JWTAuth(authUC, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case isGenerateSharePath(path):
			projH.HandleGenerateShare(w, r)
		case isExportPath(path):
			parts := splitPath(path)
			if parts[len(parts)-1] == "excel" {
				exportH.ExportExcel(w, r)
			} else {
				exportH.ExportPDF(w, r)
			}
		case isBOQEntryPath(path):
			boqH.HandleBOQEntry(w, r)
		case isBOQPath(path):
			boqH.HandleBOQ(w, r)
		case isProjectSheetPath(path):
			projSheetH.HandleSheet(w, r)
		case isProjectSheetsPath(path):
			projSheetH.HandleSheets(w, r)
		default:
			projH.HandleProject(w, r)
		}
	})))

	// ── 7. Security headers wrapper ───────────────────────────────────────────────
	secured := handler.SecureHeaders(mux)

	// ── 8. Start ──────────────────────────────────────────────────────────────────
	addr := ":" + port
	url := "http://localhost" + addr
	fmt.Printf("🏗  XTMATOR started → %s\n", url)

	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	if err := http.ListenAndServe(addr, secured); err != nil {
		log.Fatal("Server error: ", err)
	}
}

func isExportPath(p string) bool {
	parts := splitPath(p)
	return len(parts) >= 5 && (parts[len(parts)-1] == "pdf" || parts[len(parts)-1] == "excel")
}
func isBOQEntryPath(p string) bool {
	parts := splitPath(p)
	return len(parts) == 5 && parts[3] == "boq"
}
func isBOQPath(p string) bool {
	parts := splitPath(p)
	return len(parts) >= 4 && parts[3] == "boq"
}
func isProjectSheetPath(p string) bool {
	parts := splitPath(p)
	return len(parts) == 5 && parts[3] == "sheets"
}
func isProjectSheetsPath(p string) bool {
	parts := splitPath(p)
	return len(parts) >= 4 && parts[3] == "sheets"
}
func isGenerateSharePath(p string) bool {
	parts := splitPath(p)
	return len(parts) >= 4 && parts[3] == "share"
}
func splitPath(p string) []string {
	var out []string
	cur := ""
	for _, c := range p {
		if c == '/' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
