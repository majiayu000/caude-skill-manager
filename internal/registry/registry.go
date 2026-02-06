package registry

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/majiayu000/caude-skill-manager/internal/config"
)

const (
	// DefaultRegistryURL is the base URL for the skill registry
	DefaultRegistryURL = "https://raw.githubusercontent.com/majiayu000/claude-skill-registry/main"
)

// Registry represents the skill registry
type Registry struct {
	Version    string  `json:"version"`
	UpdatedAt  string  `json:"updated_at"`
	TotalCount int     `json:"total_count"`
	Skills     []Skill `json:"skills"`
}

// RegistrySource indicates where registry data came from.
type RegistrySource string

const (
	RegistrySourceRemote RegistrySource = "remote"
	RegistrySourceCache  RegistrySource = "cache"
)

// Skill represents a skill in the registry
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Install     string   `json:"install"`
	Repo        string   `json:"repo"`
	Path        string   `json:"path"`
	Branch      string   `json:"branch"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Source      string   `json:"source"`
	Stars       int      `json:"stars"`
	Featured    bool     `json:"featured"`
}

// GitHubURL returns the GitHub URL for viewing this skill's SKILL.md
func (s *Skill) GitHubURL() string {
	if s.Repo == "" {
		return ""
	}
	branch := s.Branch
	if branch == "" {
		branch = "main"
	}
	if s.Path != "" {
		return fmt.Sprintf("https://github.com/%s/blob/%s/%s/SKILL.md", s.Repo, branch, s.Path)
	}
	return fmt.Sprintf("https://github.com/%s/blob/%s/SKILL.md", s.Repo, branch)
}

// Category represents a category index
type Category struct {
	Category  string  `json:"category"`
	UpdatedAt string  `json:"updated_at"`
	Count     int     `json:"count"`
	Skills    []Skill `json:"skills"`
}

// Featured represents the featured skills response
type Featured struct {
	UpdatedAt string  `json:"updated_at"`
	Count     int     `json:"count"`
	Skills    []Skill `json:"skills"`
}

// CategoryIndexEntry represents one entry in the category index
type CategoryIndexEntry struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Count int    `json:"count"`
}

// CategoryIndex represents the category index
type CategoryIndex struct {
	UpdatedAt  string               `json:"updated_at"`
	Categories []CategoryIndexEntry `json:"categories"`
}

// SearchIndexEntry represents one skill in the compact search index
type SearchIndexEntry struct {
	Name        string   `json:"n"`
	Description string   `json:"d"`
	Category    string   `json:"c"`
	Tags        []string `json:"g"`
	Stars       int      `json:"r"`
	Install     string   `json:"i"`
	Branch      string   `json:"b"`
}

// SearchIndex represents the compact search index
type SearchIndex struct {
	Version    string             `json:"v"`
	TotalCount int                `json:"t"`
	Skills     []SearchIndexEntry `json:"s"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func registryBaseURL() string {
	base := config.GetRegistryBaseURL()
	if base == "" || base == "github" {
		return DefaultRegistryURL
	}
	return strings.TrimRight(base, "/")
}

func docsBaseURL() string {
	return registryBaseURL() + "/docs"
}

// FetchRegistry fetches the full registry
func FetchRegistry() (*Registry, error) {
	registry, _, err := FetchRegistryWithSource()
	return registry, err
}

// FetchRegistryWithSource fetches the full registry and indicates data source.
func FetchRegistryWithSource() (*Registry, RegistrySource, error) {
	url := registryBaseURL() + "/registry.json"

	resp, err := httpClient.Get(url)
	if err != nil {
		if cached, cacheErr := loadRegistryCache(); cacheErr == nil {
			return cached, RegistrySourceCache, nil
		}
		return nil, "", fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if cached, cacheErr := loadRegistryCache(); cacheErr == nil {
			return cached, RegistrySourceCache, nil
		}
		return nil, "", fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var registry Registry
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		if cached, cacheErr := loadRegistryCache(); cacheErr == nil {
			return cached, RegistrySourceCache, nil
		}
		return nil, "", fmt.Errorf("failed to parse registry: %w", err)
	}

	_ = saveRegistryCache(&registry)
	return &registry, RegistrySourceRemote, nil
}

// FetchFeatured fetches the top featured skills (52KB vs 44MB full registry)
func FetchFeatured() (*Featured, error) {
	url := docsBaseURL() + "/featured.json"

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch featured: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("featured returned status %d", resp.StatusCode)
	}

	var featured Featured
	if err := json.NewDecoder(resp.Body).Decode(&featured); err != nil {
		return nil, fmt.Errorf("failed to parse featured: %w", err)
	}

	return &featured, nil
}

// FetchCategory fetches skills for a specific category from docs
func FetchCategory(category string) (*Category, error) {
	url := fmt.Sprintf("%s/categories/%s.json", docsBaseURL(), category)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("category returned status %d", resp.StatusCode)
	}

	var cat Category
	if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
		return nil, fmt.Errorf("failed to parse category: %w", err)
	}

	return &cat, nil
}

