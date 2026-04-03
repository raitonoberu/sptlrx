package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/browser"
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
		if FlagClientId == "" {
			FlagClientId = os.Getenv("SPOTIFY_CLIENT_ID")
		}
		if FlagClientSecret == "" {
			FlagClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
		}

		if FlagClientId == "" || FlagClientSecret == "" {
			fmt.Println("Credentials not supplied or incomplete.")
			fmt.Println("Use 'sptlrx login -h' for more info.")
			fmt.Println("You may now enter your missing credentials.\n")

			reader := bufio.NewReader(os.Stdin)

			if FlagClientId == "" {
				fmt.Print("Enter spotify client ID: ")
				clientId, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read client id: %w", err)
				}
				FlagClientId = strings.TrimSpace(clientId)
			}

			if FlagClientSecret == "" {
				fmt.Print("Enter spotify client secret: ")
				clientSecret, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read client secret: %w", err)
				}
				FlagClientSecret = strings.TrimSpace(clientSecret)
			}

			fmt.Println()
		}

		if FlagClientId == "" || FlagClientSecret == "" {
			return errors.New("client_id and client_secret are required")
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

	rootCmd.AddCommand(loginCmd)
}
