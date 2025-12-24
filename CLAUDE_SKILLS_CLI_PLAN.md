# Claude Skills CLI - 项目计划

> npm for Claude Skills - 让 Skills 管理像包管理一样简单

## 项目定位

### 一句话描述

**终端原生的 Claude Skills 管理器，专注 SkillsMP 没有的批量管理、版本控制和团队协作能力。**

### 核心价值主张

```
SkillsMP = 发现 Skills 的最佳网站
claude-skills = 管理 Skills 的最佳工具
```

### 目标用户

1. **Claude Code 重度用户** - 安装了多个 skills，需要管理
2. **团队/企业用户** - 需要统一团队的 skills 配置
3. **Skills 开发者** - 需要测试、发布自己的 skills

---

## 竞品分析

### SkillsMP 现状

| 功能 | SkillsMP | 我们 |
|------|----------|------|
| Skills 发现/搜索 | ✅ 网页 + AI 搜索 | ✅ CLI 搜索 |
| 单个安装 | ✅ /plugin marketplace | ✅ 命令行 |
| 批量安装 | ❌ Coming soon | ✅ 核心功能 |
| 版本管理 | ❌ | ✅ 核心功能 |
| 更新检测 | ❌ | ✅ 核心功能 |
| 团队同步 | ❌ | ✅ 差异化 |
| 冲突检测 | ❌ | ✅ 差异化 |
| 离线使用 | ❌ | ✅ 本地缓存 |

### 差异化策略

**不竞争，做互补** - 承认 SkillsMP 在发现层的优势，专注管理层。

---

## 产品设计

### 命令设计

```bash
# 项目名: sk (skill 的缩写，简短好记)

# ============ 基础命令 ============

# 搜索 skills
sk search <keyword>
sk search "testing" --lang rust
sk search --trending

# 安装 skills
sk install <source>
sk install anthropics/skills/docx          # 从 GitHub
sk install https://github.com/user/repo    # 从 URL
sk install ./local-skill                   # 从本地

# 列出已安装
sk list
sk list --outdated                         # 显示可更新的

# 卸载
sk uninstall <skill-name>

# 查看详情
sk info <skill-name>

# ============ 差异化命令 ============

# 批量更新（核心差异）
sk update                                  # 更新所有
sk update <skill-name>                     # 更新指定
sk update --dry-run                        # 预览更新

# 团队同步（核心差异）
sk init                                    # 初始化 skills.json
sk sync                                    # 从 skills.json 同步安装
sk export                                  # 导出当前配置到 skills.json
sk lock                                    # 锁定版本到 skills.lock

# 健康检查（核心差异）
sk doctor                                  # 检测问题
  - 检测 skills 冲突
  - 检测损坏的 skills
  - 检测过期版本
  - 建议优化

# 开发者命令
sk create <name>                           # 创建 skill 模板
sk validate                                # 验证 skill 格式
sk publish                                 # 发布到 registry（未来）

# ============ 配置 ============

sk config list                             # 查看配置
sk config set <key> <value>                # 设置配置
sk config get <key>                        # 获取配置

# 配置项:
# - skills_dir: ~/.claude/skills (默认)
# - registry: github (默认) | skillsmp
# - auto_update: false (默认)
```

### 配置文件设计

#### skills.json（项目级配置）

```json
{
  "name": "my-project",
  "skills": {
    "anthropics/skills/docx": "^1.0.0",
    "obra/superpowers": "latest",
    "local:./custom-skill": "*"
  },
  "devSkills": {
    "testing-utils": "^2.0.0"
  }
}
```

#### skills.lock（锁定文件）

```json
{
  "lockVersion": 1,
  "skills": {
    "anthropics/skills/docx": {
      "version": "1.2.3",
      "resolved": "https://github.com/anthropics/skills/tree/abc123",
      "integrity": "sha256-xxx"
    }
  }
}
```

#### ~/.skrc（全局配置）

```toml
[default]
skills_dir = "~/.claude/skills"
registry = "github"

[alias]
i = "install"
u = "update"
s = "search"

[cache]
ttl = 86400  # 24 hours
```

---

## 技术方案

### 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 语言 | **Rust** | 单二进制、快、无依赖、你熟悉 |
| CLI 框架 | clap | Rust 生态标准 |
| HTTP | reqwest | 成熟稳定 |
| JSON | serde | Rust 标准 |
| Git 操作 | git2 或 调用 git | 处理 GitHub 仓库 |
| 模板 | tera | 生成 skill 模板 |

### 架构设计

```
sk/
├── src/
│   ├── main.rs              # 入口
│   ├── cli/                  # CLI 定义
│   │   ├── mod.rs
│   │   ├── search.rs
│   │   ├── install.rs
│   │   ├── update.rs
│   │   └── ...
│   ├── core/                 # 核心逻辑
│   │   ├── mod.rs
│   │   ├── skill.rs         # Skill 数据结构
│   │   ├── registry.rs      # 注册表抽象
│   │   ├── resolver.rs      # 版本解析
│   │   └── installer.rs     # 安装逻辑
│   ├── registry/             # 注册表实现
│   │   ├── mod.rs
│   │   ├── github.rs        # GitHub 源
│   │   └── skillsmp.rs      # SkillsMP 源（可选）
│   ├── config/               # 配置管理
│   │   ├── mod.rs
│   │   ├── global.rs        # ~/.skrc
│   │   └── project.rs       # skills.json
│   └── utils/                # 工具函数
│       ├── mod.rs
│       ├── fs.rs
│       ├── git.rs
│       └── hash.rs
├── Cargo.toml
├── README.md
└── SKILL.md                  # 作为 skill 被发现
```

### 数据流

