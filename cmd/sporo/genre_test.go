package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenreVersionFlagPrintsOnlyTheEmbeddedSpecVersion(t *testing.T) {
	cmd := genreCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "2.0.0" {
		t.Fatalf("genre --version: got %q, want 2.0.0", got)
	}
}
