// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

// mockAppPasswordStore is a simple in-memory store for testing.
type mockAppPasswordStore struct {
	passwords      map[string][]apppassword.AppPassword
	plainPasswords []string
	revoked        []string
}

func newMockAppPasswordStore() *mockAppPasswordStore {
	return &mockAppPasswordStore{
		passwords: make(map[string][]apppassword.AppPassword),
	}
}

func (m *mockAppPasswordStore) Create(email, deviceName, plainPassword string) (*apppassword.AppPassword, error) {
	id := fmt.Sprintf("test-id-%d", len(m.plainPasswords)+1)
	ap := &apppassword.AppPassword{
		ID:         id,
		Email:      email,
		DeviceName: deviceName,
		CreatedAt:  time.Now(),
	}
	m.passwords[email] = append(m.passwords[email], *ap)
	m.plainPasswords = append(m.plainPasswords, plainPassword)
	return ap, nil
}

func (m *mockAppPasswordStore) List(email string) ([]apppassword.AppPassword, error) {
	return m.passwords[email], nil
}

func (m *mockAppPasswordStore) Revoke(id string) error {
	m.revoked = append(m.revoked, id)
	return nil
}

func (m *mockAppPasswordStore) Verify(email, plainPassword string) (bool, error) {
	return true, nil
}

func setupTestHandler() *ProfileHandler {
	return &ProfileHandler{
		ProfileGen:   &mobileconfig.ProfileGenerator{},
		TokenStore:   qrcode.NewMemoryTokenStore(),
		AppPassStore: newMockAppPasswordStore(),
		Config: ServerConfig{
			Domain:      "example.com",
			Hostname:    "mail.example.com",
			CalDAVURL:   "https://mail.example.com/caldav",
			CardDAVURL:  "https://mail.example.com/carddav",
			CalDAVPort:  443,
			CardDAVPort: 443,
			AdminUser:   "admin",
			AdminPass:   "testpass",
		},
	}
}

func TestHandleProfileDownloadWithValidToken(t *testing.T) {
	handler := setupTestHandler()

	// Create a valid token
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)
	token, err := handler.TokenStore.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/x-apple-aspen-config" {
		t.Errorf("Expected Content-Type 'application/x-apple-aspen-config', got '%s'", contentType)
	}

	// Check Content-Disposition header
	contentDisposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "darkpipe-mail.mobileconfig") {
		t.Errorf("Expected Content-Disposition to contain filename, got '%s'", contentDisposition)
	}

	// Check body is not empty
	if w.Body.Len() == 0 {
		t.Error("Expected non-empty response body")
	}
}

func TestHandleProfileDownloadWithInvalidToken(t *testing.T) {
	handler := setupTestHandler()

	// Use invalid token
	req := httptest.NewRequest(http.MethodGet, "/profile/download?token=invalid-token", nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleProfileDownloadWithExpiredToken(t *testing.T) {
	handler := setupTestHandler()

	// Create expired token
	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Minute) // Already expired
	token, err := handler.TokenStore.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleProfileDownloadSingleUseEnforcement(t *testing.T) {
	handler := setupTestHandler()

	// Create valid token
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)
	token, err := handler.TokenStore.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w1 := httptest.NewRecorder()
	handler.HandleProfileDownload(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request expected status 200, got %d", w1.Code)
	}

	// Second request with same token should fail (single-use)
	req2 := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	w2 := httptest.NewRecorder()
	handler.HandleProfileDownload(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Second request expected status 401, got %d", w2.Code)
	}
}

func TestHandleProfileDownloadMissingToken(t *testing.T) {
	handler := setupTestHandler()

	// Request without token parameter
	req := httptest.NewRequest(http.MethodGet, "/profile/download", nil)
	w := httptest.NewRecorder()

	handler.HandleProfileDownload(w, req)

	// Should return 400 Bad Request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAutoconfig(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/mail/config-v1.1.xml?emailaddress=test@example.com", nil)
	w := httptest.NewRecorder()

	handler.HandleAutoconfig(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/xml") {
		t.Errorf("Expected Content-Type to contain 'application/xml', got '%s'", contentType)
	}

	// Verify it's valid XML
	var result interface{}
	if err := xml.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Response is not valid XML: %v", err)
	}

	// Check body contains expected elements
	body := w.Body.String()
	if !strings.Contains(body, "mail.example.com") {
		t.Error("Expected response to contain mail hostname")
	}
}

