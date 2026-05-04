// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package oauthflow

import (
	"fmt"

	"github.com/pterm/pterm"
)

// PTermUI is a rich TTY adapter.
type PTermUI struct{}

func NewPTermUI() *PTermUI { return &PTermUI{} }

func (u *PTermUI) ShowVerification(verificationURL, userCode string) {
	fmt.Println()
	pterm.DefaultBox.
		WithTitle("Authorization Required").
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
		Println(fmt.Sprintf("1. Visit: %s\n\n2. Enter code: %s", verificationURL, userCode))
	fmt.Println()
	pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString(userCode)).Render()
}

func (u *PTermUI) WaitingStart(message string) func() {
	spinner, err := pterm.DefaultSpinner.WithText(message).Start()
	if err != nil {
		return func() {}
	}
	return func() { spinner.Stop() }
}

func (u *PTermUI) ShowSuccess() {
	pterm.Success.Println("Authorization successful!")
	fmt.Println()
}

func (u *PTermUI) ShowError(err error) {
	pterm.Error.Printf("OAuth2 authorization failed: %v\n", err)
}
