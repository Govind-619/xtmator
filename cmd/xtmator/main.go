package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

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
		log.Printf("Could not open browser automatically: %v", err)
	}
}

func main() {
	// ── 1. Database ──────────────────────────────────────────────────────────────
	db, err := repository.NewSQLiteDB("xtmator.db")
	if err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer db.Close()

	// ── 2. Repositories ──────────────────────────────────────────────────────────
	userRepo    := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	dsrRepo     := repository.NewDSRRepository(db)
	boqRepo     := repository.NewBOQRepository(db)

	// ── 3. Usecases ──────────────────────────────────────────────────────────────
	authUC    := usecase.NewAuthUsecase(userRepo)
	projectUC := usecase.NewProjectUsecase(projectRepo)
	boqUC     := usecase.NewBOQUsecase(boqRepo, dsrRepo, projectRepo)

	// ── 4. Handlers ──────────────────────────────────────────────────────────────
	webH    := handler.NewWebHandler()
	authH   := handler.NewAuthHandler(authUC)
	projH   := handler.NewProjectHandler(projectUC, authUC)
	dsrH    := handler.NewDSRHandler(dsrRepo)
	boqH    := handler.NewBOQHandler(boqUC, authUC)
	exportH := handler.NewExportHandler(boqUC, authUC)

	// ── 5. Router ────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// SPA — serve React app for all non-API routes
	mux.HandleFunc("/", webH.Home)

	// Auth (public)
	mux.HandleFunc("/api/auth/register", authH.Register)
	mux.HandleFunc("/api/auth/login",    authH.Login)

	// DSR catalogue (protected)
	mux.HandleFunc("/api/dsr/categories", handler.JWTAuth(authUC, dsrH.Categories))
	mux.HandleFunc("/api/dsr/items",      handler.JWTAuth(authUC, dsrH.Items))

	// Projects (protected)
	mux.HandleFunc("/api/projects",    handler.JWTAuth(authUC, projH.HandleProjects))
	mux.HandleFunc("/api/projects/",   handler.JWTAuth(authUC, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case isExportPath(path):
			exportH.ExportPDF(w, r)
		case isBOQEntryPath(path):
			boqH.HandleBOQEntry(w, r)
		case isBOQPath(path):
			boqH.HandleBOQ(w, r)
		default:
			projH.HandleProject(w, r)
		}
	}))

	// ── 6. Server ────────────────────────────────────────────────────────────────
	port := ":3333"
	url  := "http://localhost" + port
	fmt.Printf("🏗  Xtmator started → %s\n", url)

	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal("Server error: ", err)
	}
}

// Path matchers for the /api/projects/* multiplex handler
func isExportPath(p string) bool {
	// /api/projects/42/export/pdf
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' { return p[i+1:] == "pdf"; }
	}
	return false
}
func isBOQEntryPath(p string) bool {
	// /api/projects/42/boq/7
	parts := splitPath(p)
	return len(parts) == 5 && parts[3] == "boq"
}
func isBOQPath(p string) bool {
	// /api/projects/42/boq
	parts := splitPath(p)
	return len(parts) >= 4 && parts[3] == "boq"
}
func splitPath(p string) []string {
	var out []string
	cur := ""
	for _, c := range p {
		if c == '/' {
			if cur != "" { out = append(out, cur) }
			cur = ""
		} else {
			cur += string(c)
		}
	}
	if cur != "" { out = append(out, cur) }
	return out
}
