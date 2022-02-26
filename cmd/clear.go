package cmd

import (
	"errors"
	"fmt"
	"os"
	"sptlrx/cookie"

	"github.com/muesli/coral"
)

var clearCmd = &coral.Command{
	Use:   "clear",
	Short: "Clear saved cookie",

	RunE: func(cmd *coral.Command, args []string) error {
		err := cookie.Clear()
		if err == nil {
			fmt.Println("Cookie have been cleared.")
		} else if errors.Is(err, os.ErrNotExist) {
			fmt.Println("You haven't saved any cookies 🍪")
			return nil
		}
		return err
	},
}
