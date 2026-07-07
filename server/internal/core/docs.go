package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxProjectDocBytes = 2 * 1024 * 1024

var projectDocExtensions = map[string]bool{
	".adoc":     true,
	".md":       true,
	".mdown":    true,
	".markdown": true,
	".rst":      true,
	".txt":      true,
}

var rootProjectDocPrefixes = []string{
	"readme",
	"changelog",
	"install",
	"upgrade",
	"deploy",
	"deployment",
	"operations",
	"runbook",
}

var projectDocDirs = []string{"docs", "doc", "documentation"}

// ProjectDocs returns operator documentation files that live inside a project.
func (e *Engine) ProjectDocs(project *Project) []ProjectDoc {
	if project == nil {
		return nil
	}
	return discoverProjectDocs(project.Dir)
}

// ReadProjectDoc returns the content of one discovered project documentation file.
func (e *Engine) ReadProjectDoc(project *Project, docPath string) (*ProjectDocContent, error) {
	if project == nil {
		return nil, &ErrNotFound{Msg: "project not found"}
	}
	cleaned, err := cleanProjectDocPath(docPath)
	if err != nil {
		return nil, err
	}
	var selected *ProjectDoc
	for _, doc := range discoverProjectDocs(project.Dir) {
		if doc.Path == cleaned {
			docCopy := doc
			selected = &docCopy
			break
		}
	}
	if selected == nil {
		return nil, &ErrNotFound{Msg: "project documentation not found: " + cleaned}
	}
	if selected.SizeBytes > maxProjectDocBytes {
		return nil, fmt.Errorf("project documentation file is too large to display")
	}
	fullPath, err := projectRelativePath(project.Dir, selected.Path)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return &ProjectDocContent{Doc: *selected, Content: string(content)}, nil
}

func discoverProjectDocs(projectDir string) []ProjectDoc {
	seen := map[string]bool{}
	var docs []ProjectDoc
	add := func(fullPath string) {
		doc, ok := projectDocFromPath(projectDir, fullPath)
		if !ok || seen[doc.Path] {
			return
		}
		seen[doc.Path] = true
		docs = append(docs, doc)
	}

	if entries, err := os.ReadDir(projectDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			if isRootProjectDoc(entry.Name()) {
				add(filepath.Join(projectDir, entry.Name()))
			}
		}
	}

	for _, dirName := range projectDocDirs {
		docDir := filepath.Join(projectDir, dirName)
		info, err := os.Stat(docDir)
		if err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(docDir, func(fullPath string, entry os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if entry.IsDir() {
				if fullPath != docDir && strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if isProjectDocFile(entry.Name()) {
				add(fullPath)
			}
			return nil
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Path < docs[j].Path
	})
	return docs
}

func projectDocFromPath(projectDir, fullPath string) (ProjectDoc, bool) {
	info, err := os.Lstat(fullPath)
	if err != nil || info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ProjectDoc{}, false
	}
	rel, err := filepath.Rel(projectDir, fullPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." || filepath.IsAbs(rel) {
		return ProjectDoc{}, false
	}
	rel = filepath.ToSlash(rel)
	return ProjectDoc{
		Title:     projectDocTitle(rel),
		Path:      rel,
		FileName:  filepath.Base(fullPath),
		SizeBytes: info.Size(),
		UpdatedAt: info.ModTime().UTC().Truncate(time.Second),
	}, true
}

func isRootProjectDoc(name string) bool {
	ext := filepath.Ext(name)
	base := strings.ToLower(name)
	if ext != "" {
		if !isProjectDocFile(name) {
			return false
		}
		base = strings.TrimSuffix(base, strings.ToLower(ext))
	} else if strings.Contains(base, ".") {
		return false
	}
	for _, prefix := range rootProjectDocPrefixes {
		if base == prefix || strings.HasPrefix(base, prefix+"-") || strings.HasPrefix(base, prefix+"_") {
			return true
		}
	}
	return false
}

func isProjectDocFile(name string) bool {
	return projectDocExtensions[strings.ToLower(filepath.Ext(name))]
}

func projectDocTitle(rel string) string {
	base := strings.TrimSuffix(path.Base(rel), path.Ext(rel))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.TrimSpace(base)
	if base == "" {
		return rel
	}
	words := strings.Fields(base)
	for i, word := range words {
		if strings.EqualFold(word, "readme") {
			words[i] = "README"
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func cleanProjectDocPath(docPath string) (string, error) {
	docPath = strings.TrimSpace(strings.ReplaceAll(docPath, "\\", "/"))
	if docPath == "" {
		return "", fmt.Errorf("documentation path is required")
	}
	if strings.HasPrefix(docPath, "/") {
		return "", fmt.Errorf("documentation path must be relative")
	}
	cleaned := path.Clean(docPath)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("documentation path is outside the project")
	}
	return cleaned, nil
}

func projectRelativePath(projectDir, rel string) (string, error) {
	fullPath := filepath.Join(projectDir, filepath.FromSlash(rel))
	projectAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", err
	}
	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(projectAbs, fullAbs)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || relative == ".." || filepath.IsAbs(relative) {
		return "", fmt.Errorf("documentation path is outside the project")
	}
	return fullAbs, nil
}
