package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/majiayu000/caude-skill-manager/internal/config"
	"github.com/majiayu000/caude-skill-manager/internal/skill"
	"github.com/majiayu000/caude-skill-manager/pkg/styles"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check skills health",
	Long:  `Run diagnostics to check for common issues with your skills setup.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println(styles.TitleStyle.Render(styles.IconGear + " Skills Health Check"))
		fmt.Println()

		issues := 0

		// Check skills directory exists
		skillsDir := config.GetSkillsDir()
		if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
			fmt.Printf("  %s Skills directory does not exist: %s\n",
				styles.WarningStyle.Render(styles.IconWarning),
				skillsDir,
			)
			fmt.Printf("    %s Run %s to create it\n",
				styles.MutedStyle.Render(styles.IconArrow),
				styles.CodeStyle.Render("sk install <skill>"),
			)
			issues++
		} else {
			fmt.Printf("  %s Skills directory: %s\n",
				styles.SuccessStyle.Render(styles.IconCheck),
				skillsDir,
			)
		}

		// Check installed skills
		skills, err := skill.List()
		if err != nil {
			fmt.Printf("  %s Failed to list skills: %s\n",
				styles.ErrorStyle.Render(styles.IconCross),
				err.Error(),
			)
			issues++
		} else {
			fmt.Printf("  %s Installed skills: %d\n",
				styles.SuccessStyle.Render(styles.IconCheck),
				len(skills),
			)

			// Check each skill for issues
			for _, s := range skills {
				skillIssues := checkSkillHealth(s)
				if len(skillIssues) > 0 {
					fmt.Printf("\n  %s %s has issues:\n",
						styles.WarningStyle.Render(styles.IconWarning),
						s.Name,
					)
					for _, issue := range skillIssues {
						fmt.Printf("    %s %s\n",
							styles.MutedStyle.Render(styles.IconArrow),
							issue,
						)
					}
					issues += len(skillIssues)
				}
			}
		}

		// Summary
		fmt.Println()
		if issues == 0 {
			fmt.Println(styles.SuccessStyle.Render("  All checks passed! Your skills setup is healthy."))
		} else {
			fmt.Printf(styles.WarningStyle.Render("  Found %d issue(s). See above for details.\n"), issues)
		}
		fmt.Println()
	},
}

func checkSkillHealth(s skill.Skill) []string {
	var issues []string

	// Check SKILL.md exists
	skillMdPath := filepath.Join(s.Path, "SKILL.md")
	if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
		issues = append(issues, "Missing SKILL.md file")
	}

	// Check if directory is empty
	entries, err := os.ReadDir(s.Path)
	if err != nil {
		issues = append(issues, "Cannot read skill directory")
	} else if len(entries) == 0 {
		issues = append(issues, "Skill directory is empty")
	}

	// Check for description
	if s.Description == "" {
		issues = append(issues, "No description in SKILL.md")
	}

	return issues
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
