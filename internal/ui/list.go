package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/majiayu000/caude-skill-manager/internal/skill"
	"github.com/majiayu000/caude-skill-manager/pkg/styles"
)

// SkillItem represents a skill in the list
type SkillItem struct {
	skill skill.Skill
}

func (i SkillItem) Title() string       { return i.skill.Name }
func (i SkillItem) Description() string { return i.skill.Description }
func (i SkillItem) FilterValue() string { return i.skill.Name }

// SkillListModel is the main list model
type SkillListModel struct {
	list     list.Model
	quitting bool
}

// NewSkillList creates a new skill list UI
func NewSkillList(skills []skill.Skill, width, height int) SkillListModel {
	items := make([]list.Item, len(skills))
	for i, s := range skills {
		items[i] = SkillItem{skill: s}
	}

	// Custom delegate for beautiful rendering
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.Primary).
		BorderLeftForeground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.TextDim).
		BorderLeftForeground(styles.Primary)

	l := list.New(items, delegate, width, height)
	l.Title = "Installed Skills"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.TitleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(styles.Primary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(styles.Primary)

	return SkillListModel{list: l}
}

func (m SkillListModel) Init() tea.Cmd {
	return nil
}

func (m SkillListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SkillListModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// RenderSkillCard renders a skill as a beautiful card
func RenderSkillCard(s skill.Skill) string {
	var b strings.Builder

	// Name with icon
	name := styles.SkillNameStyle.Render(styles.IconPackage + " " + s.Name)
	b.WriteString(name)
	b.WriteString("\n")

	// Description
	if s.Description != "" {
		desc := styles.SkillDescStyle.Render(s.Description)
		b.WriteString(desc)
		b.WriteString("\n")
	}

	// Meta info
	meta := []string{}
	if s.Source != "" {
		meta = append(meta, styles.IconArrow+" "+s.Source)
	}
	if !s.InstalledAt.IsZero() {
		meta = append(meta, "installed "+s.InstalledAt.Format("2006-01-02"))
	}
	if len(meta) > 0 {
		metaStr := styles.SkillMetaStyle.Render(strings.Join(meta, "  "))
		b.WriteString(metaStr)
	}

	return styles.SkillCardStyle.Render(b.String())
}

// RenderSkillTable renders skills as a table
func RenderSkillTable(skills []skill.Skill) string {
	if len(skills) == 0 {
		return styles.MutedStyle.Render("No skills installed yet.\n\nRun ") +
			styles.CodeStyle.Render("sk install <github-url>") +
			styles.MutedStyle.Render(" to install your first skill.")
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-25s  %-50s", "NAME", "DESCRIPTION")
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Rows
	for _, s := range skills {
		name := s.Name
		if len(name) > 25 {
			name = name[:22] + "..."
		}

		desc := s.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		if desc == "" {
			desc = styles.MutedStyle.Render("(no description)")
		}

		row := fmt.Sprintf("  %-25s  %-50s",
			styles.SuccessStyle.Render(name),
			styles.SkillDescStyle.Render(desc),
		)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("  %d skill(s) installed", len(skills))))

	return b.String()
}
