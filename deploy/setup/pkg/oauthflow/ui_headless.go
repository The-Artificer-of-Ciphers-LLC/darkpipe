// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package oauthflow

// HeadlessUI is a no-op UI adapter suitable for non-interactive callers.
type HeadlessUI struct{}

func NewHeadlessUI() *HeadlessUI { return &HeadlessUI{} }

func (u *HeadlessUI) ShowVerification(verificationURL, userCode string) {}
func (u *HeadlessUI) WaitingStart(message string) func()                { return func() {} }
func (u *HeadlessUI) ShowSuccess()                                      {}
func (u *HeadlessUI) ShowError(err error)                               {}