func TestHandleAutodiscover(t *testing.T) {
	handler := setupTestHandler()

	// Test GET request (simpler case)
	req := httptest.NewRequest(http.MethodGet, "/autodiscover/autodiscover.xml", nil)
	w := httptest.NewRecorder()

	handler.HandleAutodiscover(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/xml") {
		t.Errorf("Expected Content-Type to contain 'application/xml', got '%s'", contentType)
	}

	// Verify it's valid XML
	var result interface{}
	if err := xml.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Response is not valid XML: %v", err)
	}
}

func TestHandleHealth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestHandleQRGenerateWithAuth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/qr/generate?email=test@example.com", nil)
	req.SetBasicAuth("admin", "testpass")
	w := httptest.NewRecorder()

	handler.HandleQRGenerate(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("Expected Content-Type 'image/png', got '%s'", contentType)
	}

	// Check PNG magic number
	body := w.Body.Bytes()
	if len(body) < 8 {
		t.Fatal("Response body too short")
	}

	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i := 0; i < 8; i++ {
		if body[i] != pngMagic[i] {
			t.Errorf("PNG magic number mismatch at byte %d", i)
			break
		}
	}
}

func TestHandleQRGenerateWithoutAuth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/qr/generate?email=test@example.com", nil)
	w := httptest.NewRecorder()

	handler.HandleQRGenerate(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleQRImageWithAuth(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/qr/image?email=test@example.com", nil)
	req.SetBasicAuth("admin", "testpass")
	w := httptest.NewRecorder()

	handler.HandleQRImage(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("Expected Content-Type 'image/png', got '%s'", contentType)
	}

	// Should NOT have Content-Disposition (inline)
	contentDisposition := w.Header().Get("Content-Disposition")
	if contentDisposition != "" {
		t.Errorf("Expected no Content-Disposition for inline image, got '%s'", contentDisposition)
	}
}

func setupTestWebUIHandler() *WebUIHandler {
	// Parse only the templates needed for device management (skip status.html
	// which requires custom template functions registered in main.go).
	tmpl, err := template.ParseFS(embedFS, "templates/add_device.html", "templates/add_device_result.html", "templates/device_list.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse templates: %v", err))
	}
	return &WebUIHandler{
		AppPassStore: newMockAppPasswordStore(),
		TokenStore:   qrcode.NewMemoryTokenStore(),
		Config: ServerConfig{
			Domain:      "example.com",
			Hostname:    "mail.example.com",
			CalDAVURL:   "https://mail.example.com/caldav",
			CardDAVURL:  "https://mail.example.com/carddav",
			CalDAVPort:  443,
			CardDAVPort: 443,
			AdminUser:   "test+<script>alert(1)</script>@example.com",
			AdminPass:   "testpass",
		},
		templates: tmpl,
	}
}

func TestInstructionsHTMLEscaping(t *testing.T) {
	// Test that HTML special characters in email are escaped in all platform outputs.
	// The email contains a <script> tag to simulate an XSS attempt.
	platforms := []string{"android", "thunderbird", "outlook", "other"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			handler := setupTestWebUIHandler()

			form := "device_name=TestDevice&platform=" + platform
			req := httptest.NewRequest(http.MethodPost, "/devices/add", strings.NewReader(form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.SetBasicAuth("test+<script>alert(1)</script>@example.com", "testpass")
			w := httptest.NewRecorder()

			handler.HandleAddDevice(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200, got %d; body: %s", w.Code, w.Body.String())
			}

			body := w.Body.String()

			// The escaped form MUST appear
			if !strings.Contains(body, "&lt;script&gt;") {
				t.Errorf("Expected escaped <script> tag (&lt;script&gt;) in response body")
			}

			// The raw <script> tag must NOT appear in the instructions HTML.
			// We check that the literal <script>alert(1)</script> is not present.
			if strings.Contains(body, "<script>alert(1)</script>") {
				t.Errorf("Found raw <script> tag in response body — HTML escaping is broken")
			}
		})
	}
}

