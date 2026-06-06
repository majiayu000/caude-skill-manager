package github

import "testing"

func TestParseGitHubURLTrimsDirectorySkillFile(t *testing.T) {
	info, err := ParseGitHubURL("langgenius/dify/.agents/skills/frontend-testing/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if info.Path != ".agents/skills/frontend-testing" {
		t.Fatalf("unexpected path: %s", info.Path)
	}
	if info.FilePath != "" {
		t.Fatalf("expected directory install, got file path: %s", info.FilePath)
	}
	if name := GetSkillName(info); name != "frontend-testing" {
		t.Fatalf("unexpected skill name: %s", name)
	}
}

func TestParseGitHubURLKeepsSingleSkillFile(t *testing.T) {
	info, err := ParseGitHubURL("redmage123/salesforce/.agents/project_analysis_agent_SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if info.FilePath != ".agents/project_analysis_agent_SKILL.md" {
		t.Fatalf("unexpected file path: %s", info.FilePath)
	}
	if info.Path != "" {
		t.Fatalf("expected file install, got directory path: %s", info.Path)
	}
	if name := GetSkillName(info); name != "project_analysis_agent" {
		t.Fatalf("unexpected skill name: %s", name)
	}
}

func TestParseGitHubURLDoesNotTreatCommandMarkdownAsSkillFile(t *testing.T) {
	info, err := ParseGitHubURL("udecode/plate/.claude/commands/sync-testing-skill.md")
	if err != nil {
		t.Fatal(err)
	}
	if info.FilePath != "" {
		t.Fatalf("command markdown should not be treated as a skill file: %s", info.FilePath)
	}
	if info.Path != ".claude/commands/sync-testing-skill.md" {
		t.Fatalf("unexpected path: %s", info.Path)
	}
}
