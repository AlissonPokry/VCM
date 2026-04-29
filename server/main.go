package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/reelqueue-server/config"
	appdb "github.com/your-org/reelqueue-server/db"
	"github.com/your-org/reelqueue-server/handlers"
	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/services"
)

func main() {
	if err := run(); err != nil {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Printf("WARNING: ffmpeg not found on PATH; thumbnail processing disabled")
		cfg.FFmpegEnabled = false
	} else {
		cfg.FFmpegEnabled = true
	}

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return fmt.Errorf("create upload directory: %w", err)
	}
	if err := os.MkdirAll(cfg.ThumbnailDir, 0o755); err != nil {
		return fmt.Errorf("create thumbnail directory: %w", err)
	}

	conn, err := appdb.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	migrationsDir := filepath.Join(cfg.ProjectRoot, "server", "db", "migrations")
	if err := appdb.RunMigrations(conn, migrationsDir); err != nil {
		return err
	}
	if hasArg("--migrate-only") {
		log.Printf("migrations complete")
		return nil
	}

	router := buildRouter(cfg, conn)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("V.C.M Go API listening on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func buildRouter(cfg config.Config, conn *sql.DB) http.Handler {
	activityService := services.NewActivityService(conn)
	tagService := services.NewTagService(conn)
	videoService := services.NewVideoService(conn, tagService, activityService, cfg.UploadDir, cfg.ProjectRoot, cfg.DispatchTimeout())
	thumbnailService := services.NewThumbnailService(conn, cfg.ThumbnailDir, cfg.ProjectRoot, cfg.FFmpegEnabled)
	analyticsService := services.NewAnalyticsService(conn)

	videoHandler := handlers.NewVideoHandler(cfg, videoService, thumbnailService)
	tagHandler := handlers.NewTagHandler(tagService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	activityHandler := handlers.NewActivityHandler(activityService)
	n8nHandler := handlers.NewN8NHandler(videoService)

	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(appmw.CORS(cfg.CORSOrigin))

	r.Get("/thumbnails/{filename}", serveNamedFile(cfg.ThumbnailDir))
	r.Get("/uploads/{filename}", serveNamedFile(cfg.UploadDir))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		appmw.Respond(w, http.StatusOK, map[string]string{"status": "ok"}, map[string]any{})
	})
	r.Get("/api/health/n8n", func(w http.ResponseWriter, r *http.Request) {
		appmw.Respond(w, http.StatusOK, map[string]bool{"reachable": true, "protected": true}, map[string]any{})
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/videos", videoHandler.Routes)
		r.Route("/tags", tagHandler.Routes)
		r.Route("/analytics", analyticsHandler.Routes)
		r.Route("/activity", activityHandler.Routes)
		r.Route("/n8n", func(r chi.Router) {
			r.Use(appmw.N8NAuth(cfg.N8NWebhookSecret))
			n8nHandler.Routes(r)
		})
	})

	return r
}

func serveNamedFile(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := filepath.Base(chi.URLParam(r, "filename"))
		path := filepath.Join(dir, filename)
		if !withinDir(dir, path) {
			appmw.RespondError(w, http.StatusNotFound, "File not found", "FILE_NOT_FOUND")
			return
		}

		file, err := os.Open(path)
		if err != nil {
			appmw.RespondError(w, http.StatusNotFound, "File not found", "FILE_NOT_FOUND")
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			appmw.RespondError(w, http.StatusNotFound, "File not found", "FILE_NOT_FOUND")
			return
		}

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
	}
}

func withinDir(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func hasArg(arg string) bool {
	for _, value := range os.Args[1:] {
		if value == arg {
			return true
		}
	}
	return false
}