func TestHandleAddDeviceAndroidUsesSuppliedPasswordAndStoresItImmediately(t *testing.T) {
	handler := setupTestWebUIHandler()
	store := handler.AppPassStore.(*mockAppPasswordStore)

	form := "device_name=AndroidPhone&platform=android&app_password=ABCD-EFGH-JKLM-NPQR"
	req := httptest.NewRequest(http.MethodPost, "/devices/add", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("test+<script>alert(1)</script>@example.com", "testpass")
	w := httptest.NewRecorder()

	handler.HandleAddDevice(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if len(store.plainPasswords) != 1 || store.plainPasswords[0] != "ABCD-EFGH-JKLM-NPQR" {
		t.Fatalf("plainPasswords = %v, want supplied password stored once", store.plainPasswords)
	}
	if !strings.Contains(w.Body.String(), "ABCD-EFGH-JKLM-NPQR") {
		t.Fatal("response did not show the supplied app password")
	}
}

func TestHandleAddDeviceIOSDefersPasswordCreationUntilProfileDownload(t *testing.T) {
	webUI := setupTestWebUIHandler()
	store := webUI.AppPassStore.(*mockAppPasswordStore)

	form := "device_name=iPhone&platform=ios"
	req := httptest.NewRequest(http.MethodPost, "/devices/add", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("test+<script>alert(1)</script>@example.com", "testpass")
	w := httptest.NewRecorder()

	webUI.HandleAddDevice(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if len(store.plainPasswords) != 0 {
		t.Fatalf("plainPasswords before token consumption = %v, want none", store.plainPasswords)
	}
	if strings.Contains(w.Body.String(), "App Password:") {
		t.Fatal("deferred iOS setup response exposed an app password before token consumption")
	}

	token := extractProfileDownloadToken(t, w.Body.String())
	profileHandler := &ProfileHandler{
		ProfileGen:   &mobileconfig.ProfileGenerator{},
		TokenStore:   webUI.TokenStore,
		AppPassStore: webUI.AppPassStore,
		Config:       webUI.Config,
	}
	downloadReq := httptest.NewRequest(http.MethodGet, "/profile/download?token="+token, nil)
	downloadW := httptest.NewRecorder()

	profileHandler.HandleProfileDownload(downloadW, downloadReq)

	if downloadW.Code != http.StatusOK {
		t.Fatalf("Expected profile download status 200, got %d; body: %s", downloadW.Code, downloadW.Body.String())
	}
	if len(store.plainPasswords) != 1 {
		t.Fatalf("plainPasswords after token consumption = %v, want one generated password", store.plainPasswords)
	}
	if err := apppassword.ValidateAppPasswordFormat(store.plainPasswords[0]); err != nil {
		t.Fatalf("generated password failed format policy: %v", err)
	}

	secondW := httptest.NewRecorder()
	profileHandler.HandleProfileDownload(secondW, downloadReq)
	if secondW.Code != http.StatusUnauthorized {
		t.Fatalf("second profile download status = %d, want 401", secondW.Code)
	}
}

func TestHandleAddDeviceIOSRejectsSuppliedPasswordWithoutCreatingCredential(t *testing.T) {
	handler := setupTestWebUIHandler()
	store := handler.AppPassStore.(*mockAppPasswordStore)

	form := "device_name=iPhone&platform=ios&app_password=ABCD-EFGH-JKLM-NPQR"
	req := httptest.NewRequest(http.MethodPost, "/devices/add", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("test+<script>alert(1)</script>@example.com", "testpass")
	w := httptest.NewRecorder()

	handler.HandleAddDevice(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Supplied app passwords are not supported for profile downloads") {
		t.Fatal("response did not explain why supplied password was rejected")
	}
	if len(store.plainPasswords) != 0 {
		t.Fatalf("plainPasswords = %v, want none", store.plainPasswords)
	}
}

func extractProfileDownloadToken(t *testing.T, body string) string {
	t.Helper()
	matches := regexp.MustCompile(`/profile/download\?token=([A-Za-z0-9_-]+)`).FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("profile download token not found in body: %s", body)
	}
	return matches[1]
}

func TestExtractEmailFromAutodiscoverXML(t *testing.T) {
	xmlBody := `<?xml version="1.0"?>
	<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
		<Request>
			<EMailAddress>test@example.com</EMailAddress>
			<AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
		</Request>
	</Autodiscover>`

	email := extractEmailFromAutodiscoverXML(xmlBody)
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", email)
	}
}
