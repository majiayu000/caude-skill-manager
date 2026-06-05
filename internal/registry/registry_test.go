package registry

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchSearchIndexFollowsManifestShards(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search-index.json", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"deprecated_full_payload":true,"manifest":"search-index-manifest.json","v":"2026-05-24","t":2}`)
	})
	mux.HandleFunc("/search-index-manifest.json", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"v":"2026-05-24","total_count":2,"shards":[{"gzip_path":"search-shards/part-000.json.gz","count":2}]}`)
	})
	mux.HandleFunc("/search-shards/part-000.json.gz", func(w http.ResponseWriter, r *http.Request) {
		writeGzip(t, w, `{"v":"2026-05-24","count":2,"s":[{"n":"frontend-testing","d":"testing skill","c":"tst","g":["test"],"r":10,"i":"owner/repo/.agents/skills/frontend-testing/SKILL.md","b":"main"},{"n":"docx","d":"docs","c":"doc","g":[],"r":5,"i":"anthropics/skills/skills/docx","b":"main"}]}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	idx, err := fetchSearchIndexFromBaseURL(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if idx.TotalCount != 2 || len(idx.Skills) != 2 {
		t.Fatalf("expected two skills, got total=%d len=%d", idx.TotalCount, len(idx.Skills))
	}
	if idx.Skills[0].Install != "owner/repo/.agents/skills/frontend-testing/SKILL.md" {
		t.Fatalf("unexpected install: %s", idx.Skills[0].Install)
	}
}

func TestFetchCategoryFollowsManifestParts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/categories/testing.json", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"category":"testing","code":"testing","count":1,"deprecated_full_payload":true,"manifest":"categories/testing/manifest.json"}`)
	})
	mux.HandleFunc("/categories/testing/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"category":"testing","code":"testing","count":1,"parts":[{"gzip_path":"categories/testing/part-000.json.gz","count":1}]}`)
	})
	mux.HandleFunc("/categories/testing/part-000.json.gz", func(w http.ResponseWriter, r *http.Request) {
		writeGzip(t, w, `{"category":"testing","count":1,"skills":[{"name":"frontend-testing","description":"testing skill","install":"owner/repo/.agents/skills/frontend-testing/SKILL.md","category":"testing"}]}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	category, err := fetchCategoryFromBaseURL(server.URL, "testing")
	if err != nil {
		t.Fatal(err)
	}
	if category.Count != 1 || len(category.Skills) != 1 {
		t.Fatalf("expected one skill, got total=%d len=%d", category.Count, len(category.Skills))
	}
	if category.Skills[0].Name != "frontend-testing" {
		t.Fatalf("unexpected skill: %s", category.Skills[0].Name)
	}
}

func TestDedupeSkillsUsesInstallCommand(t *testing.T) {
	skills := []Skill{
		{Name: "frontend-testing-owner-repo", Install: "owner/repo/.agents/skills/frontend-testing/SKILL.md"},
		{Name: "frontend-testing", Install: "owner/repo/.agents/skills/frontend-testing/SKILL.md"},
		{Name: "docx", Install: "anthropics/skills/skills/docx"},
	}

	got := dedupeSkills(skills)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique skills, got %d", len(got))
	}
	if got[0].Name != "frontend-testing" {
		t.Fatalf("expected first skill to win, got %s", got[0].Name)
	}
}

func TestInstallableSkillRefRejectsCommandMarkdown(t *testing.T) {
	valid := []string{
		"anthropics/skills/skills/docx",
		"langgenius/dify/.agents/skills/frontend-testing/SKILL.md",
		"redmage123/salesforce/.agents/project_analysis_agent_SKILL.md",
	}
	for _, install := range valid {
		if !isInstallableSkillRef(install) {
			t.Fatalf("expected installable ref: %s", install)
		}
	}

	invalid := "udecode/plate/.claude/commands/sync-testing-skill.md"
	if isInstallableSkillRef(invalid) {
		t.Fatalf("expected command markdown to be rejected: %s", invalid)
	}
}

func TestResolveInstallFromIndexRejectsCommandMarkdown(t *testing.T) {
	idx := &SearchIndex{
		Skills: []SearchIndexEntry{
			{Name: "sync-testing-skill", Install: "udecode/plate/.claude/commands/sync-testing-skill.md"},
			{Name: "frontend-testing", Install: "langgenius/dify/.agents/skills/frontend-testing/SKILL.md"},
		},
	}

	if _, err := resolveInstallFromIndex(idx, "sync-testing-skill"); err == nil {
		t.Fatal("expected command markdown ref to be rejected")
	}

	install, err := resolveInstallFromIndex(idx, "frontend-testing")
	if err != nil {
		t.Fatal(err)
	}
	if install != "langgenius/dify/.agents/skills/frontend-testing/SKILL.md" {
		t.Fatalf("unexpected install: %s", install)
	}
}

func writeGzip(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(b.Bytes()); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
}
