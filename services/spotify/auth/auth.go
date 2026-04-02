package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

const authPath = "sptlrx/spotify-auth.json"

type Auth struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func New(clientId, clientSecret string) *Auth {
	return &Auth{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
}

func Load() (*Auth, error) {
	path, err := xdg.StateFile(authPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	auth := Auth{}
	if err := json.NewDecoder(f).Decode(&auth); err != nil {
		return nil, err
	}

	return &auth, nil
}

func (a *Auth) GetAuthUrl(port int) string {
	return "https://accounts.spotify.com/authorize?" + url.Values{
		"client_id":     {a.ClientId},
		"response_type": {"code"},
		"redirect_uri":  {getRedirectUri(port)},
		"scope":         {"user-read-currently-playing user-read-playback-state"},
	}.Encode()
}

func (a *Auth) Login(ctx context.Context, port int) error {
	code, err := getAuthCode(ctx, port)
	if err != nil {
		return err
	}

	return a.refresh(ctx, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {getRedirectUri(port)},
	})
}

func (a *Auth) GetToken(ctx context.Context) (string, error) {
	if time.Until(a.ExpiresAt) > 5*time.Second {
		return a.AccessToken, nil
	}

	if a.RefreshToken == "" {
		return "", errors.New("refresh_token can't be empty")
	}

	if err := a.refresh(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {a.RefreshToken},
	}); err != nil {
		return "", err
	}

	return a.AccessToken, a.Write()
}

func (a *Auth) Write() error {
	path, err := xdg.StateFile(authPath)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(a)
}

func (a *Auth) refresh(ctx context.Context, form url.Values) error {
	if a.ClientId == "" || a.ClientSecret == "" {
		return errors.New("client_id & client_secret can't be empty")
	}

	reqBody := strings.NewReader(form.Encode())
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://accounts.spotify.com/api/token", reqBody)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(a.ClientId, a.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}

	var response refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	a.AccessToken = response.AccessToken
	a.ExpiresAt = time.Now().Add(time.Second * time.Duration(response.ExpiresIn))
	if response.RefreshToken != "" {
		a.RefreshToken = response.RefreshToken
	}

	return nil
}

func getRedirectUri(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", port)
}

func getAuthCode(ctx context.Context, port int) (string, error) {
	mux := http.NewServeMux()
	ctx, cancel := context.WithCancel(ctx)

	var code string
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code = r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(400)
			w.Write([]byte("error! no code in query"))
			return
		}

		w.Write([]byte("success! you can close this tab"))
		cancel()
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return "", err
	}
	return code, nil
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
