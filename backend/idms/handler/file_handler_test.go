package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestOwnsKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("username", "alice")

	// Same user's file
	if !ownsKey(c, "alice/1234_report.pdf") {
		t.Fatalf("expected ownsKey to be true for alice's file")
	}

	// Another user's file
	if ownsKey(c, "bob/1234_report.pdf") {
		t.Fatalf("expected ownsKey to be false for bob's file")
	}

	// Subfolder or prefix bypass attempt like "alice_evil/file"
	if ownsKey(c, "alice_evil/1234_file.pdf") {
		t.Fatalf("expected ownsKey to be false for prefix collision")
	}
}

