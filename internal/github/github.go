package github

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/majiayu000/caude-skill-manager/internal/config"
)

// RepoInfo contains parsed GitHub repository information
type RepoInfo struct {
	Owner    string
	Repo     string
	Path     string // subdirectory path (for monorepo skills)
	Branch   string
	FullURL  string
	CloneURL string
}

// ParseGitHubURL parses various GitHub URL formats
// Supports:
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo/tree/branch/path
//   - owner/repo
//   - owner/repo/path
func ParseGitHubURL(input string) (*RepoInfo, error) {
	input = strings.TrimSpace(input)
	input = strings.TrimSuffix(input, "/")
	input = strings.TrimSuffix(input, ".git")

	info := &RepoInfo{}

	// Full GitHub URL
	if strings.HasPrefix(input, "https://github.com/") {
		// Pattern: https://github.com/owner/repo/tree/branch/path
		treePattern := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/tree/([^/]+)(?:/(.+))?`)
		if matches := treePattern.FindStringSubmatch(input); len(matches) >= 4 {
			info.Owner = matches[1]
			info.Repo = matches[2]
			info.Branch = matches[3]
			if len(matches) > 4 {
				info.Path = matches[4]
			}
		} else {
			// Pattern: https://github.com/owner/repo
			simplePattern := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)`)
			if matches := simplePattern.FindStringSubmatch(input); len(matches) >= 3 {
				info.Owner = matches[1]
				info.Repo = matches[2]
				info.Branch = "main" // default
			} else {
				return nil, fmt.Errorf("invalid GitHub URL format: %s", input)
			}
		}
	} else {
		// Short format: owner/repo or owner/repo/path
		parts := strings.Split(input, "/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid format, expected owner/repo: %s", input)
		}
		info.Owner = parts[0]
		info.Repo = parts[1]
		info.Branch = "main"
		if len(parts) > 2 {
			info.Path = strings.Join(parts[2:], "/")
		}
	}

	info.FullURL = fmt.Sprintf("https://github.com/%s/%s", info.Owner, info.Repo)
	info.CloneURL = info.FullURL + ".git"

	return info, nil
}

// DownloadAndExtract downloads a repository and extracts to skills directory
func DownloadAndExtract(info *RepoInfo, targetName string) error {
	// Ensure skills directory exists
	if err := config.EnsureSkillsDir(); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Download as zip
	zipURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.zip",
		info.Owner, info.Repo, info.Branch)

	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try 'master' branch if 'main' fails
		if info.Branch == "main" {
			info.Branch = "master"
			zipURL = fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.zip",
				info.Owner, info.Repo, info.Branch)
			resp, err = http.Get(zipURL)
			if err != nil {
				return fmt.Errorf("failed to download: %w", err)
			}
			defer resp.Body.Close()
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed with status: %s", resp.Status)
		}
	}

	// Create temp file for zip
	tmpFile, err := os.CreateTemp("", "sk-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save zip: %w", err)
	}
	tmpFile.Close()

	// Extract zip
	targetDir := filepath.Join(config.GetSkillsDir(), targetName)

	// Try the specified path first
	err = extractZip(tmpFile.Name(), targetDir, info)
	if err != nil && info.Path != "" {
		// If path doesn't work, try common skill locations
		// e.g., "docx" -> "skills/docx" for anthropics/skills repo
		alternativePaths := []string{
			"skills/" + info.Path,
			"skill/" + info.Path,
		}

		for _, altPath := range alternativePaths {
			infoCopy := *info
			infoCopy.Path = altPath
			os.RemoveAll(targetDir) // Clean up failed attempt
			if err = extractZip(tmpFile.Name(), targetDir, &infoCopy); err == nil {
				return nil
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	return nil
}

// extractZip extracts the zip file to target directory
func extractZip(zipPath, targetDir string, info *RepoInfo) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find the actual root prefix from the zip (it might vary)
	var rootPrefix string
	for _, f := range r.File {
		// First entry should be the root directory
		if strings.Count(f.Name, "/") == 1 && strings.HasSuffix(f.Name, "/") {
			rootPrefix = f.Name
			break
		}
	}

	if rootPrefix == "" {
		// Fallback to expected format
		rootPrefix = fmt.Sprintf("%s-%s/", info.Repo, info.Branch)
	}

	subPath := ""
	if info.Path != "" {
		subPath = info.Path + "/"
	}

	fullPrefix := rootPrefix + subPath

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	extractedFiles := 0
	for _, f := range r.File {
		// Skip files not in the target path
		if !strings.HasPrefix(f.Name, fullPrefix) {
			continue
		}

		// Calculate relative path
		relPath := strings.TrimPrefix(f.Name, fullPrefix)
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(targetDir, relPath)
		if !isWithinDir(targetDir, targetPath) {
			return fmt.Errorf("zip entry escapes target dir: %s", relPath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, f.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Extract file
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
		extractedFiles++
	}

	// Verify SKILL.md exists
	skillMdPath := filepath.Join(targetDir, "SKILL.md")
	if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
		// Clean up
		os.RemoveAll(targetDir)
		if extractedFiles == 0 {
			return fmt.Errorf("no files found at path '%s' - check if the path is correct", info.Path)
		}
		return fmt.Errorf("no SKILL.md found - this doesn't appear to be a valid skill")
	}

	return nil
}

func isWithinDir(root, target string) bool {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// GetSkillName determines the skill name from RepoInfo
func GetSkillName(info *RepoInfo) string {
	if info.Path != "" {
		// Use the last part of the path
		parts := strings.Split(info.Path, "/")
		return parts[len(parts)-1]
	}
	return info.Repo
}
