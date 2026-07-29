package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httptransport "github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/http"
	"github.com/rafixcs/shorter-url-design-system/src/internal/infrastructure/repository"
	"github.com/rafixcs/shorter-url-design-system/src/internal/server"
	"github.com/rafixcs/shorter-url-design-system/src/internal/service"
)

type userResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type shortURLResponse struct {
	UserID   string `json:"user_id"`
	LongURL  string `json:"long_url"`
	ShortURL string `json:"shor_url"`
}

func TestAuthenticationAndShortURLFlow(t *testing.T) {
	t.Parallel()

	repo := repository.NewInMemRepository()
	userService := service.NewUserService(repo)
	shorterService := service.NewService(repo)
	authService := service.NewAuthService(repo, "e2e-test-secret", time.Hour)

	router := server.NewRouter(
		httptransport.NewUserHandler(userService),
		httptransport.NewShorterHandler(shorterService),
		httptransport.NewAuthHandler(authService),
		httptransport.NewAuthMiddleware(authService),
	)

	testServer := httptest.NewServer(router)
	t.Cleanup(testServer.Close)

	client := testServer.Client()
	client.Timeout = 5 * time.Second

	t.Run("status endpoint is public", func(t *testing.T) {
		response := doRequest(t, client, http.MethodGet, testServer.URL+"/status", "", nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)
	})

	var createdUser userResponse
	t.Run("create user", func(t *testing.T) {
		response := doRequest(t, client, http.MethodPost, testServer.URL+"/api/v1/users/", "", map[string]string{
			"name":     "End To End User",
			"email":    "e2e@example.com",
			"password": "strong-test-password",
		})
		defer response.Body.Close()
		assertStatus(t, response, http.StatusCreated)
		decodeJSON(t, response.Body, &createdUser)

		if createdUser.ID == "" {
			t.Fatal("expected the created user to have an ID")
		}
		if createdUser.Password != "" {
			t.Fatal("password must not be exposed in the user response")
		}
	})

	t.Run("protected endpoint rejects a missing token", func(t *testing.T) {
		response := doRequest(t, client, http.MethodGet, testServer.URL+"/api/v1/users/"+createdUser.ID, "", nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusUnauthorized)
	})

	t.Run("login rejects invalid credentials", func(t *testing.T) {
		response := doRequest(t, client, http.MethodPost, testServer.URL+"/api/v1/auth/login", "", map[string]string{
			"email":    createdUser.Email,
			"password": "wrong-password",
		})
		defer response.Body.Close()
		assertStatus(t, response, http.StatusUnauthorized)
	})

	var authentication loginResponse
	t.Run("login with email returns a JWT", func(t *testing.T) {
		response := doRequest(t, client, http.MethodPost, testServer.URL+"/api/v1/auth/login", "", map[string]string{
			"email":    createdUser.Email,
			"password": "strong-test-password",
		})
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)
		decodeJSON(t, response.Body, &authentication)

		if authentication.AccessToken == "" {
			t.Fatal("expected login to return an access token")
		}
		if authentication.TokenType != "Bearer" {
			t.Fatalf("expected Bearer token type, got %q", authentication.TokenType)
		}
		if !authentication.ExpiresAt.After(time.Now()) {
			t.Fatal("expected token expiration to be in the future")
		}
	})

	t.Run("login with user name also returns a JWT", func(t *testing.T) {
		response := doRequest(t, client, http.MethodPost, testServer.URL+"/api/v1/auth/login", "", map[string]string{
			"user":     createdUser.Name,
			"password": "strong-test-password",
		})
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)

		var result loginResponse
		decodeJSON(t, response.Body, &result)
		if result.AccessToken == "" {
			t.Fatal("expected user-name login to return an access token")
		}
	})

	t.Run("get user with JWT", func(t *testing.T) {
		response := doRequest(
			t,
			client,
			http.MethodGet,
			testServer.URL+"/api/v1/users/"+createdUser.ID,
			authentication.AccessToken,
			nil,
		)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)

		var user userResponse
		decodeJSON(t, response.Body, &user)
		if user.ID != createdUser.ID {
			t.Fatalf("expected user ID %q, got %q", createdUser.ID, user.ID)
		}
	})

	const originalURL = "https://example.com/articles/end-to-end-test"
	var shortened shortURLResponse
	t.Run("create short URL with JWT", func(t *testing.T) {
		response := doRequest(
			t,
			client,
			http.MethodPost,
			testServer.URL+"/api/v1/urls/",
			authentication.AccessToken,
			map[string]string{"user_id": createdUser.ID, "long_url": originalURL},
		)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusCreated)
		decodeJSON(t, response.Body, &shortened)

		if shortened.ShortURL == "" {
			t.Fatal("expected a generated short URL")
		}
		if shortened.LongURL != originalURL {
			t.Fatalf("expected long URL %q, got %q", originalURL, shortened.LongURL)
		}
	})

	t.Run("short URL redirects to original URL", func(t *testing.T) {
		redirectClient := *client
		redirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}

		response := doRequest(
			t,
			&redirectClient,
			http.MethodGet,
			testServer.URL+"/"+shortened.ShortURL,
			"",
			nil,
		)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusFound)

		if location := response.Header.Get("Location"); location != originalURL {
			t.Fatalf("expected redirect location %q, got %q", originalURL, location)
		}
	})
}

func doRequest(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	token string,
	body any,
) *http.Response {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		encodedBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request body: %v", err)
		}
		requestBody = bytes.NewReader(encodedBody)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return response
}

func assertStatus(t *testing.T, response *http.Response, expected int) {
	t.Helper()
	if response.StatusCode == expected {
		return
	}

	body, _ := io.ReadAll(response.Body)
	t.Fatalf("expected status %d, got %d: %s", expected, response.StatusCode, body)
}

func decodeJSON(t *testing.T, body io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}
