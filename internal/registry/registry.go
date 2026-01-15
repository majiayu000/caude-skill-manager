package registry

import (
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
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Source      string   `json:"source"`
	Stars       int      `json:"stars"`
	Featured    bool     `json:"featured"`
}

// Category represents a category index
type Category struct {
	Category  string  `json:"category"`
	UpdatedAt string  `json:"updated_at"`
	Count     int     `json:"count"`
	Skills    []Skill `json:"skills"`
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

// FetchCategory fetches skills for a specific category
func FetchCategory(category string) (*Category, error) {
	url := fmt.Sprintf("%s/categories/%s.json", registryBaseURL(), category)

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

// Search searches for skills matching the keyword
func Search(keyword string) ([]Skill, error) {
	registry, err := FetchRegistry()
	if err != nil {
		return nil, err
	}

	keyword = strings.ToLower(keyword)
	var results []Skill

	for _, skill := range registry.Skills {
		// Match by name
		if strings.Contains(strings.ToLower(skill.Name), keyword) {
			results = append(results, skill)
			continue
		}

		// Match by description
		if strings.Contains(strings.ToLower(skill.Description), keyword) {
			results = append(results, skill)
			continue
		}

		// Match by tags
		for _, tag := range skill.Tags {
			if strings.Contains(strings.ToLower(tag), keyword) {
				results = append(results, skill)
				break
			}
		}
	}

	return results, nil
}

// SearchWithSource searches for skills and returns the data source.
func SearchWithSource(keyword string) ([]Skill, RegistrySource, error) {
	registry, source, err := FetchRegistryWithSource()
	if err != nil {
		return nil, "", err
	}

	keyword = strings.ToLower(keyword)
	var results []Skill

	for _, skill := range registry.Skills {
		if strings.Contains(strings.ToLower(skill.Name), keyword) {
			results = append(results, skill)
			continue
		}

		if strings.Contains(strings.ToLower(skill.Description), keyword) {
			results = append(results, skill)
			continue
		}

		for _, tag := range skill.Tags {
			if strings.Contains(strings.ToLower(tag), keyword) {
				results = append(results, skill)
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
