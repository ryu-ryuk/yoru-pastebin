package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryu-ryuk/yoru/internal/config"
	"github.com/ryu-ryuk/yoru/internal/database"
	"github.com/ryu-ryuk/yoru/internal/paste"
	"github.com/ryu-ryuk/yoru/pkg/crypt"
	"github.com/ryu-ryuk/yoru/pkg/idgen"
	"github.com/ryu-ryuk/yoru/pkg/storage"
)

const (
	templatesDir  = "./web/templates/"
	staticDir     = "./web/static"
	maxUploadSize = 30 * 1024 * 1024
)

type PageData struct {
	CurrentYear int
	Message     string
	StatusCode  int
	Paste       *paste.Paste
	PasteID     string
	CurrentPort int
}

type Server struct {
	httpServer *http.Server
	config     *config.Config
	pasteRepo  paste.Repository
	templates  *template.Template
	storage    storage.Storage
	baseURL    string
}

func NewServer(cfg *config.Config, db *database.DB) *Server {
	repo := paste.NewPGRepository(db.Pool)
	tmpl := mustInitTemplates()

	// Initialize hybrid storage
	storageInstance, err := storage.NewHybridStorage(
		storage.S3Config{
			Bucket:          cfg.S3.Bucket,
			Region:          cfg.S3.Region,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
		},
		"./data/uploads",   // local storage directory
		cfg.Server.BaseURL, // base URL for constructing download links
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	log.Printf("Storage initialized: %s", storageInstance.GetStorageInfo())

	return &Server{
		config:    cfg,
		pasteRepo: repo,
		templates: tmpl,
		storage:   storageInstance,
		baseURL:   cfg.Server.BaseURL,
	}
}

func mustInitTemplates() *template.Template {
	funcMap := template.FuncMap{
		"FormatFileSize": FormatFileSize,
		"js":             template.JSEscapeString,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(templatesDir + "*.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	return tmpl
}

func (s *Server) Start() error {
	router := s.setupRoutes()
	handler := s.withSecurityHeaders(router)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("server live on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("shutting down server...")
	return s.httpServer.Shutdown(ctx)
}
func withStaticCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex())
	mux.HandleFunc("POST /", s.handleCreatePaste())

	mux.HandleFunc("GET /{id}", s.handleGetPasteRedirect())
	mux.HandleFunc("GET /{id}/", s.handleGetPaste())
	mux.HandleFunc("POST /{id}/", s.handleGetPaste())
	mux.HandleFunc("GET /file/{id}/download", s.handleFileDownload())

	mux.HandleFunc("POST /api/v1/pastes", s.apiCreatePaste())
	mux.HandleFunc("GET /api/v1/pastes/{id}", s.apiGetPaste())
	mux.HandleFunc("GET /health", s.handleHealth())

	// Static files only
	fs := http.FileServer(http.Dir(staticDir))
	cachedFS := withStaticCache(http.StripPrefix("/static/", fs))
	mux.Handle("GET /static/", cachedFS)

	return mux
}

func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", `
		default-src 'self';
		script-src 'self' 'unsafe-inline';
		style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://rsms.me;
		font-src 'self' https://fonts.gstatic.com https://rsms.me;
		`)

		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")

		if !strings.HasPrefix(r.URL.Path, "/static/") {
			// only disable caching on dynamic routes
			w.Header().Set("Cache-Control", "no-store")
		}

		log.Printf("[%s] %s %s", getRealIP(r), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func getRealIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.Split(xff, ",")[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func FormatFileSize(size int64) string {
	const (
		B  = 1
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func (s *Server) renderTemplate(w http.ResponseWriter, tmpl string, data PageData, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	err := s.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func apiRespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode error: %v", err)
		http.Error(w, `{"error":"internal json error"}`, http.StatusInternalServerError)
	}
}

func apiErrorJSON(w http.ResponseWriter, statusCode int, msg string) {
	apiRespondJSON(w, statusCode, map[string]string{"error": msg})
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.renderTemplate(w, "index.html", PageData{
			CurrentYear: time.Now().Year(),
			CurrentPort: s.config.Server.Port,
		}, http.StatusOK)
	}
}

// handles the submission of a new paste.
func (s *Server) handleCreatePaste() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Method Not Allowed",
				StatusCode:  http.StatusMethodNotAllowed,
				CurrentYear: time.Now().Year(),
			}, http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			s.renderTemplate(w, "error.html", PageData{
				Message:    "file too large (max 30 MB)",
				StatusCode: http.StatusRequestEntityTooLarge,
			}, http.StatusRequestEntityTooLarge)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		id, err := idgen.GenerateSecureID(s.config.Paste.IDLength)
		if err != nil {
			s.renderTemplate(w, "error.html", PageData{Message: "id gen error", StatusCode: 500}, 500)
			return
		}

		newPaste := &paste.Paste{
			ID:        id,
			CreatedAt: time.Now(),
		}

		content := r.FormValue("content")
		file, header, fileErr := r.FormFile("file")
		hasContent := content != ""
		hasFile := fileErr == nil

		log.Printf("DEBUG: content=%s, hasContent=%v, fileErr=%v, hasFile=%v", content, hasContent, fileErr, hasFile)
		if header != nil {
			log.Printf("DEBUG: filename=%s, size=%d", header.Filename, header.Size)
		}

		if !hasContent && !hasFile {
			s.renderTemplate(w, "error.html", PageData{
				Message:    "please provide content or file",
				StatusCode: http.StatusBadRequest,
			}, http.StatusBadRequest)
			return
		}

		if hasFile {
			defer file.Close()
			newPaste.IsFile = true
			newPaste.FileName = header.Filename
			newPaste.MimeType = header.Header.Get("Content-Type")
			newPaste.FileSize = header.Size

			if header.Size > maxUploadSize {
				s.renderTemplate(w, "error.html", PageData{
					Message:    "file too big",
					StatusCode: http.StatusRequestEntityTooLarge,
				}, http.StatusRequestEntityTooLarge)
				return
			}

			s3Key := fmt.Sprintf("uploads/%s/%s", id, header.Filename)
			newPaste.S3Key = &s3Key

			// Upload using the storage interface
			err := s.storage.Upload(ctx, s3Key, file, storage.FileInfo{
				Key:          s3Key,
				OriginalName: header.Filename,
				Size:         header.Size,
				ContentType:  newPaste.MimeType,
			})
			if err != nil {
				log.Printf("File upload failed: %v", err)
				s.renderTemplate(w, "error.html", PageData{Message: "upload failed", StatusCode: 500}, 500)
				return
			}
		} else {
			if len(content) > s.config.Paste.MaxContentSizeBytes {
				s.renderTemplate(w, "error.html", PageData{
					Message:    "paste too long",
					StatusCode: http.StatusRequestEntityTooLarge,
				}, http.StatusRequestEntityTooLarge)
				return
			}

			newPaste.Language = r.FormValue("language")
			if newPaste.Language == "" {
				newPaste.Language = "plaintext"
			}

			password := r.FormValue("password")
			if password != "" {
				hash, err := crypt.GenerateHash(password, s.config.Security.BcryptCost)
				if err != nil {
					s.renderTemplate(w, "error.html", PageData{Message: "password error", StatusCode: 500}, 500)
					return
				}
				newPaste.PasswordHash = &hash

				salt, err := crypt.GenerateSalt()
				if err != nil {
					s.renderTemplate(w, "error.html", PageData{Message: "salt gen error", StatusCode: 500}, 500)
					return
				}
				newPaste.Salt = salt

				key := crypt.DeriveKey([]byte(password), salt)
				encrypted, iv, err := crypt.Encrypt([]byte(content), key)
				if err != nil {
					s.renderTemplate(w, "error.html", PageData{Message: "encrypt error", StatusCode: 500}, 500)
					return
				}
				newPaste.Content = encrypted
				newPaste.EncryptedIV = iv
			} else {
				newPaste.Content = content
			}
		}

		exp := r.FormValue("expires_in_minutes")
		minutes := 0
		fmt.Sscanf(exp, "%d", &minutes)
		newPaste.ExpiresAt = s.config.Paste.GetExpirationTime(minutes)

		if err := s.pasteRepo.CreatePaste(ctx, newPaste); err != nil {
			s.renderTemplate(w, "error.html", PageData{Message: "db save error", StatusCode: 500}, 500)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/%s/", id), http.StatusFound)
	}
}

// redirects /{id} to /{id}/ for consistency
func (s *Server) handleGetPasteRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			s.renderTemplate(w, "error.html", PageData{
				Message:     "missing paste id",
				StatusCode:  http.StatusBadRequest,
				CurrentYear: time.Now().Year(),
			}, http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/%s/", id), http.StatusMovedPermanently)
	}
}

// retrieves and displays a paste.
func (s *Server) handleGetPaste() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			s.renderTemplate(w, "error.html", PageData{
				Message:     "missing paste id",
				StatusCode:  http.StatusBadRequest,
				CurrentYear: time.Now().Year(),
			}, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		p, err := s.pasteRepo.GetPasteByID(ctx, id)
		if err != nil {
			log.Printf("get paste err: %v", err)
			s.renderTemplate(w, "error.html", PageData{
				Message:     "paste not found or expired",
				StatusCode:  http.StatusNotFound,
				CurrentYear: time.Now().Year(),
			}, http.StatusNotFound)
			return
		}

		if p.IsExpired() {
			s.renderTemplate(w, "error.html", PageData{
				Message:     "paste has expired",
				StatusCode:  http.StatusNotFound,
				CurrentYear: time.Now().Year(),
			}, http.StatusNotFound)
			return
		}

		// For file pastes, show the paste.html page with file info and download button
		// Don't automatically redirect to download - let user choose

		if p.IsProtected() {
			if r.Method == http.MethodGet {
				s.renderTemplate(w, "password_prompt.html", PageData{
					PasteID:     p.ID,
					CurrentYear: time.Now().Year(),
					CurrentPort: s.config.Server.Port,
				}, http.StatusOK)
				return
			}

			pass := r.FormValue("password")
			if pass == "" {
				s.renderTemplate(w, "password_prompt.html", PageData{
					PasteID:     p.ID,
					Message:     "password is required",
					CurrentYear: time.Now().Year(),
				}, http.StatusUnauthorized)
				return
			}

			if err := crypt.CompareHashAndPassword(*p.PasswordHash, pass); err != nil {
				s.renderTemplate(w, "password_prompt.html", PageData{
					PasteID:     p.ID,
					Message:     "incorrect password",
					CurrentYear: time.Now().Year(),
				}, http.StatusUnauthorized)
				return
			}

			if len(p.Salt) == 0 {
				s.renderTemplate(w, "error.html", PageData{
					Message:     "salt missing. recreate paste.",
					StatusCode:  500,
					CurrentYear: time.Now().Year(),
				}, 500)
				return
			}

			key := crypt.DeriveKey([]byte(pass), p.Salt)
			content, err := crypt.Decrypt(p.Content, p.EncryptedIV, key)
			if err != nil {
				s.renderTemplate(w, "error.html", PageData{
					Message:     "decryption failed",
					StatusCode:  500,
					CurrentYear: time.Now().Year(),
				}, 500)
				return
			}
			p.Content = string(content)
		}

		s.renderTemplate(w, "paste.html", PageData{
			Paste:       p,
			CurrentYear: time.Now().Year(),
			CurrentPort: s.config.Server.Port,
		}, http.StatusOK)
	}
}

// --- API HANDLERS ---

// this  represents the JSON request body for creating a paste via API.
type CreatePasteRequest struct {
	Content          string `json:"content"`
	Language         string `json:"language"`
	ExpiresInMinutes int    `json:"expires_in_minutes"`
	Password         string `json:"password"`
}

// this represents the JSON response for retrieving a paste via API.
type GetPasteResponse struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// no password_hash, salt, encrypted_iv for API response
}

// handles the API request to create a new paste.
func (s *Server) apiCreatePaste() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			apiErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req CreatePasteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apiErrorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}

		if req.Content == "" {
			apiErrorJSON(w, http.StatusBadRequest, "content is required")
			return
		}
		if len(req.Content) > s.config.Paste.MaxContentSizeBytes {
			apiErrorJSON(w, http.StatusRequestEntityTooLarge, "paste too large")
			return
		}

		language := req.Language
		if language == "" {
			language = "plaintext"
		}

		var (
			hash, encryptedContent string
			salt, iv               []byte
			err                    error
		)

		if req.Password != "" {
			hashPtr, err := crypt.GenerateHash(req.Password, s.config.Security.BcryptCost)
			if err != nil {
				apiErrorJSON(w, 500, "password hash error")
				return
			}
			hash = hashPtr

			salt, err = crypt.GenerateSalt()
			if err != nil {
				apiErrorJSON(w, 500, "salt gen error")
				return
			}

			key := crypt.DeriveKey([]byte(req.Password), salt)
			encryptedContent, iv, err = crypt.Encrypt([]byte(req.Content), key)
			if err != nil {
				apiErrorJSON(w, 500, "encryption failed")
				return
			}
		} else {
			encryptedContent = req.Content
		}

		id, err := idgen.GenerateSecureID(s.config.Paste.IDLength)
		if err != nil {
			apiErrorJSON(w, 500, "id gen error")
			return
		}

		p := &paste.Paste{
			ID:           id,
			Content:      encryptedContent,
			Language:     language,
			CreatedAt:    time.Now(),
			ExpiresAt:    s.config.Paste.GetExpirationTime(req.ExpiresInMinutes),
			PasswordHash: nilIfEmpty(hash),
			Salt:         salt,
			EncryptedIV:  iv,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := s.pasteRepo.CreatePaste(ctx, p); err != nil {
			apiErrorJSON(w, 500, "db error")
			return
		}

		apiRespondJSON(w, http.StatusCreated, map[string]string{
			"id":  p.ID,
			"url": fmt.Sprintf("https://paste.alokranjan.me/%s/", p.ID),
		})
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// handles the API request to retrieve a paste.
func (s *Server) apiGetPaste() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			apiErrorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		id := r.PathValue("id")
		if id == "" {
			apiErrorJSON(w, http.StatusBadRequest, "missing paste id")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		p, err := s.pasteRepo.GetPasteByID(ctx, id)
		if err != nil {
			apiErrorJSON(w, http.StatusNotFound, "paste not found or expired")
			return
		}
		if p.IsExpired() {
			apiErrorJSON(w, http.StatusNotFound, "paste expired")
			return
		}

		if p.IsProtected() {
			pass := r.URL.Query().Get("password")
			if pass == "" {
				apiErrorJSON(w, http.StatusUnauthorized, "password required")
				return
			}
			if err := crypt.CompareHashAndPassword(*p.PasswordHash, pass); err != nil {
				apiErrorJSON(w, http.StatusUnauthorized, "incorrect password")
				return
			}

			if len(p.Salt) == 0 {
				apiErrorJSON(w, 500, "missing salt")
				return
			}
			key := crypt.DeriveKey([]byte(pass), p.Salt)
			decrypted, err := crypt.Decrypt(p.Content, p.EncryptedIV, key)
			if err != nil {
				apiErrorJSON(w, 500, "decrypt failed")
				return
			}
			p.Content = string(decrypted)
		}

		apiRespondJSON(w, 200, GetPasteResponse{
			ID:        p.ID,
			Content:   p.Content,
			Language:  p.Language,
			CreatedAt: p.CreatedAt,
			ExpiresAt: p.ExpiresAt,
		})
	}
}

func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		health := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		}

		// Check storage availability
		if hybrid, ok := s.storage.(*storage.HybridStorage); ok {
			health["storage"] = hybrid.GetStorageInfo()
			health["storage_available"] = s.storage.IsAvailable(ctx)
		} else {
			health["storage_available"] = s.storage.IsAvailable(ctx)
		}

		// Check database connectivity
		// We can add a simple ping to the database here if needed

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}
}

func (s *Server) handleFileDownload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing paste ID", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get paste record from database
		p, err := s.pasteRepo.GetPasteByID(ctx, id)
		if err != nil {
			log.Printf("File download - paste not found: %v", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Check if it's actually a file paste
		if !p.IsFile || p.S3Key == nil {
			http.Error(w, "Not a file paste", http.StatusBadRequest)
			return
		}

		// Check if paste has expired
		if p.IsExpired() {
			http.Error(w, "File has expired", http.StatusGone)
			return
		}

		// Check storage type and serve file securely
		// For local/hybrid storage with local fallback, serve file directly
		filePath := filepath.Join("./data/uploads", *p.S3Key)

		// Security check: ensure file path is within uploads directory
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}

		uploadsDir, _ := filepath.Abs("./data/uploads")
		if !strings.HasPrefix(absPath, uploadsDir) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check if file exists locally (for hybrid storage that might have fallen back to local)
		if _, err := os.Stat(absPath); err == nil {
			// File exists locally, serve it
			w.Header().Set("Content-Type", p.MimeType)
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", p.FileName))
			http.ServeFile(w, r, absPath)
			return
		}

		// If file doesn't exist locally, try S3 presigned URL
		url, err := s.storage.GetDownloadURL(ctx, *p.S3Key)
		if err != nil {
			log.Printf("Failed to get download URL: %v", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}