// FetchCategoryIndex fetches the category index listing all available categories
func FetchCategoryIndex() (*CategoryIndex, error) {
	url := docsBaseURL() + "/categories/index.json"

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("category index returned status %d", resp.StatusCode)
	}

	var idx CategoryIndex
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, fmt.Errorf("failed to parse category index: %w", err)
	}

	return &idx, nil
}

// FetchSearchIndex fetches the gzip-compressed search index (~9MB vs 44MB full registry)
func FetchSearchIndex() (*SearchIndex, RegistrySource, error) {
	if cached, err := loadSearchIndexCache(); err == nil {
		return cached, RegistrySourceCache, nil
	}

	url := docsBaseURL() + "/search-index.json.gz"

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch search index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("search index returned status %d", resp.StatusCode)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decompress search index: %w", err)
	}
	defer gzReader.Close()

	var idx SearchIndex
	if err := json.NewDecoder(gzReader).Decode(&idx); err != nil {
		return nil, "", fmt.Errorf("failed to parse search index: %w", err)
	}

	_ = saveSearchIndexCache(&idx)
	return &idx, RegistrySourceRemote, nil
}

// categoryCodeToName maps short category codes back to full names
var categoryCodeToName = map[string]string{
	"dev": "development",
	"ops": "devops",
	"sec": "security",
	"doc": "documents",
	"des": "design",
	"tst": "testing",
	"prd": "product",
	"mkt": "marketing",
	"pro": "productivity",
	"dat": "data",
	"off": "official",
	"oth": "other",
}

// entryToSkill converts a compact SearchIndexEntry to a full Skill
func entryToSkill(e SearchIndexEntry) Skill {
	// Resolve category code to full name
	cat := e.Category
	if full, ok := categoryCodeToName[cat]; ok {
		cat = full
	}

	// Extract repo from install path
	repo := e.Install
	if parts := strings.SplitN(repo, "/", 3); len(parts) >= 2 {
		repo = parts[0] + "/" + parts[1]
	}

	return Skill{
		Name:        e.Name,
		Description: e.Description,
		Install:     e.Install,
		Repo:        repo,
		Branch:      e.Branch,
		Category:    cat,
		Tags:        e.Tags,
		Stars:       e.Stars,
	}
}

// Search searches for skills matching the keyword
func Search(keyword string) ([]Skill, error) {
	skills, _, err := SearchWithSource(keyword)
	return skills, err
}

// SearchWithSource searches using the compact search index (9MB gzip vs 44MB full registry)
func SearchWithSource(keyword string) ([]Skill, RegistrySource, error) {
	idx, source, err := FetchSearchIndex()
	if err != nil {
		return nil, "", err
	}

	keyword = strings.ToLower(keyword)
	var results []Skill

	for _, entry := range idx.Skills {
		if strings.Contains(strings.ToLower(entry.Name), keyword) {
			results = append(results, entryToSkill(entry))
			continue
		}

		if strings.Contains(strings.ToLower(entry.Description), keyword) {
			results = append(results, entryToSkill(entry))
			continue
		}

		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), keyword) {
				results = append(results, entryToSkill(entry))
				break
			}
		}
	}

	return results, source, nil
}

// GetByCategory returns skills in a category
func GetByCategory(category string) ([]Skill, error) {
	cat, err := FetchCategory(category)
	if err != nil {
		// Fallback to filtering from full registry
		registry, err := FetchRegistry()
		if err != nil {
			return nil, err
		}

		var results []Skill
		for _, skill := range registry.Skills {
			if strings.EqualFold(skill.Category, category) {
				results = append(results, skill)
			}
		}
		return results, nil
	}

	return cat.Skills, nil
}

// ResolveInstall finds a skill by name and returns its install string.
func ResolveInstall(name string) (string, RegistrySource, error) {
	registry, source, err := FetchRegistryWithSource()
	if err != nil {
		return "", "", err
	}

	for _, skill := range registry.Skills {
		if strings.EqualFold(skill.Name, name) {
			return skill.Install, source, nil
		}
	}

	return "", source, fmt.Errorf("no skill named %q in registry", name)
}

func loadRegistryCache() (*Registry, error) {
	path := config.RegistryCachePath()
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	ttl := time.Duration(config.GetRegistryTTL()) * time.Hour
	if time.Since(info.ModTime()) > ttl {
		return nil, fmt.Errorf("registry cache expired")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	return &registry, nil
}

func saveRegistryCache(registry *Registry) error {
	path := config.RegistryCachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func searchIndexCachePath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(cacheDir, "sk", "search-index.json")
}

func loadSearchIndexCache() (*SearchIndex, error) {
	path := searchIndexCachePath()
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	ttl := time.Duration(config.GetRegistryTTL()) * time.Hour
	if time.Since(info.ModTime()) > ttl {
		return nil, fmt.Errorf("search index cache expired")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var idx SearchIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}

	return &idx, nil
}

func saveSearchIndexCache(idx *SearchIndex) error {
	path := searchIndexCachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(idx)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
