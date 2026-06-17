// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"embed"
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/authpolicy"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
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
	suppliedPassword := r.FormValue("app_password")

	if deviceName == "" {
		h.renderAddDevice(w, email, "Device name is required", false)
		return
	}

	if platform == "" {
		h.renderAddDevice(w, email, "Platform is required", false)
		return
	}

	if suppliedPassword != "" && (platform == string(onboarding.PlatformIOS) || platform == string(onboarding.PlatformMacOS)) {
		h.renderAddDevice(w, email, "Supplied app passwords are not supported for profile downloads", false)
		return
	}

	module := onboarding.New(&mobileconfig.ProfileGenerator{}, h.TokenStore, h.AppPassStore, onboarding.Config{
		Domain:      h.Config.Domain,
		Hostname:    h.Config.Hostname,
		CalDAVURL:   h.Config.CalDAVURL,
		CardDAVURL:  h.Config.CardDAVURL,
		CalDAVPort:  h.Config.CalDAVPort,
		CardDAVPort: h.Config.CardDAVPort,
	})
	intent, err := module.IssueSetupIntent(onboarding.SetupIntentInput{
		Email:      email,
		DeviceName: deviceName,
		Platform:   onboarding.Platform(platform),
	})
	if err != nil {
		log.Printf("Failed to issue setup intent: %v", err)
		h.renderAddDevice(w, email, "Failed to create setup token", false)
		return
	}
	if platform == string(onboarding.PlatformIOS) || platform == string(onboarding.PlatformMacOS) {
		downloadURL := "https://" + h.Config.Hostname + "/profile/download?token=" + intent.Token
		png, err := qrcode.GenerateQRCodePNG(downloadURL, 256)
		if err != nil {
			log.Printf("Failed to generate setup QR: %v", err)
			h.renderAddDevice(w, email, "Failed to generate setup QR", false)
			return
		}
		setup := &onboarding.PlatformSetup{
			Platform:      onboarding.Platform(platform),
			Title:         "iOS/macOS Setup",
			Steps:         []string{"Scan the QR code below with your device camera, OR", "Click the \"Download Profile\" button below", "Follow the prompts to install the configuration profile", "Your email, calendar, and contacts will sync automatically"},
			DownloadLabel: "Download Profile (.mobileconfig)",
			QRCodeData:    base64.StdEncoding.EncodeToString(png),
			ProfileURL:    downloadURL,
			TokenExpiry:   intent.ExpiresAt.Format(time.RFC3339),
			TokenTTL:      time.Until(intent.ExpiresAt).Round(time.Second).String(),
		}
		h.renderAddDeviceResult(w, AddDeviceResultData{
			Email:         email,
			DeviceName:    deviceName,
			Platform:      string(setup.Platform),
			QRCodeData:    setup.QRCodeData,
			ProfileURL:    setup.ProfileURL,
			Title:         setup.Title,
			Steps:         setup.Steps,
			DownloadLabel: setup.DownloadLabel,
			TokenExpiry:   setup.TokenExpiry,
			TokenTTL:      setup.TokenTTL,
		})
		return
	}
	consumed, err := module.ConsumeSetupIntent(onboarding.ConsumeSetupInput{
		Token:               intent.Token,
		SuppliedAppPassword: suppliedPassword,
	})
	if err != nil {
		log.Printf("Failed to consume setup intent: %v", err)
		h.renderAddDevice(w, email, "Failed to generate setup", false)
		return
	}
	setup := consumed.Setup

	// Render result page
	data := AddDeviceResultData{
		Email:         email,
		DeviceName:    deviceName,
		AppPassword:   consumed.AppPassword,
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

	h.renderAddDeviceResult(w, data)
}

func (h *WebUIHandler) renderAddDeviceResult(w http.ResponseWriter, data AddDeviceResultData) {
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
