package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080"
)

// User represents the structure for signup and signin.
type User struct {
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
	Region       string `json:"region"`
}

// Post represents the structure for retrieving posts.
type Post struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	Region     string `json:"region"`
	AuthorName string `json:"author_name"`
}

// TestAPIFlow performs a full integration test of the user and post APIs.
func TestAPIFlow(t *testing.T) {
	t.Log("Starting API integration test flow...")

	// Use a random email for each test run to ensure user uniqueness.
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(100000)
	userEmail := fmt.Sprintf("testuser_%d@example.com", randomInt)
	userRegion := "USA"

	t.Logf("Test configuration: Email=%s, Region=%s", userEmail, userRegion)

	// Create a single client to reuse for the entire test flow.
	client := &http.Client{Timeout: 10 * time.Second}

	// --- Step 1: Sign up a new user ---
	signupUser := User{
		Name:         "Test User",
		EmailAddress: userEmail,
		Password:     "password123",
		Region:       userRegion,
	}
	t.Log("Step 1: Testing user signup...")
	signupBody, _ := json.Marshal(signupUser)
	resp, err := client.Post(baseURL+"/signup", "application/json", bytes.NewBuffer(signupBody))
	require.NoError(t, err, "Signup request should not fail")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Signup should return 201 Created")
	t.Log("✓ User signup successful")

	// --- Step 2: Sign in to get a token ---
	signinCreds := map[string]string{
		"email_address": userEmail,
		"password":      "password123",
	}
	signinBody, _ := json.Marshal(signinCreds)
	resp, err = client.Post(baseURL+"/signin", "application/json", bytes.NewBuffer(signinBody))
	require.NoError(t, err, "Signin request should not fail")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Signin should return 200 OK")

	var signinResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&signinResponse)
	require.NoError(t, err, "Should be able to decode signin response")
	token, ok := signinResponse["token"]
	require.True(t, ok, "Signin response should contain a token")
	require.NotEmpty(t, token, "Token should not be empty")

	// --- Step 3: Create a new post ---
	postContent := "This is a test post for the USA region."
	postPayload := map[string]string{"content": postContent}
	postBody, _ := json.Marshal(postPayload)

	req, _ := http.NewRequest("POST", baseURL+"/user-posts", bytes.NewBuffer(postBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err, "Create post request should not fail")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create post should return 201 Created")
	t.Log("Step 3: Creating a new post...")

	// Decode the created post ID from the response
	var createPostResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&createPostResponse)
	require.NoError(t, err, "Should be able to decode create post response")
	postID, ok := createPostResponse["post_id"]
	require.True(t, ok, "Create post response should contain a post_id")
	require.NotEmpty(t, postID, "Post ID should not be empty")
	t.Logf("✓ Created post with ID: %s", postID)

	t.Log("Step 4: Getting all posts for the user...")
	req, _ = http.NewRequest("GET", baseURL+"/user-posts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err, "Get all posts request should not fail")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get all posts should return 200 OK")

	var userPosts []Post
	err = json.NewDecoder(resp.Body).Decode(&userPosts)
	require.NoError(t, err, "Should be able to decode get all posts response")
	require.Len(t, userPosts, 1, "Should be exactly one post for the user")
	t.Log("✓ Successfully retrieved user posts")

	createdPost := userPosts[0]
	assert.Equal(t, postContent, createdPost.Content, "Post content should match")
	assert.Equal(t, userRegion, createdPost.Region, "Post region should match the user's region")

	t.Log("Step 5: Getting specific post by ID...")
	req, _ = http.NewRequest("GET", baseURL+"/user-post/"+postID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err, "Get single post request should not fail")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get single post should return 200 OK")

	var singlePost Post
	err = json.NewDecoder(resp.Body).Decode(&singlePost)
	require.NoError(t, err, "Should be able to decode get single post response")

	assert.Equal(t, postContent, singlePost.Content, "Single post content should match")
	assert.Equal(t, signupUser.Name, singlePost.AuthorName, "Author name should match")
	assert.Equal(t, userRegion, singlePost.Region, "Single post region should match the user's region")
	t.Log("✓ Successfully retrieved and verified individual post")
	t.Log("✓ All test steps completed successfully")
}
