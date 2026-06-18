// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package setupwizard

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/compose"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/config"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/migrate"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/secrets"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/validate"
	"github.com/pterm/pterm"
)

const (
	quickSetupMode    = "Quick (recommended defaults)"
	advancedSetupMode = "Advanced (customize everything)"
)

// Flow orchestrates interactive setup behind one setup wizard Module.
type Flow struct{}

// NewFlow creates a setup wizard flow.
func NewFlow() *Flow { return &Flow{} }

// Run executes the interactive setup wizard.
func (f *Flow) Run() {
	pterm.DefaultHeader.WithFullWidth().Println("DarkPipe Setup Wizard")
	pterm.Info.Println("This wizard will help you configure DarkPipe mail server")
	fmt.Println()

	cfg, ok := prepareConfig()
	if !ok {
		return
	}

	var setupMode string
	modePrompt := &survey.Select{
		Message: "Setup mode:",
		Options: []string{quickSetupMode, advancedSetupMode},
		Default: quickSetupMode,
	}
	if err := survey.AskOne(modePrompt, &setupMode); err != nil {
		pterm.Error.Println("Setup cancelled")
		return
	}

	fmt.Println()
	pterm.DefaultSection.Println("Configuration")

	if err := askRequiredQuestions(cfg); err != nil {
		pterm.Error.Printf("Setup failed: %v\n", err)
		return
	}

	if !isQuickSetupMode(setupMode) {
		if err := askAdvancedQuestions(cfg); err != nil {
			pterm.Error.Printf("Setup failed: %v\n", err)
			return
		}
	}

	if err := migrate.ValidateConfig(cfg); err != nil {
		pterm.Error.Printf("Configuration validation failed: %v\n", err)
		return
	}

	fmt.Println()
	pterm.DefaultSection.Println("Configuration Summary")
	displaySummary(cfg)

	proceed := false
	confirmPrompt := &survey.Confirm{
		Message: "Proceed with this configuration?",
		Default: true,
	}
	if err := survey.AskOne(confirmPrompt, &proceed); err != nil || !proceed {
		pterm.Warning.Println("Setup cancelled")
		return
	}

	generateOutputs(cfg)
}

func existingConfigExists() bool {
	_, err := os.Stat(config.ConfigFile)
	return err == nil
}

func prepareConfig() (*config.Config, bool) {
	cfg := config.DefaultConfig()
	if !existingConfigExists() {
		return cfg, true
	}

	existingCfg, err := config.LoadConfig(config.ConfigFile)
	if err != nil {
		pterm.Warning.Printf("Failed to load existing config: %v\n", err)
		return cfg, true
	}

	pterm.Info.Printf("Found existing configuration (version %s)\n", existingCfg.Version)

	upgrade := false
	prompt := &survey.Confirm{
		Message: "Upgrade existing configuration?",
		Default: true,
	}
	if err := survey.AskOne(prompt, &upgrade); err != nil {
		pterm.Error.Println("Setup cancelled")
		return nil, false
	}

	if !upgrade {
		pterm.Warning.Println("Using existing configuration without changes")
		return nil, false
	}

	if err := migrate.Migrate(existingCfg); err != nil {
		pterm.Error.Printf("Migration failed: %v\n", err)
		return nil, false
	}

	return existingCfg, true
}

func isQuickSetupMode(setupMode string) bool {
	return setupMode == quickSetupMode
}

func askRequiredQuestions(cfg *config.Config) error {
	if err := askMailDomain(cfg); err != nil {
		return err
	}
	if err := askRelayHostname(cfg); err != nil {
		return err
	}
	return askAdminEmail(cfg)
}

func askAdvancedQuestions(cfg *config.Config) error {
	if err := askMailServer(cfg); err != nil {
		return err
	}
	if err := askWebmail(cfg); err != nil {
		return err
	}
	if err := askCalendar(cfg); err != nil {
		return err
	}
	if err := askTransport(cfg); err != nil {
		return err
	}
	if err := askQueueEnabled(cfg); err != nil {
		return err
	}
	return askStrictMode(cfg)
}

func generateOutputs(cfg *config.Config) {
	fmt.Println()
	pterm.DefaultSection.Println("Generating Configuration Files")

	progressbar, _ := pterm.DefaultProgressbar.WithTotal(4).WithTitle("Setup Progress").Start()

	progressbar.UpdateTitle("Generating Docker secrets...")
	if err := secrets.GenerateSecrets(config.SecretsDir); err != nil {
		pterm.Error.Printf("Failed to generate secrets: %v\n", err)
		return
	}
	progressbar.Increment()

	progressbar.UpdateTitle("Creating docker-compose.yml...")
	if err := compose.Generate(cfg, "docker-compose.yml"); err != nil {
		pterm.Error.Printf("Failed to generate compose file: %v\n", err)
		return
	}
	progressbar.Increment()

	progressbar.UpdateTitle("Saving configuration...")
	if err := config.SaveConfig(cfg, config.ConfigFile); err != nil {
		pterm.Error.Printf("Failed to save config: %v\n", err)
		return
	}
	progressbar.Increment()

	progressbar.UpdateTitle("Creating setup marker...")
	if err := os.WriteFile(".darkpipe-configured", []byte("configured"), 0644); err != nil {
		pterm.Error.Printf("Failed to create marker file: %v\n", err)
		return
	}
	progressbar.Increment()

	progressbar.Stop()

	fmt.Println()
	pterm.DefaultBox.WithTitle("Setup Complete!").WithTitleTopCenter().Println(setupCompleteMessage(cfg.MailDomain))
}

