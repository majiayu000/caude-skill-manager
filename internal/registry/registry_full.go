package registry

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type registrySchemaVersion string

type registryManifest struct {
	SchemaVersion registrySchemaVersion `json:"schema_version"`
	GeneratedAt   string                `json:"generated_at"`
	TotalCount    int                   `json:"total_count"`
	Shards        []artifactPart        `json:"shards"`
}

type registryShard struct {
	SchemaVersion registrySchemaVersion `json:"schema_version"`
	GeneratedAt   string                `json:"generated_at"`
	Shard         string                `json:"shard"`
	Count         int                   `json:"count"`
	Skills        []Skill               `json:"skills"`
}

func (v *registrySchemaVersion) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*v = registrySchemaVersion(text)
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*v = registrySchemaVersion(number.String())
		return nil
	}

	return fmt.Errorf("invalid schema_version %s", strings.TrimSpace(string(data)))
}

func fetchRegistryFromBaseURL(baseURL string) (*Registry, error) {
	var registry Registry
	if err := fetchJSON(artifactURL(baseURL, "registry.json"), &registry); err != nil {
		return nil, err
	}

	if registry.DeprecatedFullPayload && len(registry.Skills) == 0 {
		if registry.Manifest == "" {
			return nil, fmt.Errorf("registry pointer is missing manifest")
		}
		return fetchRegistryFromManifest(baseURL, registry.Manifest, &registry)
	}

	for i := range registry.Skills {
		normalizeRegistrySkill(&registry.Skills[i])
	}
	return &registry, nil
}

func fetchRegistryFromManifest(baseURL, manifestPath string, pointer *Registry) (*Registry, error) {
	var manifest registryManifest
	if err := fetchJSON(artifactURL(baseURL, manifestPath), &manifest); err != nil {
		return nil, fmt.Errorf("failed to fetch registry manifest %s: %w", manifestPath, err)
	}

	registry := &Registry{
		Version:    string(manifest.SchemaVersion),
		UpdatedAt:  manifest.GeneratedAt,
		TotalCount: manifest.TotalCount,
	}
	if pointer != nil {
		if registry.Version == "" {
			registry.Version = pointer.Version
		}
		if registry.UpdatedAt == "" {
			registry.UpdatedAt = pointer.UpdatedAt
		}
		if registry.TotalCount == 0 {
			registry.TotalCount = pointer.TotalCount
		}
	}

	for _, shard := range manifest.Shards {
		shardPath := shard.GzipPath
		if shardPath == "" {
			shardPath = shard.Path
		}
		if shardPath == "" {
			return nil, fmt.Errorf("registry manifest contains empty shard path")
		}

		var payload registryShard
		if err := fetchJSON(artifactURL(baseURL, shardPath), &payload); err != nil {
			return nil, fmt.Errorf("failed to fetch registry shard %s: %w", shardPath, err)
		}
		for i := range payload.Skills {
			normalizeRegistrySkill(&payload.Skills[i])
			registry.Skills = append(registry.Skills, payload.Skills[i])
		}
	}

	if registry.TotalCount == 0 {
		registry.TotalCount = len(registry.Skills)
	}
	return registry, nil
}

func normalizeRegistrySkill(skill *Skill) {
	if skill.Install != "" {
		skill.Install = normalizeInstallForBranch(skill.Install, skill.Branch)
		return
	}
	if skill.Repo == "" {
		return
	}

	path := strings.Trim(skill.Path, "/")
	branch := skill.Branch
	if branch == "" {
		branch = "main"
	}
	if branch == "main" {
		if path == "" {
			skill.Install = skill.Repo
			return
		}
		skill.Install = skill.Repo + "/" + path
		return
	}
	if path == "" {
		skill.Install = fmt.Sprintf("https://github.com/%s/tree/%s", skill.Repo, url.PathEscape(branch))
		return
	}
	skill.Install = fmt.Sprintf("https://github.com/%s/tree/%s/%s", skill.Repo, url.PathEscape(branch), path)
}

func normalizeRegistrySkills(skills []Skill) {
	for i := range skills {
		normalizeRegistrySkill(&skills[i])
	}
}

func normalizeInstallForBranch(install, branch string) string {
	install = strings.TrimSpace(install)
	if install == "" {
		return ""
	}
	if branch == "" || branch == "main" {
		return install
	}
	if !strings.HasPrefix(install, "https://github.com/") {
		return githubTreeInstall(install, branch)
	}
	if strings.Contains(install, "/tree/") {
		return install
	}
	ref := strings.TrimPrefix(install, "https://github.com/")
	return githubTreeInstall(ref, branch)
}

func githubTreeInstall(ref, branch string) string {
	parts := strings.Split(strings.Trim(ref, "/"), "/")
	if len(parts) < 2 {
		return ref
	}

	repo := parts[0] + "/" + parts[1]
	if len(parts) == 2 {
		return fmt.Sprintf("https://github.com/%s/tree/%s", repo, url.PathEscape(branch))
	}
	path := strings.Join(parts[2:], "/")
	return fmt.Sprintf("https://github.com/%s/tree/%s/%s", repo, url.PathEscape(branch), path)
}

func repoFromInstallRef(install string) string {
	ref := strings.TrimSpace(install)
	ref = strings.TrimPrefix(ref, "https://github.com/")
	parts := strings.Split(ref, "/")
	if len(parts) < 2 {
		return ref
	}
	return parts[0] + "/" + parts[1]
}
