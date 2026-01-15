# sk - Claude Skills Manager

<p align="center">
  <a href="https://github.com/majiayu000/caude-skill-manager/releases"><img src="https://img.shields.io/github/v/release/majiayu000/caude-skill-manager?style=flat-square" alt="Release"></a>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square" alt="PRs Welcome">
</p>

> **npm for Claude Code Skills** — The package manager for Claude Code skills

```
   _____ __ __
  / ___// //_/
  \__ \/ ,<
 ___/ / /| |
/____/_/ |_|
```

## Why sk?

- 🚀 **One-command install** — `sk install user/repo`
- 🔄 **Batch update** — `sk update` (coming soon)
- 🔍 **Smart search** — `sk search testing`
- 🩺 **Health check** — `sk doctor` to find issues
- 🎨 **Beautiful TUI** — Built with [Charm](https://charm.sh)

## Quick Start

### Install

```bash
# Using Go
go install github.com/majiayu000/caude-skill-manager@latest

# Or download from releases
# macOS (Apple Silicon)
curl -L https://github.com/majiayu000/caude-skill-manager/releases/latest/download/sk_darwin_arm64.tar.gz | tar xz
sudo mv sk /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/majiayu000/caude-skill-manager/releases/latest/download/sk_darwin_amd64.tar.gz | tar xz
sudo mv sk /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/majiayu000/caude-skill-manager/releases/latest/download/sk_linux_amd64.tar.gz | tar xz
sudo mv sk /usr/local/bin/
```

### Usage

```bash
# Install a skill from GitHub
sk install anthropics/skills/docx
sk install obra/superpowers
sk install https://github.com/user/repo

# Install a skill by registry name
sk install docx

# List installed skills
sk list

# Search for skills
sk search           # Show popular skills
sk search testing   # Search by keyword

# Get skill details
sk info my-skill

# Remove a skill
sk uninstall my-skill

# Check health
sk doctor

# Update skills
sk update
```

## Demo

```
$ sk list

📦 Installed Skills

  NAME                       DESCRIPTION
───────────────────────────────────────────────────────────────────────────────
  docx                       Comprehensive document creation and editing...
  frontend-design            Create distinctive, production-grade frontend...
  superpowers                20+ battle-tested skills for Claude Code

  3 skill(s) installed
```

```
$ sk search

★ Popular Skills

Anthropic Official (anthropics/skills)
  Official Claude Code skills from Anthropic

    📦 docx                      sk install anthropics/skills/docx
    📦 pdf                       sk install anthropics/skills/pdf
    📦 pptx                      sk install anthropics/skills/pptx
    ...
```

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `sk install <source>` | `i`, `add` | Install a skill from GitHub |
| `sk list` | `ls`, `l` | List installed skills |
| `sk search [keyword]` | `s`, `find` | Search for skills |
| `sk info <name>` | `show` | Show skill details |
| `sk uninstall <name>` | `rm`, `remove` | Remove a skill |
| `sk update [name]` | `up` | Update skills |
| `sk doctor` | - | Check skills health |

## Supported Sources

```bash
# From registry by name
sk install docx

# Short format
sk install owner/repo              # Entire repo
sk install owner/repo/path         # Subdirectory (monorepo)

# Full URL
sk install https://github.com/owner/repo
sk install https://github.com/owner/repo/tree/main/path

# Examples
sk install anthropics/skills/docx  # Official Anthropic skill
sk install obra/superpowers        # Community skill
```

## vs SkillsMP

[SkillsMP](https://skillsmp.com) is the best website to **discover** skills.

`sk` is the best tool to **manage** skills.

They're complementary:
1. Find skills on SkillsMP
2. Install & manage with `sk`

## Configuration

Skills are installed to `~/.claude/skills/` by default.

Config file: `~/.skrc`

```json
{
  "skills_dir": "~/.claude/skills",
  "registry": "https://raw.githubusercontent.com/majiayu000/claude-skill-registry/main",
  "registry_ttl_hours": 24
}
```

Registry cache:
- Location: `~/.cache/sk/registry.json`
- TTL: `registry_ttl_hours` (cache is ignored after expiry)

## Built With

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions
- [Huh](https://github.com/charmbracelet/huh) — Form components

## Contributing

PRs welcome!

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  Made with ❤️ for the Claude Code community
</p>
