package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/browser"
	"github.com/raitonoberu/sptlrx/config"
	"github.com/raitonoberu/sptlrx/services/spotify/auth"
	"github.com/spf13/cobra"
)

var (
	FlagPort int

	FlagClientId     string
	FlagClientSecret string
)

// executeCommand runs a shell command and returns its trimmed output
func executeCommand(cmd string) (string, error) {
	if cmd == "" {
		return "", nil
	}

	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Spotify",

	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.Load()
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		// Priority: flags -> config commands -> config static values -> env vars
		clientId := FlagClientId
		clientSecret := FlagClientSecret

		if clientId == "" && conf != nil && conf.Spotify.ClientIdCmd != "" {
			val, err := executeCommand(conf.Spotify.ClientIdCmd)
			if err != nil {
				return fmt.Errorf("client-id-cmd failed: %w", err)
			}
			clientId = val
		}

		if clientId == "" && conf != nil && conf.Spotify.ClientId != "" {
			clientId = conf.Spotify.ClientId
		}

		if clientId == "" {
			clientId = os.Getenv("SPOTIFY_CLIENT_ID")
		}

		if clientSecret == "" && conf != nil && conf.Spotify.ClientSecretCmd != "" {
			val, err := executeCommand(conf.Spotify.ClientSecretCmd)
			if err != nil {
				return fmt.Errorf("client-secret-cmd failed: %w", err)
			}
			clientSecret = val
		}

		if clientSecret == "" && conf != nil && conf.Spotify.ClientSecret != "" {
			clientSecret = conf.Spotify.ClientSecret
		}

		if clientSecret == "" {
			clientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
		}

		if clientId == "" || clientSecret == "" {
			return errors.New("client_id and client_secret must be provided")
		}

		auth := auth.New(clientId, clientSecret)
		url := auth.GetAuthUrl(FlagPort)

		fmt.Println("Login URL:", url)
		browser.OpenURL(url)

		if err := auth.Login(cmd.Context(), FlagPort); err != nil {
			return err
		}

		if err := auth.Write(); err != nil {
			return err
		}

		fmt.Println("Success! You can use sptlrx now")
		return nil
	},
}

func init() {
	loginCmd.Flags().IntVar(&FlagPort, "port", 8888, "port to use for login callback")
	loginCmd.Flags().StringVar(&FlagClientId, "client-id", "", "spotify client id")
	loginCmd.Flags().StringVar(&FlagClientSecret, "client-secret", "", "spotify client secret")

	rootCmd.AddCommand(pipeCmd)
}
