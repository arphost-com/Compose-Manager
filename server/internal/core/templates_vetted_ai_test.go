package core

import "testing"

func TestVettedAIStackTemplates(t *testing.T) {
	templates := vettedAIStackTemplates()
	want := map[string]string{
		"librechat":       "workflow-rag",
		"onyx":            "workflow-rag",
		"khoj":            "personal-agents",
		"docsgpt":         "workflow-rag",
		"openmemory-mem0": "workflow-rag",
		"langfuse":        "observability",
		"phoenix":         "observability",
		"promptfoo":       "evals",
		"firecrawl":       "search",
		"crawl4ai":        "search",
	}

	if len(templates) != len(want) {
		t.Fatalf("vetted AI template count = %d, want %d", len(templates), len(want))
	}

	seen := make(map[string]StackTemplate, len(templates))
	for _, template := range templates {
		seen[template.ID] = template
		if template.Category != "ai" {
			t.Fatalf("template %s category = %q, want ai", template.ID, template.Category)
		}
		if template.Source == "" {
			t.Fatalf("template %s has empty source", template.ID)
		}
		if template.Image == "" {
			t.Fatalf("template %s has empty image", template.ID)
		}
		if template.ComposeContent == "" {
			t.Fatalf("template %s has empty compose content", template.ID)
		}
		if template.EnvContent == "" {
			t.Fatalf("template %s has empty env content", template.ID)
		}
	}

	for id, subcategory := range want {
		template, ok := seen[id]
		if !ok {
			t.Fatalf("missing vetted AI template %s", id)
		}
		if template.Subcategory != subcategory {
			t.Fatalf("template %s subcategory = %q, want %q", id, template.Subcategory, subcategory)
		}
	}
}
