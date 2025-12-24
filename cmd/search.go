package cmd

import (
	"fmt"
	"strings"

	"github.com/majiayu000/caude-skill-manager/internal/registry"
	"github.com/majiayu000/caude-skill-manager/pkg/styles"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search [keyword]",
	Aliases: []string{"s", "find"},
	Short:   "Search for skills in the registry",
	Long: `Search for Claude Code skills in the registry.

Uses the skill-registry for fast and reliable results.`,
	Example: `  sk search testing
  sk search pdf
  sk search --category documents
  sk search --popular`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		popular, _ := cmd.Flags().GetBool("popular")
		category, _ := cmd.Flags().GetString("category")

		fmt.Println()

		// Show popular/featured skills
		if popular || (len(args) == 0 && category == "") {
			showFeaturedSkills()
			return
		}

		// Show by category
		if category != "" {
			showByCategory(category)
			return
		}

		// Search by keyword
		keyword := args[0]
		searchRegistry(keyword)
	},
}

func showFeaturedSkills() {
	fmt.Println(styles.TitleStyle.Render(styles.IconStar + " Popular Skills"))
	fmt.Println()

	// Try to fetch from registry
	reg, err := registry.FetchRegistry()
	if err != nil {
		fmt.Println(styles.WarningStyle.Render("Could not fetch registry: " + err.Error()))
		fmt.Println(styles.MutedStyle.Render("Showing cached data..."))
		fmt.Println()
		showFallbackSkills()
		return
	}

	// Group by source
	sources := make(map[string][]registry.Skill)
	for _, skill := range reg.Skills {
		sources[skill.Source] = append(sources[skill.Source], skill)
	}

	// Show featured first
	for source, skills := range sources {
		fmt.Printf("%s %s\n",
			styles.SkillNameStyle.Render(source),
			styles.MutedStyle.Render(fmt.Sprintf("(%d skills)", len(skills))),
		)
		fmt.Println()

		for _, skill := range skills {
			starStr := ""
			if skill.Stars > 0 {
				starStr = fmt.Sprintf(" %s%d", styles.IconStar, skill.Stars)
			}
			featuredStr := ""
			if skill.Featured {
				featuredStr = " " + styles.BadgeStyle.Render("featured")
			}

			fmt.Printf("    %s %-22s%s%s\n",
				styles.SuccessStyle.Render(styles.IconPackage),
				skill.Name,
				styles.MutedStyle.Render(starStr),
				featuredStr,
			)
			if skill.Description != "" {
				desc := skill.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				fmt.Printf("       %s\n", styles.SkillDescStyle.Render(desc))
			}
			fmt.Printf("       %s sk install %s\n\n",
				styles.MutedStyle.Render(styles.IconArrow),
				skill.Install,
			)
		}
	}

	fmt.Println(styles.MutedStyle.Render("─────────────────────────────────────────────────"))
	fmt.Printf("%s Registry: %d skills | Updated: %s\n",
		styles.MutedStyle.Render(styles.IconInfo),
		reg.TotalCount,
		reg.UpdatedAt[:10],
	)
	fmt.Println()
}

func showFallbackSkills() {
	// Hardcoded fallback when registry is unavailable
	skills := []struct {
		name, install, desc string
	}{
		{"docx", "anthropics/skills/docx", "Document creation and editing"},
		{"pdf", "anthropics/skills/pdf", "PDF document manipulation"},
		{"pptx", "anthropics/skills/pptx", "PowerPoint presentations"},
		{"superpowers", "obra/superpowers", "20+ battle-tested skills"},
	}

	for _, s := range skills {
		fmt.Printf("    %s %-22s\n",
			styles.SuccessStyle.Render(styles.IconPackage),
			s.name,
		)
		fmt.Printf("       %s\n", styles.SkillDescStyle.Render(s.desc))
		fmt.Printf("       %s sk install %s\n\n",
			styles.MutedStyle.Render(styles.IconArrow),
			s.install,
		)
	}
}

func showByCategory(category string) {
	fmt.Printf("%s Category: %s\n\n",
		styles.TitleStyle.Render(styles.IconFolder),
		styles.CodeStyle.Render(category),
	)

	skills, err := registry.GetByCategory(category)
	if err != nil {
		fmt.Println(styles.RenderError("Failed to fetch category: " + err.Error()))
		return
	}

	if len(skills) == 0 {
		fmt.Println(styles.MutedStyle.Render("No skills found in this category."))
		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("Available categories: documents, development, design, testing, productivity, data"))
		return
	}

	for _, skill := range skills {
		fmt.Printf("  %s %s\n",
			styles.SuccessStyle.Render(styles.IconPackage),
			styles.SkillNameStyle.Render(skill.Name),
		)
		if skill.Description != "" {
			desc := skill.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("     %s\n", styles.SkillDescStyle.Render(desc))
		}
		fmt.Printf("     %s sk install %s\n\n",
			styles.MutedStyle.Render(styles.IconArrow),
			skill.Install,
		)
	}

	fmt.Printf("\n%s Found %d skill(s) in '%s'\n\n",
		styles.MutedStyle.Render(styles.IconInfo),
		len(skills),
		category,
	)
}

func searchRegistry(keyword string) {
	fmt.Printf("%s Searching for '%s'...\n\n",
		styles.SpinnerStyle.Render(styles.IconSearch),
		styles.CodeStyle.Render(keyword),
	)

	skills, err := registry.Search(keyword)
	if err != nil {
		fmt.Println(styles.RenderError("Search failed: " + err.Error()))
		return
	}

	if len(skills) == 0 {
		fmt.Println(styles.MutedStyle.Render("No skills found matching your query."))
		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("Try:"))
		fmt.Println(styles.MutedStyle.Render("  • sk search --popular"))
		fmt.Println(styles.MutedStyle.Render("  • sk search --category documents"))
		fmt.Println(styles.MutedStyle.Render("  • Browse: https://skillsmp.com"))
		fmt.Println()
		return
	}

	fmt.Printf("%s Found %d skill(s):\n\n",
		styles.SuccessStyle.Render(styles.IconCheck),
		len(skills),
	)

	for i, skill := range skills {
		// Highlight matched keyword
		name := skill.Name
		desc := skill.Description

		fmt.Printf("%s %s",
			styles.SuccessStyle.Render(fmt.Sprintf("%2d.", i+1)),
			styles.SkillNameStyle.Render(name),
		)

		if skill.Stars > 0 {
			fmt.Printf("  %s %d",
				styles.MutedStyle.Render(styles.IconStar),
				skill.Stars,
			)
		}

		if skill.Featured {
			fmt.Printf(" %s", styles.BadgeStyle.Render("featured"))
		}

		fmt.Println()

		if desc != "" {
			if len(desc) > 70 {
				desc = desc[:67] + "..."
			}
			fmt.Printf("    %s\n", styles.SkillDescStyle.Render(desc))
		}

		// Tags
		if len(skill.Tags) > 0 {
			tags := strings.Join(skill.Tags, ", ")
			if len(tags) > 50 {
				tags = tags[:47] + "..."
			}
			fmt.Printf("    %s %s\n",
				styles.MutedStyle.Render("tags:"),
				styles.MutedStyle.Render(tags),
			)
		}

		fmt.Printf("    %s sk install %s\n\n",
			styles.MutedStyle.Render(styles.IconArrow),
			skill.Install,
		)
	}
}

func init() {
	searchCmd.Flags().BoolP("popular", "p", false, "Show popular/featured skills")
	searchCmd.Flags().StringP("category", "c", "", "Filter by category (documents, development, design, testing)")
	rootCmd.AddCommand(searchCmd)
}
