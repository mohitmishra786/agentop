package cmd

import (
	"github.com/muesli/termenv"
)

func defaultThemeName() string {
	if !termenv.HasDarkBackground() {
		return "light"
	}
	return "dark"
}