```
用户命令
    ↓
CLI 解析 (clap)
    ↓
核心逻辑
    ↓
┌─────────────────────────────────┐
│  Registry (数据源)               │
│  ├── GitHub API                 │
│  ├── SkillsMP API (可选)        │
│  └── Local Cache                │
└─────────────────────────────────┘
    ↓
Installer (安装器)
    ↓
~/.claude/skills/
```

---

## 开发路线图

### Phase 1: MVP（1 周）

**目标**: 能用，解决基本安装问题

```markdown
核心功能:
- [x] sk install <github-url>     # 从 GitHub 安装
- [x] sk list                      # 列出已安装
- [x] sk uninstall <name>          # 卸载
- [x] sk info <name>               # 查看详情

基础设施:
- [x] CLI 框架搭建
- [x] 配置文件读写 (~/.skrc)
- [x] Skills 目录管理
- [x] 基本错误处理
```

**发布**: Reddit r/ClaudeAI 测试

---

### Phase 2: 差异化功能（1 周）

**目标**: 建立护城河，做 SkillsMP 没有的

```markdown
核心功能:
- [ ] sk update [--all]            # 批量更新 ⭐
- [ ] sk init                      # 初始化 skills.json
- [ ] sk sync                      # 从配置同步安装 ⭐
- [ ] sk export                    # 导出配置
- [ ] sk doctor                    # 健康检查 ⭐

增强:
- [ ] 版本检测和比较
- [ ] 本地缓存机制
- [ ] 更新通知
```

**发布**: Hacker News Show HN

---

### Phase 3: 搜索与发现（1 周）

**目标**: 完整的包管理体验

```markdown
核心功能:
- [ ] sk search <keyword>          # 搜索 skills
- [ ] sk search --trending         # 热门 skills
- [ ] sk install <short-name>      # 短名称安装

数据源:
- [ ] GitHub Search API 集成
- [ ] 本地索引缓存
- [ ] (可选) SkillsMP API 集成
```

---

### Phase 4: 开发者体验（1 周）

**目标**: 帮助 skill 开发者

```markdown
核心功能:
- [ ] sk create <name>             # 创建 skill 模板
- [ ] sk validate                  # 验证 SKILL.md
- [ ] sk test                      # 测试 skill
- [ ] sk link                      # 软链接开发中的 skill

文档:
- [ ] 完善 README
- [ ] 添加 CONTRIBUTING.md
- [ ] 添加示例和教程
```

---

### Phase 5: 生态建设（持续）

```markdown
- [ ] 版本锁定 (skills.lock)
- [ ] 依赖解析
- [ ] 私有 registry 支持
- [ ] CI/CD 集成
- [ ] VS Code 扩展
- [ ] 官方 registry（如果做大了）
```

---

## 营销策略

### 发布节奏

```
Week 1: MVP 完成 → Reddit r/ClaudeAI 软发布
Week 2: 差异化功能 → Hacker News Show HN
Week 3: 完善搜索 → Product Hunt
Week 4: 持续迭代 → 技术博客、Twitter
```

### README 结构

```markdown
# sk - Claude Skills 管理器

> npm for Claude Skills

## 为什么用 sk？

- 🚀 **一键安装** - `sk install user/skill`
- 🔄 **批量更新** - `sk update` 更新所有 skills
- 👥 **团队同步** - `sk sync` 统一团队配置
- 🔍 **智能搜索** - `sk search "testing"`
- 🩺 **健康检查** - `sk doctor` 发现问题

## 30 秒上手

[安装命令]
[演示 GIF]
[基础用法]

## vs SkillsMP

SkillsMP 是发现 skills 的最佳网站。
sk 是管理 skills 的最佳工具。
它们是互补的。

## 完整文档

...
```

### 传播渠道

| 渠道 | 时机 | 内容重点 |
|------|------|----------|
| Reddit r/ClaudeAI | MVP | "我做了个 CLI 工具" |
| Hacker News | 差异化功能后 | 技术深度 + Show HN |
| X/Twitter | 持续 | 功能更新、使用技巧 |
| V2EX | 中文版后 | 中文社区 |
| 掘金/知乎 | 稳定后 | 教程文章 |

---

## 风险与应对

| 风险 | 可能性 | 应对 |
|------|--------|------|
| SkillsMP 做了官方 CLI | 中 | 专注差异化（团队同步、doctor） |
| Anthropic 官方出工具 | 低 | 早期建立用户基础，快速迭代 |
| Claude Code 改 skills 机制 | 中 | 抽象 adapter，快速适配 |
| 用户量不够 | 中 | 内容营销，教程引流 |

---

## 成功指标

### Phase 1 目标

- [ ] GitHub Stars: 100+
- [ ] Reddit 帖子: 50+ upvotes
- [ ] 周活用户: 50+

### Phase 2 目标

- [ ] GitHub Stars: 500+
- [ ] HN 首页
- [ ] 周活用户: 200+

### 长期目标

- [ ] GitHub Stars: 2000+
- [ ] 成为 Claude Code 用户的标配工具
- [ ] 被 SkillsMP 或 Anthropic 推荐

---

## 立即行动

### 今天

1. 创建 GitHub 仓库
2. 搭建 Rust 项目骨架
3. 实现 `sk list` 和 `sk install`

### 本周

1. 完成 MVP 全部功能
2. 写好 README（带 GIF）
3. Reddit 软发布

---

## 备选方案

如果 Rust 开发周期太长，可以考虑：

### Plan B: TypeScript + Bun

```bash
# 更快的开发速度
# 可以复用 npm 生态
# 缺点: 需要 runtime
```

### Plan C: Go

```bash
# 单二进制
# 开发速度比 Rust 快
# 缺点: 二进制稍大
```

---

*Let's ship it! 🚀*
