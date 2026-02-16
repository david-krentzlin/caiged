# Creating Spins

## What is a Spin?

A **spin** is a role-specific container configuration that tailors the AI agent environment for a particular task (QA, engineering, code review, security audit, etc.).

Each spin defines:
- **Agent instructions** (`AGENTS.md` or `AGENT.md`) - the persona, rules, and workflows
- **Skills** (`skills/`) - domain-specific capabilities loaded by OpenCode
- **MCP servers** (`mcp/`) - Model Context Protocol server configurations
- **Documentation** (`README.md`) - spin-specific usage notes

## Directory Structure

```
spins/
└── <spin-name>/
    ├── AGENTS.md          # Agent instructions (preferred, falls back to AGENT.md)
    ├── README.md          # Spin documentation
    ├── skills/            # OpenCode skills
    │   ├── skill-name-1/
    │   │   └── SKILL.md
    │   ├── skill-name-2/
    │   │   └── SKILL.md
    │   └── README.md      # Optional: skills overview
    └── mcp/               # MCP server configs
        ├── config.json    # MCP server definitions
        └── README.md      # Optional: MCP setup notes
```

## File Specifications

### `AGENTS.md` (or `AGENT.md`)

The agent instructions file defines:
- **Purpose**: What role does this agent play?
- **Hard rules**: Non-negotiable constraints (e.g., "no production code changes")
- **Process**: Step-by-step workflows
- **Tone and conduct**: How the agent should communicate
- **Decision criteria**: What makes a good/bad outcome?

**Example structure:**
```markdown
# Agent: <Role Name>

## Purpose
Brief description of what this agent does.

## Hard Rules (Non-Negotiable)
1. Rule one
2. Rule two

## Process
1. Step one
2. Step two

## Operating Mode
How the agent approaches tasks...
```

See `spins/qa/AGENTS.md` for a complete example.

### `skills/`

Skills extend the agent with specialized capabilities. Each skill lives in its own directory:

```
skills/
└── skill-name/
    └── SKILL.md
```

OpenCode loads skills from this directory. Each `SKILL.md` defines:
- Skill name
- Description
- When to use it
- How to invoke it
- Expected inputs/outputs

**Example skill structure:**
```markdown
# Skill: <Skill Name>

## Description
What this skill does...

## When to Use
Situations where this skill applies...

## Usage
How to invoke and what to expect...
```

See `spins/qa/skills/` for examples (security_review, performance_testing, etc.).

### `mcp/`

Model Context Protocol server configurations. This directory contains JSON files defining MCP servers the agent can use.

**Example `mcp/config.json`:**
```json
{
  "servers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-package"]
    }
  }
}
```

See OpenCode's MCP documentation for full configuration options.

### `README.md`

Spin-specific documentation covering:
- Quick overview of the spin's purpose
- Key skills and capabilities
- Onboarding instructions
- Usage notes and tips

## Creating a New Spin

### Step 1: Create the Directory Structure

```bash
cd /path/to/caiged
mkdir -p spins/<spin-name>/{skills,mcp}
touch spins/<spin-name>/AGENTS.md
touch spins/<spin-name>/README.md
touch spins/<spin-name>/skills/README.md
touch spins/<spin-name>/mcp/README.md
```

### Step 2: Write Agent Instructions

Edit `spins/<spin-name>/AGENTS.md` with:
- Clear role definition
- Hard rules and constraints
- Process and workflows
- Decision criteria

Use `spins/qa/AGENTS.md` as a template.

### Step 3: Add Skills (Optional)

For each skill:

```bash
mkdir -p spins/<spin-name>/skills/<skill-name>
touch spins/<spin-name>/skills/<skill-name>/SKILL.md
```

Write the skill definition in `SKILL.md`.

### Step 4: Configure MCP Servers (Optional)

If your spin needs MCP servers:

```bash
touch spins/<spin-name>/mcp/config.json
```

Define server configurations in JSON format.

### Step 5: Document the Spin

Edit `spins/<spin-name>/README.md` with:
- Purpose and use cases
- Key features
- Setup instructions
- Examples

## Build Process (Automatic)

**No code changes required!** The build process automatically handles new spins:

1. **CLI discovers spin**: When you run `caiged run ... --spin <name>` (or `caiged build ... --spin <name>`), the CLI checks for `spins/<name>/`
2. **Docker builds**: `Dockerfile` uses `ARG SPIN` and `COPY spins/${SPIN}/` to copy files
3. **Container setup**: `entrypoint.sh` copies spin assets to `/root/.config/opencode/`:
   - `AGENTS.md` → `/root/.config/opencode/AGENTS.md`
   - `skills/` → `/root/.config/opencode/skills/`
   - `mcp/` → `/root/.config/opencode/mcp/`

### Testing Your Spin

```bash
# Build the spin image
caiged build . --spin <spin-name>

# Run the spin
caiged run . --spin <spin-name>
```

### Example: Adding an "engineer" Spin

```bash
cd /path/to/caiged

# Create structure
mkdir -p spins/engineer/{skills,mcp}

# Write instructions
cat > spins/engineer/AGENTS.md <<'EOF'
# Agent: Software Engineer

## Purpose
Implement features, fix bugs, refactor code.

## Hard Rules
1. Write tests for all new code
2. Follow existing code style
3. No breaking changes without approval

## Process
1. Understand requirements
2. Design solution
3. Implement with tests
4. Verify locally
EOF

# Document the spin
cat > spins/engineer/README.md <<'EOF'
# Engineer Spin

Focused on feature implementation and bug fixes.
Includes standard engineering workflows.
EOF

# Build and run
caiged build . --spin engineer
caiged run . --spin engineer
```

## Spin Design Guidelines

### Good Spin Characteristics

✅ **Single, clear purpose**: QA, engineering, review, security, etc.  
✅ **Well-defined constraints**: What can/can't this agent do?  
✅ **Actionable workflows**: Step-by-step processes  
✅ **Measurable outcomes**: Clear success/failure criteria  
✅ **Focused skills**: Only include relevant capabilities  

### Anti-patterns

❌ **Kitchen sink spins**: Trying to do everything  
❌ **Vague instructions**: "Help with coding tasks"  
❌ **Missing constraints**: No guardrails on agent behavior  
❌ **Overlapping purposes**: Multiple spins doing the same thing  

## Maintenance

### Updating a Spin

Edit files in `spins/<name>/`, then rebuild:

```bash
caiged build . --spin <name>
```

### Removing a Spin

```bash
rm -rf spins/<name>
docker rmi caiged:<name>
```

## Contributing Spins

When contributing a new spin:

1. Follow the directory structure above
2. Include comprehensive `AGENTS.md` with clear examples
3. Document all skills with usage instructions
4. Test the spin on a real project
5. Submit a PR with:
   - Spin implementation in `spins/<name>/`
   - Updated `README.md` listing the new spin
   - Example usage in the PR description

## FAQ

**Q: Can I have spin-specific tools/dependencies?**  
A: Not directly. Spins share the base image. For tool differences, use conditional logic in scripts or skills.

**Q: Can spins share skills?**  
A: Skills are copied per-spin. To share, symlink or copy common skills into multiple spin directories.

**Q: How do I pass data between spins?**  
A: Spins run in separate containers. Share data via the mounted workspace or external services.

**Q: Can I use languages other than Markdown for instructions?**  
A: OpenCode expects Markdown. Use code blocks for examples in other languages.

**Q: Do I need to restart containers when updating skills?**  
A: Yes. Run `caiged session restart . --spin <name>` to rebuild and restart.
