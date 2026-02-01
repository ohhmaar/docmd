package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func GetOAuthConfig(creds *GoogleCredentials, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     creds.Installed.ClientID,
		ClientSecret: creds.Installed.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes: []string{
			drive.DriveFileScope,
		},
	}
}

func Authenticate(creds *GoogleCredentials) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)

	config := GetOAuthConfig(creds, redirectURL)

	state, err := generateState()
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			if r.URL.Query().Get("state") != state {
				errChan <- fmt.Errorf("state mismatch - possible CSRF attack")
				http.Error(w, "State mismatch", http.StatusBadRequest)
				return
			}

			if errMsg := r.URL.Query().Get("error"); errMsg != "" {
				errChan <- fmt.Errorf("authorization error: %s", errMsg)
				http.Error(w, "Authorization failed", http.StatusBadRequest)
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				errChan <- fmt.Errorf("no authorization code received")
				http.Error(w, "No code received", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>docmd - Authorization Successful</title></head>
<body style="font-family: sans-serif; text-align: center; padding: 50px;">
<h1>Authorization Successful!</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>`)

			codeChan <- code
		}),
	}

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Println("Opening browser for Google authorization...")
	fmt.Printf("If the browser doesn't open, visit this URL:\n%s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Println("Could not open browser automatically.")
	}

	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authorization timeout - no response within 5 minutes")
	}

	server.Shutdown(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func IsConfigured() bool {
	return false
}
