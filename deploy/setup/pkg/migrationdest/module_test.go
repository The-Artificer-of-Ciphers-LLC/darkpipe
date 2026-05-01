package migrationdest

import "testing"

func TestValidate(t *testing.T) {
	m := New()
	err := m.Validate(Config{DestIMAP: "localhost:993", DestUser: "u", DestPass: "p", TLSPolicy: RequireTLS})
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}

func TestParseHostPort(t *testing.T) {
	h, p, err := parseHostPort("mail.example.com:1143", 993)
	if err != nil || h != "mail.example.com" || p != 1143 {
		t.Fatalf("parse failed: %v %s %d", err, h, p)
	}
}
