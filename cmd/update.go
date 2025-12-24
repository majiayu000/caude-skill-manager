package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/majiayu000/caude-skill-manager/internal/skill"
	"github.com/majiayu000/caude-skill-manager/pkg/styles"
)

var updateCmd = &cobra.Command{
	Use:     "update [skill-name]",
	Aliases: []string{"up", "upgrade"},
	Short:   "Update installed skills",
	Long: `Update one or all installed skills to their latest versions.

If no skill name is provided, all skills will be updated.`,
	Example: `  sk update           # Update all skills
  sk update my-skill  # Update specific skill`,
	Run: func(cmd *cobra.Command, args []string) {
		skills, err := skill.List()
		if err != nil {
			fmt.Println(styles.RenderError("Failed to list skills: " + err.Error()))
			return
		}

		if len(skills) == 0 {
			fmt.Println(styles.RenderWarning("No skills installed."))
			return
		}

		fmt.Println()
		fmt.Println(styles.TitleStyle.Render(styles.IconSync + " Checking for updates..."))
		fmt.Println()

		// TODO: Implement actual update logic
		// For now, just show the installed skills
		for _, s := range skills {
			if len(args) > 0 && s.Name != args[0] {
				continue
			}
			fmt.Printf("  %s %s %s\n",
				styles.SuccessStyle.Render(styles.IconCheck),
				s.Name,
				styles.MutedStyle.Render("(already up to date)"),
			)
		}

		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("  Update functionality coming soon!"))
		fmt.Println(styles.MutedStyle.Render("  For now, use: sk uninstall <name> && sk install <source>"))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
