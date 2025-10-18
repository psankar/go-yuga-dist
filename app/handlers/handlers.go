package handlers

import (
	"encoding/json"
	"go-yuga-dist/app/model"
	"go-yuga-dist/app/store"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	store store.Store
}

func NewServer(store store.Store) *Server {
	return &Server{store: store}
}

// writeJSON is a helper for writing JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

// writeJSONError is a helper for writing JSON error responses.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) Signup(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if user.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Password cannot be empty")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash password: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to process request")
		return
	}
	user.Password = string(hashedPassword)

	if err := s.store.CreateUser(r.Context(), &user); err != nil {
		log.Printf("ERROR: Failed to create user: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) Signin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		EmailAddress string `json:"email_address"`
		Password     string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), creds.EmailAddress)
	if err != nil {
		log.Printf("ERROR: Failed to get user by email %s: %v", creds.EmailAddress, err)
		writeJSONError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		log.Printf("ERROR: Password mismatch for user %s", creds.EmailAddress)
		writeJSONError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := generateToken()
	if err != nil {
		log.Printf("ERROR: Failed to generate token: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	session := &model.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.store.CreateSession(r.Context(), session); err != nil {
		log.Printf("ERROR: Failed to create session: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	var postReq struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&postReq); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Printf("ERROR: User ID not found in context")
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user, err := s.store.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR: Failed to get user by ID %s: %v", userID, err)
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	post := &model.Post{
		UserID:  userID,
		Content: postReq.Content,
		Region:  user.Region,
	}

	newPostID, err := s.store.CreatePost(r.Context(), post)
	if err != nil {
		log.Printf("ERROR: Failed to create post: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"post_id": newPostID})
}

func (s *Server) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Printf("ERROR: User ID not found in context")
		writeJSONError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	posts, err := s.store.GetPostsByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("ERROR: Failed to get posts for user %s: %v", userID, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func (s *Server) GetUserPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Path[len("/user-post/"):]
	post, err := s.store.GetPostByID(r.Context(), postID)
	if err != nil {
		log.Printf("ERROR: Failed to get post by ID %s: %v", postID, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to get post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}
