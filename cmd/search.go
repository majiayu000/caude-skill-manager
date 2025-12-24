package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/majiayu000/caude-skill-manager/pkg/styles"
)

// GitHubSearchResult represents GitHub search API response
type GitHubSearchResult struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Stars       int    `json:"stargazers_count"`
		URL         string `json:"html_url"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"items"`
}

// KnownSkillSource represents a known skills repository
type KnownSkillSource struct {
	Name        string
	Repo        string
	Description string
	Skills      []string
}

// Known skill sources
var knownSources = []KnownSkillSource{
	{
		Name:        "Anthropic Official",
		Repo:        "anthropics/skills",
		Description: "Official Claude Code skills from Anthropic",
		Skills: []string{
			"docx", "pdf", "pptx", "xlsx",
			"frontend-design", "canvas-design", "mcp-builder",
			"webapp-testing", "web-artifacts-builder",
			"brand-guidelines", "internal-comms",
			"algorithmic-art", "slack-gif-creator", "theme-factory",
			"skill-creator", "doc-coauthoring",
		},
	},
	{
		Name:        "Obra Superpowers",
		Repo:        "obra/superpowers",
		Description: "20+ battle-tested skills for Claude Code",
		Skills:      []string{"superpowers"},
	},
}

var searchCmd = &cobra.Command{
	Use:     "search <keyword>",
	Aliases: []string{"s", "find"},
	Short:   "Search for skills on GitHub",
	Long: `Search for Claude Code skills on GitHub.

Searches for repositories containing SKILL.md files.`,
	Example: `  sk search testing
  sk search "react component"
  sk search --popular`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		popular, _ := cmd.Flags().GetBool("popular")
		limit, _ := cmd.Flags().GetInt("limit")

		fmt.Println()

		// Show popular/known skills
		if popular || len(args) == 0 {
			showPopularSkills()
			return
		}

		keyword := strings.Join(args, " ")

		fmt.Printf("%s Searching for '%s'...\n\n",
			styles.SpinnerStyle.Render(styles.IconSearch),
			styles.CodeStyle.Render(keyword),
		)

		// First search in known sources
		localResults := searchKnownSources(keyword)
		if len(localResults) > 0 {
			fmt.Printf("%s Found %d matching skill(s) from known sources:\n\n",
				styles.SuccessStyle.Render(styles.IconCheck),
				len(localResults),
			)
			for i, result := range localResults {
				fmt.Printf("%s %s\n",
					styles.SuccessStyle.Render(fmt.Sprintf("%2d.", i+1)),
					styles.SkillNameStyle.Render(result),
				)
				fmt.Printf("    %s sk install %s\n\n",
					styles.MutedStyle.Render(styles.IconArrow),
					result,
				)
			}
		}

		// Then search GitHub
		results, err := searchGitHub(keyword, limit)
		if err != nil {
			// Don't fail, just show warning
			fmt.Println(styles.WarningStyle.Render("GitHub search unavailable: " + err.Error()))
		} else if len(results.Items) > 0 {
			fmt.Printf("%s Found %d repo(s) on GitHub:\n\n",
				styles.SuccessStyle.Render(styles.IconCheck),
				len(results.Items),
			)

			for i, item := range results.Items {
				fmt.Printf("%s %s",
					styles.SuccessStyle.Render(fmt.Sprintf("%2d.", i+1)),
					styles.SkillNameStyle.Render(item.FullName),
				)
				fmt.Printf("  %s %d\n",
					styles.MutedStyle.Render(styles.IconStar),
					item.Stars,
				)
				if item.Description != "" {
					desc := item.Description
					if len(desc) > 70 {
						desc = desc[:67] + "..."
					}
					fmt.Printf("    %s\n", styles.SkillDescStyle.Render(desc))
				}
				fmt.Printf("    %s sk install %s\n\n",
					styles.MutedStyle.Render(styles.IconArrow),
					item.FullName,
				)
			}
		}

		if len(localResults) == 0 && (err != nil || len(results.Items) == 0) {
			fmt.Println(styles.MutedStyle.Render("No skills found matching your query."))
			fmt.Println()
			fmt.Println(styles.MutedStyle.Render("Try:"))
			fmt.Println(styles.MutedStyle.Render("  • sk search --popular"))
			fmt.Println(styles.MutedStyle.Render("  • Browsing SkillsMP: https://skillsmp.com"))
		}

		fmt.Println()
	},
}

func showPopularSkills() {
	fmt.Println(styles.TitleStyle.Render(styles.IconStar + " Popular Skills"))
	fmt.Println()

	for _, source := range knownSources {
		fmt.Printf("%s %s\n",
			styles.SkillNameStyle.Render(source.Name),
			styles.MutedStyle.Render("("+source.Repo+")"),
		)
		fmt.Printf("  %s\n\n", styles.SkillDescStyle.Render(source.Description))

		for _, skill := range source.Skills {
			installCmd := source.Repo
			if skill != source.Repo && !strings.HasSuffix(source.Repo, skill) {
				installCmd = source.Repo + "/" + skill
			}
			fmt.Printf("    %s %-25s %s\n",
				styles.SuccessStyle.Render(styles.IconPackage),
				skill,
				styles.MutedStyle.Render("sk install "+installCmd),
			)
		}
		fmt.Println()
	}

	fmt.Println(styles.MutedStyle.Render("─────────────────────────────────────────────────"))
	fmt.Printf("%s More skills at: %s\n",
		styles.MutedStyle.Render(styles.IconInfo),
		styles.CodeStyle.Render("https://skillsmp.com"),
	)
	fmt.Println()
}

func searchKnownSources(keyword string) []string {
	keyword = strings.ToLower(keyword)
	var results []string

	for _, source := range knownSources {
		for _, skill := range source.Skills {
			if strings.Contains(strings.ToLower(skill), keyword) {
				results = append(results, source.Repo+"/"+skill)
			}
		}
	}

	return results
}

func searchGitHub(keyword string, limit int) (*GitHubSearchResult, error) {
	// Search for repos with SKILL.md file or claude-skill topic
	query := fmt.Sprintf("%s SKILL.md OR topic:claude-skill OR topic:claude-code-skill", keyword)
	searchURL := fmt.Sprintf(
		"https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=%d",
		url.QueryEscape(query),
		limit,
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "sk-claude-skills-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func init() {
	searchCmd.Flags().IntP("limit", "l", 10, "Maximum number of results")
	searchCmd.Flags().BoolP("popular", "p", false, "Show popular skills")
	rootCmd.AddCommand(searchCmd)
}
