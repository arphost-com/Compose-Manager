package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectDocsDiscoversRootAndDocsDirectoryFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "README.md"), "# Project\n")
	writeTestFile(t, filepath.Join(dir, "RUNBOOK"), "operations\n")
	writeTestFile(t, filepath.Join(dir, "notes.md"), "# Internal notes\n")
	writeTestFile(t, filepath.Join(dir, ".env"), "SECRET=value\n")
	writeTestFile(t, filepath.Join(dir, "docs", "runbook.md"), "# Runbook\n")
	writeTestFile(t, filepath.Join(dir, "docs", ".draft", "hidden.md"), "# Hidden\n")

	engine := NewEngine(dir, "")
	project := &Project{Name: "demo", Dir: dir}
	docs := engine.ProjectDocs(project)

	if len(docs) != 4 {
		t.Fatalf("expected 4 docs, got %d: %#v", len(docs), docs)
	}
	if docs[0].Path != generatedProjectDocPath {
		t.Fatalf("expected generated guide first, got %q", docs[0].Path)
	}
	if docs[1].Path != "README.md" {
		t.Fatalf("expected README second, got %q", docs[1].Path)
	}
	if docs[2].Path != "RUNBOOK" {
		t.Fatalf("expected RUNBOOK third, got %q", docs[2].Path)
	}
	if docs[3].Path != "docs/runbook.md" {
		t.Fatalf("expected docs/runbook.md fourth, got %q", docs[3].Path)
	}
}

func TestProjectDocsIncludesGeneratedGuideWhenNoProjectDocsExist(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, "compose.yml")
	writeTestFile(t, composePath, "services:\n  web:\n    image: nginx:stable\n")

	engine := NewEngine(dir, "")
	project := &Project{
		Name:        "demo",
		Dir:         dir,
		ComposeFile: composePath,
		ImageSources: []ImageSource{{
			Service:    "web",
			Image:      "nginx:stable",
			SourceType: "registry",
			Registry:   "docker.io",
		}},
	}

	docs := engine.ProjectDocs(project)
	if len(docs) != 1 {
		t.Fatalf("expected only generated guide, got %d docs: %#v", len(docs), docs)
	}
	if docs[0].Path != generatedProjectDocPath {
		t.Fatalf("expected generated guide path, got %q", docs[0].Path)
	}

	content, err := engine.ReadProjectDoc(project, generatedProjectDocPath)
	if err != nil {
		t.Fatalf("expected generated guide to be readable: %v", err)
	}
	if content.Doc.Path != generatedProjectDocPath {
		t.Fatalf("unexpected generated doc path: %q", content.Doc.Path)
	}
	if !strings.Contains(content.Content, "# demo") {
		t.Fatalf("generated guide missing project heading: %q", content.Content)
	}
	if !strings.Contains(content.Content, "nginx:stable") {
		t.Fatalf("generated guide missing image source: %q", content.Content)
	}
}

func TestReadProjectDocRequiresDiscoveredDoc(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "README.md"), "# Project\n")
	writeTestFile(t, filepath.Join(dir, "notes.md"), "# Internal notes\n")

	engine := NewEngine(dir, "")
	project := &Project{Name: "demo", Dir: dir}

	content, err := engine.ReadProjectDoc(project, "README.md")
	if err != nil {
		t.Fatalf("expected README to be readable: %v", err)
	}
	if content.Content != "# Project\n" {
		t.Fatalf("unexpected content: %q", content.Content)
	}

	if _, err := engine.ReadProjectDoc(project, "notes.md"); err == nil {
		t.Fatal("expected non-discovered root markdown file to be rejected")
	}
	if _, err := engine.ReadProjectDoc(project, "../README.md"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
