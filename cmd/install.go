package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/majiayu000/caude-skill-manager/internal/github"
	"github.com/majiayu000/caude-skill-manager/internal/skill"
	"github.com/majiayu000/caude-skill-manager/internal/ui"
	"github.com/majiayu000/caude-skill-manager/pkg/styles"
)

var (
	installName string // custom name for the skill
	installForce bool  // force reinstall
)

var installCmd = &cobra.Command{
	Use:     "install <source>",
	Aliases: []string{"i", "add"},
	Short:   "Install a skill from GitHub",
	Long: `Install a Claude Code skill from GitHub.

Supported formats:
  owner/repo                     Install entire repo
  owner/repo/path/to/skill       Install skill from subdirectory
  https://github.com/owner/repo  Full GitHub URL
`,
	Example: `  sk install anthropics/skills/docx
  sk install obra/superpowers
  sk install https://github.com/user/repo`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]

		// Parse GitHub URL
		info, err := github.ParseGitHubURL(source)
		if err != nil {
			fmt.Println(styles.RenderError(err.Error()))
			os.Exit(1)
		}

		// Determine skill name
		skillName := installName
		if skillName == "" {
			skillName = github.GetSkillName(info)
		}

		// Check if already installed
		if skill.Exists(skillName) && !installForce {
			fmt.Println(styles.RenderWarning(fmt.Sprintf("Skill '%s' is already installed.", skillName)))
			fmt.Println(styles.MutedStyle.Render("Use --force to reinstall."))
			os.Exit(1)
		}

		// Remove existing if force
		if installForce && skill.Exists(skillName) {
			skill.Remove(skillName)
		}

		fmt.Println()
		fmt.Printf("%s Installing %s\n", styles.SpinnerStyle.Render("⠋"), styles.CodeStyle.Render(skillName))
		fmt.Printf("  %s %s\n", styles.MutedStyle.Render("from"), info.FullURL)
		fmt.Println()

		// Download and install with spinner
		err = ui.RunWithSpinner("Downloading...", func() (string, error) {
			if err := github.DownloadAndExtract(info, skillName); err != nil {
				return "", err
			}

			// Get installed skill info
			s, _ := skill.Get(skillName)
			result := styles.RenderSuccess(fmt.Sprintf("Installed %s", styles.CodeStyle.Render(skillName)))
			if s != nil && s.Description != "" {
				result += "\n  " + styles.SkillDescStyle.Render(s.Description)
			}
			return result, nil
		})

		if err != nil {
			fmt.Println(styles.RenderError(err.Error()))
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("  Skill installed to: ") + skill.GetSkillDir(skillName))
		fmt.Println()
	},
}

func init() {
	installCmd.Flags().StringVarP(&installName, "name", "n", "", "Custom name for the skill")
	installCmd.Flags().BoolVarP(&installForce, "force", "f", false, "Force reinstall if already exists")
	rootCmd.AddCommand(installCmd)
}
