# DarkPipe Context

## Domain terms

- **Profile onboarding**: issuing mailbox onboarding artifacts for a user/device on the home device profile server.
- **Profile onboarding issuance**: the single flow that creates a token-backed setup intent for any supported platform; on token consumption it handles app password provisioning (generate by default, or accept a supplied app password that satisfies the DarkPipe app password format policy), consumes the single-use onboarding token, generates the onboarding artifact, rolls back the app password if artifact generation fails, and never persists plaintext app passwords.
- **Profile onboarding runtime assembly**: the startup-time flow that wires the home device profile server process: environment-derived profile configuration, app password Adapter selection, Profile onboarding handlers, Web UI handlers, monitoring sources, route registration, background delivery parsing, and graceful shutdown.
- **Cloud relay runtime assembly**: the startup-time flow that wires the cloud relay daemon process: environment-derived relay configuration, notification Adapter selection, strict TLS setup, transport forwarder selection, encrypted queue and overflow storage, SMTP server startup, background queue processing, and graceful shutdown.
- **Delivery log ingestion**: reading mail delivery log lines, parsing delivery outcomes, and recording them into monitoring status for the home device profile server.
- **Onboarding artifact**: a platform-specific setup payload, such as a `.mobileconfig` payload, QR PNG, autoconfig/autodiscover payload, or manual setup field set, backed by a single-use onboarding token where possible.
- **Single-use onboarding token**: a short-lived token consumed on first successful retrieval or render of a generated onboarding artifact for any supported platform.
- **Privacy redaction**: shared logging policy that masks user identifiers and secret-bearing query parameters before relay, profile, or diagnostic logs are emitted.
