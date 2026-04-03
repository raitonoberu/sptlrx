package cmd

import (
	"errors"
	"fmt"
	"os"

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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Spotify",

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config to check for credentials
		conf, err := config.Load()
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		// Priority: flags -> config -> env vars
		if FlagClientId == "" {
			if conf != nil && conf.Spotify.ClientId != "" {
				FlagClientId = conf.Spotify.ClientId
			} else {
				FlagClientId = os.Getenv("SPOTIFY_CLIENT_ID")
			}
		}
		if FlagClientSecret == "" {
			if conf != nil && conf.Spotify.ClientSecret != "" {
				FlagClientSecret = conf.Spotify.ClientSecret
			} else {
				FlagClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
			}
		}

		if FlagClientId == "" || FlagClientSecret == "" {
			return errors.New("client_id and client_secret must be provided")
		}

		auth := auth.New(FlagClientId, FlagClientSecret)
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
