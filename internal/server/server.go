package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ryu-ryuk/yoru/internal/config"   
	"github.com/ryu-ryuk/yoru/internal/database" 
	"github.com/ryu-ryuk/yoru/internal/paste"    
	"github.com/ryu-ryuk/yoru/pkg/idgen" // to generate secure paste IDs
	"github.com/ryu-ryuk/yoru/pkg/crypt" // for password hashing and encryption
	"golang.org/x/crypto/bcrypt"
)

// template paths
const (
	templatesDir = "./web/templates/"
)

// common struct to pass data to HTML templates.
type PageData struct {
	CurrentYear int
	Message     string
	StatusCode  int
	Paste       *paste.Paste
	PasteID     string 
	CurrentPort int 
}

// holds the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	config     *config.Config
	pasteRepo  paste.Repository
	templates  *template.Template 
}

// creates a new HTTP server instance.
func NewServer(cfg *config.Config, db *database.DB) *Server {
	repo := paste.NewPGRepository(db.Pool)

	// load the templates once at startup
	tmpl, err := template.ParseGlob(templatesDir + "*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	return &Server{
		config:    cfg,
		pasteRepo: repo,
		templates: tmpl, // assign the loaded templates
	}
}

// initializes and starts the HTTP server.
func (s *Server) Start() error {
	router := http.NewServeMux()

	router.HandleFunc("GET /", s.handleIndex())
	router.HandleFunc("POST /", s.handleCreatePaste())
	router.HandleFunc("GET /{id}/", s.handleGetPaste())
	router.HandleFunc("POST /{id}/", s.handleGetPaste()) // POST for password submission

	fs := http.FileServer(http.Dir("./web/static"))
	router.HandleFunc("GET /static/", http.StripPrefix("/static/", fs).ServeHTTP)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("HTTP server listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// --- HELPER FOR RENDERING TEMPLATES ---
func (s *Server) renderTemplate(w http.ResponseWriter, tmpl string, data PageData, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode) // set status code 
	err := s.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		log.Printf("ERROR: Failed to execute template %s: %v", tmpl, err)
		http.Error(w, "Internal server error: Failed to render page", http.StatusInternalServerError)
	}
}

// --- HANDLERS ---

// serves the paste creation form.
func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		content := r.FormValue("content")
		if content == "" {
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Paste content cannot be empty.",
				StatusCode:  http.StatusBadRequest,
				CurrentYear: time.Now().Year(),
			}, http.StatusBadRequest)
			return
		}

		if len(content) > s.config.Paste.MaxContentSizeBytes {
			s.renderTemplate(w, "error.html", PageData{
				Message:     fmt.Sprintf("Paste content exceeds maximum size of %d bytes.", s.config.Paste.MaxContentSizeBytes),
				StatusCode:  http.StatusRequestEntityTooLarge,
				CurrentYear: time.Now().Year(),
			}, http.StatusRequestEntityTooLarge)
			return
		}

		expiresInMinutes := 0
		if val := r.FormValue("expires_in_minutes"); val != "" {
			fmt.Sscanf(val, "%d", &expiresInMinutes)
		}
		expiresAt := s.config.Paste.GetExpirationTime(expiresInMinutes)

		language := r.FormValue("language")
		if language == "" {
			language = "plaintext" // default if not chosen
		}

		password := r.FormValue("password")
		var passwordHash *string
		var pasteSalt []byte
		var encryptedIV []byte
		var pasteContentToStore string // base64-encoded ciphertext

		if password != "" {
			// hash password for comparison later (stored in DB)
			hash, err := crypt.GenerateHash(password, s.config.Security.BcryptCost)
			if err != nil {
				log.Printf("Error hashing password: %v", err)
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Failed to process password.",
					StatusCode:  http.StatusInternalServerError,
					CurrentYear: time.Now().Year(),
				}, http.StatusInternalServerError)
				return
			}
			passwordHash = &hash

			// generate a unique salt for this paste's encryption key derivation
			salt, err := crypt.GenerateSalt()
			if err != nil {
				log.Printf("Error generating encryption salt: %v", err)
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Failed to secure paste.",
					StatusCode:  http.StatusInternalServerError,
					CurrentYear: time.Now().Year(),
				}, http.StatusInternalServerError)
				return
			}
			pasteSalt = salt

			// derive the actual encryption key using PBKDF2
			encryptionKey := crypt.DeriveKey([]byte(password), pasteSalt)

			// encrypt content and Base64 encode it for storage in TEXT column
			base64Ciphertext, iv, err := crypt.Encrypt([]byte(content), encryptionKey)
			if err != nil {
				log.Printf("Error encrypting paste content: %v", err)
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Failed to encrypt content.",
					StatusCode:  http.StatusInternalServerError,
					CurrentYear: time.Now().Year(),
				}, http.StatusInternalServerError)
				return
			}
			pasteContentToStore = base64Ciphertext
			encryptedIV = iv
		} else {
			pasteContentToStore = content // no encryption, store raw content
		}

		id, err := idgen.GenerateSecureID(s.config.Paste.IDLength)
		if err != nil {
			log.Printf("Error generating paste ID: %v", err)
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Failed to generate paste ID.",
				StatusCode:  http.StatusInternalServerError,
				CurrentYear: time.Now().Year(),
			}, http.StatusInternalServerError)
			return
		}

		newPaste := &paste.Paste{
			ID:           id,
			Content:      pasteContentToStore,
			Language:     language,
			CreatedAt:    time.Now(),
			ExpiresAt:    expiresAt,
			PasswordHash: passwordHash,
			Salt:         pasteSalt, // hold the salt 
			EncryptedIV:  encryptedIV,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.pasteRepo.CreatePaste(ctx, newPaste); err != nil {
			log.Printf("Error creating paste in DB: %v", err)
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Failed to save paste. Database error.",
				StatusCode:  http.StatusInternalServerError,
				CurrentYear: time.Now().Year(),
			}, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/%s/", id), http.StatusFound)
	}
}

