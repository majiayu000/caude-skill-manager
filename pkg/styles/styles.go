package styles

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors - Modern, vibrant color palette
var (
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#06B6D4") // Cyan
	Success   = lipgloss.Color("#10B981") // Green
	Warning   = lipgloss.Color("#F59E0B") // Amber
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	Text      = lipgloss.Color("#F9FAFB") // White
	TextDim   = lipgloss.Color("#9CA3AF") // Gray-400
	BgDark    = lipgloss.Color("#1F2937") // Gray-800
	BgDarker  = lipgloss.Color("#111827") // Gray-900
)

// Logo - ASCII art logo
var Logo = `
   _____ __ __
  / ___// //_/
  \__ \/ ,<
 ___/ / /| |
/____/_/ |_|
`

// Styles
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	// Subtle title
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			Italic(true)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Success)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Error)

	// Warning message
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	// Info/muted text
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Command/code style
	CodeStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Background(BgDark).
			Padding(0, 1)

	// Box style for content
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// List item styles
	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(Primary).
				Bold(true)

	// Badge styles
	BadgeStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Primary).
			Padding(0, 1).
			Bold(true)

	InstalledBadge = lipgloss.NewStyle().
			Foreground(Text).
			Background(Success).
			Padding(0, 1)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Primary)

	// Progress bar
	ProgressFilled = lipgloss.NewStyle().
			Foreground(Primary)

	ProgressEmpty = lipgloss.NewStyle().
			Foreground(Muted)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(Muted)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Skill card style
	SkillCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(1, 2).
			MarginBottom(1)

	SkillNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Text)

	SkillDescStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	SkillMetaStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)
)

// Icons (default to Unicode, can downgrade to ASCII)
var (
	IconCheck    = "✓"
	IconCross    = "✗"
	IconArrow    = "→"
	IconBullet   = "•"
	IconStar     = "★"
	IconBox      = "◼"
	IconCircle   = "●"
	IconDiamond  = "◆"
	IconSparkle  = "✨"
	IconPackage  = "📦"
	IconFolder   = "📁"
	IconFile     = "📄"
	IconSearch   = "🔍"
	IconDownload = "⬇"
	IconUpload   = "⬆"
	IconSync     = "🔄"
	IconWarning  = "⚠"
	IconInfo     = "ℹ"
	IconGear     = "⚙"
)

func init() {
	if !supportsUTF8() {
		IconCheck = "[ok]"
		IconCross = "[x]"
		IconArrow = "->"
		IconBullet = "*"
		IconStar = "*"
		IconBox = "#"
		IconCircle = "o"
		IconDiamond = "<>"
		IconSparkle = "*"
		IconPackage = "[pkg]"
		IconFolder = "[dir]"
		IconFile = "[file]"
		IconSearch = "[search]"
		IconDownload = "[dl]"
		IconUpload = "[up]"
		IconSync = "[sync]"
		IconWarning = "[!]"
		IconInfo = "[i]"
		IconGear = "[cfg]"
	}
}

func supportsUTF8() bool {
	envs := []string{"LC_ALL", "LC_CTYPE", "LANG"}
	for _, key := range envs {
		if v := os.Getenv(key); strings.Contains(strings.ToUpper(v), "UTF-8") || strings.Contains(strings.ToUpper(v), "UTF8") {
			return true
		}
	}
	return false
}

// Helper functions
func RenderSuccess(msg string) string {
	return SuccessStyle.Render(IconCheck+" ") + msg
}

func RenderError(msg string) string {
	return ErrorStyle.Render(IconCross+" ") + msg
}

func RenderWarning(msg string) string {
	return WarningStyle.Render(IconWarning+" ") + msg
}

func RenderInfo(msg string) string {
	return MutedStyle.Render(IconInfo+" ") + msg
}

func RenderBadge(text string) string {
	return BadgeStyle.Render(text)
}

func RenderInstalledBadge() string {
	return InstalledBadge.Render("installed")
}
