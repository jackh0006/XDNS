//go:build !android

package main

import (
	"context"
	"os"

	"github.com/charmbracelet/x/term"
	"xdns-go/internal/client"
	"xdns-go/internal/clientui"
)

func canPromptStartup() bool {
	return term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stderr.Fd())
}

func runClient(ctx context.Context, app *client.Client, intro func()) error {
	if clientui.ShouldUse(app.TerminalUIMode()) {
		return clientui.Run(ctx, app, intro)
	}
	intro()
	return app.Run(ctx)
}