func setupCompleteMessage(mailDomain string) string {
	return "DarkPipe has been configured successfully.\n\n" +
		"Next steps:\n" +
		"  1. Review docker-compose.yml\n" +
		"  2. Set up DNS records: darkpipe-dns-setup --domain " + mailDomain + "\n" +
		"  3. Start services: docker compose up -d\n" +
		"  4. Check logs: docker compose logs -f"
}

func askMailDomain(cfg *config.Config) error {
	prompt := &survey.Input{
		Message: "Primary mail domain:",
		Default: cfg.MailDomain,
	}
	validator := survey.WithValidator(func(val interface{}) error {
		domain := val.(string)
		spinner, _ := pterm.DefaultSpinner.Start("Validating DNS for " + domain + "...")

		err := validate.ValidateDomain(domain)
		if err != nil {
			spinner.Fail("DNS validation warning: " + err.Error())
			pterm.Warning.Println("You can continue setup, but make sure to configure DNS before starting services")
			return nil
		}

		spinner.Success("DNS validation passed")
		return nil
	})

	return survey.AskOne(prompt, &cfg.MailDomain, validator)
}

func askRelayHostname(cfg *config.Config) error {
	prompt := &survey.Input{
		Message: "Cloud relay hostname (must have port 25 open):",
		Default: cfg.RelayHostname,
	}
	validator := survey.WithValidator(func(val interface{}) error {
		hostname := val.(string)
		spinner, _ := pterm.DefaultSpinner.Start("Testing SMTP port 25 on " + hostname + "...")

		err := validate.ValidateSMTPPort(hostname, 25, 5*time.Second)
		if err != nil {
			spinner.Fail("Port 25 test warning: " + err.Error())
			pterm.Warning.Println("Many VPS providers block port 25. Ensure your provider allows SMTP.")
			return nil
		}

		spinner.Success("Port 25 is accessible")
		return nil
	})

	return survey.AskOne(prompt, &cfg.RelayHostname, validator)
}

func askAdminEmail(cfg *config.Config) error {
	prompt := &survey.Input{
		Message: "Admin email address:",
		Default: cfg.AdminEmail,
	}
	return survey.AskOne(prompt, &cfg.AdminEmail)
}

func askMailServer(cfg *config.Config) error {
	prompt := &survey.Select{
		Message: "Mail server component:",
		Options: []string{"stalwart", "maddy", "postfix-dovecot"},
		Default: cfg.MailServer,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"stalwart":        "Modern, all-in-one (Rust, IMAP4rev2/JMAP/CalDAV/CardDAV)",
				"maddy":           "Minimal, Go-based single binary",
				"postfix-dovecot": "Traditional, battle-tested MTA+IMAP",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.MailServer)
}

func askWebmail(cfg *config.Config) error {
	prompt := &survey.Select{
		Message: "Webmail component:",
		Options: []string{"none", "roundcube", "snappymail"},
		Default: cfg.Webmail,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"none":       "No webmail (IMAP only)",
				"roundcube":  "Traditional, feature-rich, PHP-based",
				"snappymail": "Modern, fast, lightweight (recommended)",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.Webmail)
}

func askCalendar(cfg *config.Config) error {
	options := calendarOptions(cfg)
	prompt := &survey.Select{
		Message: "Calendar/Contacts component:",
		Options: options,
		Default: cfg.Calendar,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"none":     "No calendar/contacts",
				"radicale": "Standalone CalDAV/CardDAV server",
				"builtin":  "Stalwart built-in CalDAV/CardDAV",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.Calendar)
}

func calendarOptions(cfg *config.Config) []string {
	if cfg.MailServer == "stalwart" {
		return []string{"none", "radicale", "builtin"}
	}

	if cfg.Calendar == "builtin" {
		cfg.Calendar = "radicale"
	}
	return []string{"none", "radicale"}
}

func askTransport(cfg *config.Config) error {
	prompt := &survey.Select{
		Message: "Transport layer:",
		Options: []string{"wireguard", "mtls"},
		Default: cfg.Transport,
		Description: func(value string, index int) string {
			descriptions := map[string]string{
				"wireguard": "WireGuard VPN (recommended, simpler NAT traversal)",
				"mtls":      "Mutual TLS (for restrictive networks)",
			}
			return descriptions[value]
		},
	}
	return survey.AskOne(prompt, &cfg.Transport)
}

func askQueueEnabled(cfg *config.Config) error {
	prompt := &survey.Confirm{
		Message: "Enable message queuing (for offline relay)?",
		Default: cfg.QueueEnabled,
	}
	return survey.AskOne(prompt, &cfg.QueueEnabled)
}

func askStrictMode(cfg *config.Config) error {
	prompt := &survey.Confirm{
		Message: "Enable TLS strict mode?",
		Default: cfg.StrictMode,
	}
	return survey.AskOne(prompt, &cfg.StrictMode)
}

func displaySummary(cfg *config.Config) {
	data := [][]string{
		{"Mail Domain", cfg.MailDomain},
		{"Relay Hostname", cfg.RelayHostname},
		{"Admin Email", cfg.AdminEmail},
		{"Mail Server", cfg.MailServer},
		{"Webmail", cfg.Webmail},
		{"Calendar/Contacts", cfg.Calendar},
		{"Transport", cfg.Transport},
		{"Queue Enabled", fmt.Sprintf("%v", cfg.QueueEnabled)},
		{"TLS Strict Mode", fmt.Sprintf("%v", cfg.StrictMode)},
	}

	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Setting", "Value"},
	}).WithData(data).Render()
}
