package cmd

import (
	"github.com/muesli/termenv"
)

var env = termenv.EnvColorProfile()

func defaultThemeName() string {
	if !termenv.HasDarkBackground() {
		return "light"
	}
	return "dark"
}
