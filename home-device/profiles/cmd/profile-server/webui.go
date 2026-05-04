// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/authpolicy"
	"github.com/darkpipe/darkpipe/profiles/pkg/onboarding"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

//go:embed templates/*.html static/*.css
var embedFS embed.FS

// WebUIHandler handles web UI requests for device management
type WebUIHandler struct {
	AppPassStore apppassword.Store
	TokenStore   qrcode.TokenStore
	Config       ServerConfig
	templates    *template.Template
}

// NewWebUIHandler creates a new web UI handler
func NewWebUIHandler(appPassStore apppassword.Store, tokenStore qrcode.TokenStore, config ServerConfig) *WebUIHandler {
	tmpl, err := template.ParseFS(embedFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	return &WebUIHandler{
		AppPassStore: appPassStore,
		TokenStore:   tokenStore,
		Config:       config,
		templates:    tmpl,
	}
}

// Device represents a device (app password) in the UI
type Device struct {
	ID         string
	Email      string
	DeviceName string
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// AddDeviceData holds data for the add device page
type AddDeviceData struct {
	Email   string
	Error   string
	Success bool
}

// AddDeviceResultData holds data for the add device result page
type AddDeviceResultData struct {
	Email         string
	DeviceName    string
	AppPassword   string
	Platform      string
	QRCodeData    string // base64-encoded PNG
	ProfileURL    string
	Title         string
	Summary       string
	Steps         []string
	Fields        []onboarding.InstructionField
	DownloadLabel string
	TokenExpiry   string
	TokenTTL      string
}

// DeviceListData holds data for the device list page
type DeviceListData struct {
	Email   string
	Devices []Device
	Error   string
	Success string
}

// HandleDeviceList displays all devices for the authenticated user
func (h *WebUIHandler) HandleDeviceList(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// List app passwords for this user
	appPasswords, err := h.AppPassStore.List(email)
	if err != nil {
		log.Printf("Failed to list app passwords for %s: %v", logEmail(email), err)
		h.renderDeviceList(w, email, nil, "Failed to load devices", "")
		return
	}

	// Convert to Device structs
	devices := make([]Device, len(appPasswords))
	for i, ap := range appPasswords {
		devices[i] = Device{
			ID:         ap.ID,
			Email:      ap.Email,
			DeviceName: ap.DeviceName,
			CreatedAt:  ap.CreatedAt,
			LastUsedAt: ap.LastUsedAt,
		}
	}

	success := r.URL.Query().Get("success")
	h.renderDeviceList(w, email, devices, "", success)
}

// HandleAddDevice displays the add device form
func (h *WebUIHandler) HandleAddDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if r.Method == http.MethodGet {
		h.renderAddDevice(w, email, "", false)
		return
	}

	if r.Method == http.MethodPost {
		h.processAddDevice(w, r, email)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// processAddDevice handles the form submission
func (h *WebUIHandler) processAddDevice(w http.ResponseWriter, r *http.Request, email string) {
	if err := r.ParseForm(); err != nil {
		h.renderAddDevice(w, email, "Invalid form data", false)
		return
	}

	deviceName := r.FormValue("device_name")
	platform := r.FormValue("platform")

	if deviceName == "" {
		h.renderAddDevice(w, email, "Device name is required", false)
		return
	}

	if platform == "" {
		h.renderAddDevice(w, email, "Platform is required", false)
		return
	}

	// Generate app password
	plainPassword, err := apppassword.GenerateAppPassword()
	if err != nil {
		log.Printf("Failed to generate app password: %v", err)
		h.renderAddDevice(w, email, "Failed to generate password", false)
		return
	}

	// Store app password
	_, err = h.AppPassStore.Create(email, deviceName, plainPassword)
	if err != nil {
		log.Printf("Failed to create app password: %v", err)
		h.renderAddDevice(w, email, "Failed to save password", false)
		return
	}

	setup, err := onboarding.New(nil, h.TokenStore, h.AppPassStore, onboarding.Config{
		Domain:      h.Config.Domain,
		Hostname:    h.Config.Hostname,
		CalDAVURL:   h.Config.CalDAVURL,
		CardDAVURL:  h.Config.CardDAVURL,
		CalDAVPort:  h.Config.CalDAVPort,
		CardDAVPort: h.Config.CardDAVPort,
	}).GeneratePlatformSetup(onboarding.Platform(platform), email, plainPassword)
	if err != nil {
		log.Printf("Failed to generate platform setup: %v", err)
		h.renderAddDevice(w, email, "Failed to generate setup instructions", false)
		return
	}

	// Render result page
	data := AddDeviceResultData{
		Email:         email,
		DeviceName:    deviceName,
		AppPassword:   plainPassword,
		Platform:      string(setup.Platform),
		QRCodeData:    setup.QRCodeData,
		ProfileURL:    setup.ProfileURL,
		Title:         setup.Title,
		Summary:       setup.Summary,
		Steps:         setup.Steps,
		Fields:        setup.Fields,
		DownloadLabel: setup.DownloadLabel,
		TokenExpiry:   setup.TokenExpiry,
		TokenTTL:      setup.TokenTTL,
	}

	if err := h.templates.ExecuteTemplate(w, "add_device_result.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleRevokeDevice revokes an app password
func (h *WebUIHandler) HandleRevokeDevice(w http.ResponseWriter, r *http.Request) {
	_, ok := h.authenticate(w, r)
	if !ok {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/devices?error=Invalid+form+data", http.StatusSeeOther)
		return
	}

	deviceID := r.FormValue("device_id")
	if deviceID == "" {
		http.Redirect(w, r, "/devices?error=Missing+device+ID", http.StatusSeeOther)
		return
	}

	// Revoke the app password
	if err := h.AppPassStore.Revoke(deviceID); err != nil {
		log.Printf("Failed to revoke device %s: %v", deviceID, err)
		http.Redirect(w, r, "/devices?error=Failed+to+revoke+device", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/devices?success=Device+revoked+successfully", http.StatusSeeOther)
}

func (h *WebUIHandler) authPolicyModule() authpolicy.Module {
	return authpolicy.New(authpolicy.Config{AdminUser: h.Config.AdminUser, AdminPass: h.Config.AdminPass})
}

// authenticate checks Basic Auth credentials
func (h *WebUIHandler) authenticate(w http.ResponseWriter, r *http.Request) (string, bool) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Device Management"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", false
	}

	email, err := h.authPolicyModule().Verify(username, password, true)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Basic realm="Device Management"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", false
	}

	return email, true
}

// renderDeviceList renders the device list template
func (h *WebUIHandler) renderDeviceList(w http.ResponseWriter, email string, devices []Device, errorMsg, successMsg string) {
	data := DeviceListData{
		Email:   email,
		Devices: devices,
		Error:   errorMsg,
		Success: successMsg,
	}

	if err := h.templates.ExecuteTemplate(w, "device_list.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// renderAddDevice renders the add device template
func (h *WebUIHandler) renderAddDevice(w http.ResponseWriter, email string, errorMsg string, success bool) {
	data := AddDeviceData{
		Email:   email,
		Error:   errorMsg,
		Success: success,
	}

	if err := h.templates.ExecuteTemplate(w, "add_device.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ServeStatic serves static assets
func (h *WebUIHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	// Remove /static/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" || strings.Contains(path, "..") {
		http.NotFound(w, r)
		return
	}

	content, err := embedFS.ReadFile("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Set content type based on extension
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	w.Write(content)
}