// retrieves and displays a paste.
func (s *Server) handleGetPaste() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Paste ID missing.",
				StatusCode:  http.StatusBadRequest,
				CurrentYear: time.Now().Year(),
			}, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		p, err := s.pasteRepo.GetPasteByID(ctx, id)
		if err != nil {
			if err.Error() == "paste not found" {
				log.Printf("DEBUG: Paste %s not found in DB or expired.", id) // DEBUG 
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Paste not found or has expired.",
					StatusCode:  http.StatusNotFound,
					CurrentYear: time.Now().Year(),
				}, http.StatusNotFound)
				return
			}
			log.Printf("Error getting paste from DB: %v", err)
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Failed to retrieve paste.",
				StatusCode:  http.StatusInternalServerError,
				CurrentYear: time.Now().Year(),
			}, http.StatusInternalServerError)
			return
		}

		if p.IsExpired() {
			log.Printf("DEBUG: Paste %s found but marked as expired.", p.ID) // DEBUG 
			s.renderTemplate(w, "error.html", PageData{
				Message:     "Paste has expired.",
				StatusCode:  http.StatusNotFound,
				CurrentYear: time.Now().Year(),
			}, http.StatusNotFound)
			return
		}

		if p.IsProtected() {
			if r.Method == http.MethodGet {
				s.renderTemplate(w, "password_prompt.html", PageData{
					PasteID: p.ID,
					CurrentYear: time.Now().Year(),
					CurrentPort: s.config.Server.Port, 
				}, http.StatusOK)
				return
			}

			// a POST request with password submitted
			submittedPassword := r.FormValue("password")
			if submittedPassword == "" {
				log.Printf("DEBUG: Paste %s protected, POST request but no password submitted.", p.ID) // DEBUG 
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Password required to view this paste.",
					StatusCode:  http.StatusUnauthorized,
					CurrentYear: time.Now().Year(),
				}, http.StatusUnauthorized)
				return
			}

			// verify password
			if err := crypt.CompareHashAndPassword(*p.PasswordHash, submittedPassword); err != nil {
				if err == bcrypt.ErrMismatchedHashAndPassword {
					log.Printf("DEBUG: Paste %s protected, incorrect password.", p.ID) // DEBUG
					s.renderTemplate(w, "error.html", PageData{
						Message:     "Incorrect password.",
						StatusCode:  http.StatusUnauthorized,
						CurrentYear: time.Now().Year(),
					}, http.StatusUnauthorized)
					return
				}
				log.Printf("Error comparing password hash: %v", err)
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Failed to verify password.",
					StatusCode:  http.StatusInternalServerError,
					CurrentYear: time.Now().Year(),
				}, http.StatusInternalServerError)
				return
			}

			// get key from submitted password + stored salt
			// the salt must be present if the paste is protected.
			if p.Salt == nil || len(p.Salt) == 0 {
				log.Printf("Security alert: Paste %s is protected but missing salt (DB might be old).", p.ID) // DEBUG
				s.renderTemplate(w, "error.html", PageData{
					Message:    "Security error: Salt missing for protected paste. Please recreate paste.", 
					StatusCode: http.StatusInternalServerError,
					CurrentYear: time.Now().Year(),
				}, http.StatusInternalServerError)
				return
			}
			decryptionKey := crypt.DeriveKey([]byte(submittedPassword), p.Salt)

			decryptedContent, err := crypt.Decrypt(p.Content, p.EncryptedIV, decryptionKey)
			if err != nil {
				log.Printf("Error decrypting paste %s: %v", p.ID, err) // DEBUG
				s.renderTemplate(w, "error.html", PageData{
					Message:     "Failed to decrypt paste content. (Possible corrupted data or wrong key/password)",
					StatusCode:  http.StatusInternalServerError,
					CurrentYear: time.Now().Year(),
				}, http.StatusInternalServerError)
				return
			}
			p.Content = string(decryptedContent) 
			log.Printf("DEBUG: Paste %s decrypted successfully. Content length: %d", p.ID, len(p.Content)) // DEBUG
		} else {
            log.Printf("DEBUG: Paste %s is not protected. Content length: %d", p.ID, len(p.Content)) // DEBUG
        }
		
		// final check before rendering, ensuring content is not empty if it should be present
		if len(p.Content) == 0 {
			log.Printf("DEBUG: Paste %s has empty content after processing. Check data or decryption.", p.ID)
			s.renderTemplate(w, "error.html", PageData{
				Message: "Paste content is empty.",
				StatusCode: http.StatusInternalServerError,
				CurrentYear: time.Now().Year(),
			}, http.StatusInternalServerError)
			return
		}



		s.renderTemplate(w, "paste.html", PageData{
			Paste:       p,
			CurrentYear: time.Now().Year(),
			CurrentPort: s.config.Server.Port, 
		}, http.StatusOK)
	}
}

// helper function for debugging logs
func min(a, b int) int {
    if a < b { return a }
    return b
}
