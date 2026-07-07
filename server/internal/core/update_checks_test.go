package core

import "testing"

func TestDigestFromVerboseManifestObject(t *testing.T) {
	raw := []byte(`{"Descriptor":{"digest":"sha256:abc123"},"SchemaV2Manifest":{}}`)
	if got := digestFromManifest(raw); got != "sha256:abc123" {
		t.Fatalf("digestFromManifest() = %q", got)
	}
}

func TestDigestFromVerboseManifestList(t *testing.T) {
	raw := []byte(`[{"Ref":"example:latest","Descriptor":{"digest":"sha256:def456"}}]`)
	if got := digestFromManifest(raw); got != "sha256:def456" {
		t.Fatalf("digestFromManifest() = %q", got)
	}
}

func TestDigestFromManifestDoesNotGuess(t *testing.T) {
	raw := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:not-the-image-digest"}}`)
	if got := digestFromManifest(raw); got != "" {
		t.Fatalf("digestFromManifest() = %q", got)
	}
}

func TestDigestMatchesRepoDigest(t *testing.T) {
	if !digestMatches("docker.io/library/nginx@sha256:abc123", "sha256:abc123") {
		t.Fatal("expected repo digest to match remote digest")
	}
	if digestMatches("docker.io/library/nginx@sha256:abc123", "sha256:def456") {
		t.Fatal("expected different digests not to match")
	}
}
