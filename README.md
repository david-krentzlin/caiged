# Caiged for OpenCode

Caiged is a tool to gloss over the boring plumbing to run OpenCode in server mode in a docker container.
The goal is to restrict what your agent can do while maintaining relatively good UX and resource usage.

It allows you to create specially tailored OpenCode instances, that have their own AGENTS.md, and skills pre-configured,
which I call spins. This way you can have development, qa, security review etc. done with the tools you need in an easy way.

## Caution

This is in active development and things may change rapidly. 

## Quickstart

### Install

```bash
git clone https://github.com/david-krentzlin/caiged.git ~/.caiged
cd ~/.caiged
make install
```


### Run

Inside the working directory of your project invoke the following: 

```bash
caiged run . --spin dev             # Start/connect in current directory
```

This creates (or connects to) a container with OpenCode server running, preconfigured with the dev spin.
You will be automatically connected to OpenCode TUI (on your local machine) which connects to the OpenCode server in the container.


### Read the docs

```bash
man caiged
```

---

## Why Use Caiged?

If you have to use coding agents, then at least try to do it sensibly and don't give them access to your entire machine.
The natural approach to that is to put your agent into a dockerized environment. This can be tedious to do and maintain and you will likely end up
with a bunch of shell scripts to help yourself. Caiged has just taken this problem and turned it into a little go program and a process that makes this
easier and gets you productive.

## How is that different to docker sandbox?

[Docker sandboxes](https://docs.docker.com/ai/sandboxes/) solve fundamentally the same problem, but make different trade-offs.

Docker sandboxes give you higher degrees of isolation, but at the cost of requiring more resources and making some use-cases very cumbersome.
If you want to run you agent in yolo mode without much supervision, docker sandboxes are the better option for you. Also, docker sandboxes provide
some other niceties like making sure you git configuration (username etc. are present).

**The most important feature of docker sandboxes remains however the level of isolation. It does not give the container access to the docker socket,
which means that it's not possible to access files that the docker daemon can access.**

Caiged, makes a different trade-off. It does not use a full vm but just a docker container, which is lighter on resource usage.
**By default, docker socket access is disabled for security.** You can optionally enable it with the `--enable-docker-sock` flag if your workflow requires docker-in-docker capabilities.
When enabled, the agent will see other docker containers. If that is not acceptable for you, please use docker sandboxes or something else.
In my workflows however I found that filesystem isolation is what I really care about.
(More than once have I seen the agent wander off into the distance on my filesystem way outside the current working directory.)

---

## How is that different to opencode-devcontainers?

[opencode-devcontainers](https://github.com/athal7/opencode-devcontainers) is an OpenCode plugin that provides branch-based workspace isolation using devcontainers or git worktrees. Both tools aim to isolate your agent environments, but they approach the problem differently:

**opencode-devcontainers** (plugin-based approach):
- Works as an OpenCode plugin that runs inside a single OpenCode session
- Creates isolated clones/worktrees per branch with auto-assigned ports (13000-13099)
- Routes commands to different workspaces using `/devcontainer` or `/worktree` commands
- Focused on multi-branch workflows - great for working on multiple issues simultaneously
- Requires the agent to manage workspace switching within a session
- Lighter weight, less overhead per workspace

**caiged** (container-per-project approach):
- Manages entire containerized OpenCode servers (one container = one project/spin)
- Each container has its own independent OpenCode server with separate configuration
- Focused on role-based isolation (dev, qa, security) with tailored AGENTS.md and skills per spin
- Persistent sessions that survive stop/resume cycles with all installed packages intact
- Designed for supervised agent workflows where you connect to specific environments
- Each container is a complete, isolated development environment

**When to use what:**

Use **opencode-devcontainers** if you:
- Want to work on multiple branches of the same project simultaneously
- Need lightweight filesystem isolation for quick branch switching
- Prefer managing everything within a single OpenCode session
- Want automatic branch workspace management

Use **caiged** if you:
- Want different agent personas/roles (dev, qa, security) with specialized tooling
- Need complete environment isolation per project (different dependencies, databases)
- Want persistent containers that preserve installed tools across sessions
- Prefer explicit, supervised control over separate containerized environments
- Want to connect from any directory to running containers by name

**Can they work together?**
Yes! You could run caiged to create a containerized dev environment with a specific spin, and then use opencode-devcontainers inside that container for branch-based isolation. They solve different layers of the isolation problem.

---

## Prerequisites

You will need to have the following available on your host system:

* docker
* OpenCode
* go


## How does it work


Let's have a more detailed look at what happens when you create your first container

```bash
cd /path/to/your/project
caiged run . --spin dev
```



1. **CLI** (`caiged`) builds the base image + spin image (if needed)
2. **Container** starts with your project directory bind-mounted
3. **OpenCode server** runs inside container with password authentication
4. **OpenCode TUI** connects to the server via HTTP from host

In a picture that looks something like this:

```mermaid
flowchart TB
  subgraph Host[Host Machine]
    CLI[caiged CLI]
    TUI["OpenCode TUI<br/>(connects to server)"]
    TERM[your terminal]
    CLI -->|launches| TUI
  end

  subgraph Container[Container]
    SERVER["OpenCode Server<br/>(in tmux session)<br/>- password protected<br/>- port mapped to host"]
    IMG["spin image (qa/dev/...)<br/>- tools via mise<br/>- opencode config<br/>- agent instructions"]
    SERVER -->|loads config from| IMG
  end

  CLI -->|docker run| IMG
  TUI -->|http://localhost:PORT| SERVER
```


**Container naming**: `caiged-<spin>-<project>`
- Default project name: last two path segments of your working directory
- Override: `--project <name>`

**Password generation**: Deterministic SHA256 hash from container name + salt
- Salt stored in `~/.config/caiged/salt` (created once)
- CLI regenerates same password for connecting to existing containers


---

## Spins

Spins live under `spins/` and define a role-specific environment:
- `AGENTS.md`: detailed agent instructions and persona
- `skills/`: domain-specific skills (test generation, security review, etc.)
- `mcp/`: MCP server configs
- `README.md`: spin-specific documentation

**Creating new spins:**
The build process automatically handles new spins - just create a directory under `spins/<name>/` with the required files and run:

```bash
caiged run . --spin <name>
```

See [SPINS.md](SPINS.md) for detailed instructions on creating and contributing spins

---

### Common Workflows

**Default - just run it:**
```bash
caiged run . --spin dev   # Current directory, dev spin
```

**Explicit workdir:**
```bash
caiged run /path/to/project --spin qa
```

**Connect to an existing container by project name:**
```bash
caiged connect <container-name>
```

**List all containers with easy-to-copy commands:**
```bash
caiged containers list
```

**Stop everything:**
```bash
caiged containers stop-all
```

**Open a shell in a container (for debugging):**
```bash
caiged containers shell <container-name>
```

**Force rebuild with latest tools:**
```bash
OPENCODE_VERSION=latest caiged run . --spin dev --rebuild-images
```

By default, caiged tries to match your host OpenCode client version for container builds by using `opencode --version` when `OPENCODE_VERSION` is not explicitly set.

**Pass selected host secrets into the container:**
```bash
export JFROG_OIDC_USER=...
export JFROG_OIDC_TOKEN=...

caiged run . --spin dev --secret-env JFROG_OIDC_USER --secret-env JFROG_OIDC_TOKEN
```

## Troubleshooting

### Control keys not working in `caiged connect`

If `Ctrl+C` (or other control keys) stops working only when attached to a container server, the most common cause is a host/client and container/server OpenCode version mismatch.

Check versions:

```bash
opencode --version
docker exec <container-name> zsh -lc 'start-opencode --version'
```

If they differ, align versions and rebuild/recreate the container:

```bash
# Option A: upgrade host client to match container
opencode upgrade <version>

# Option B: pin container version to match host
OPENCODE_VERSION=<version> caiged run . --spin <spin> --rebuild-images
```

Notes:
- Containers are persistent, so old versions remain until you recreate the container.
- If your terminal state gets corrupted after aborting a TUI session, run `stty sane` (and `reset` if needed).

## Container / Spin Security

- **Network**: uses bridge networking with port mapping
  - Bridge networking with port mapping allows secure OpenCode server access from host
  - Each container gets a unique port (starting at 4096) mapped to container port 4096
- **Docker socket**: disabled by default for security; enable with `--enable-docker-sock` if docker-in-docker is required
- **GitHub config**: mounted read-only from `~/.config/gh`; make read-write with `--mount-gh-rw`
- **OpenCode auth reuse**: host `~/.local/share/opencode/auth.json` is mounted read-only when available; disable with `--no-mount-opencode-auth`
- **Secret env passthrough**: only explicitly listed host env vars are passed to the container (`--secret-env NAME`, repeatable)

### Credentials

The container includes `gh` (GitHub CLI). Host `~/.config/gh` is mounted read-only by default.

OpenCode authentication behavior:
- By default, caiged mounts host `~/.local/share/opencode/auth.json` read-only when available
- If that file does not exist on the host, OpenCode starts unauthenticated in the container and you need to complete auth once there (`/connect`)
- Disable host auth reuse with `--no-mount-opencode-auth`

### Secret environment variables:
- Canonical approach: pass only explicit host env vars with `--secret-env`
- Pass selected host secrets with `--secret-env NAME` (repeatable), for example `JFROG_OIDC_USER` and `JFROG_OIDC_TOKEN`
- Or provide a Docker-compatible env file with `--secret-env-file /path/to/secrets.env`

**Example:**

```bash
caiged . --spin dev --secret-env JFROG_OIDC_USER --secret-env JFROG_OIDC_TOKEN
```

## Is this vibe-coded?

It is built using agentic coding but with oversight and I did multiple rounds of refactoring along the way.
But at its core it's built mostly by an agent, to be fast and try ideas.
Iff this becomes something that turns out to be viable, I will probably rewrite bigger parts manually and make sure
I know every bit of it.

## Missing features

The following is what is still missing and an idea of what might come

* Provide GIT configuration
  - provide an easy way to commit under the user's git user name (optionally)
  - however I would like to use a different ssh-key just for the agent which is trusted on my github (this gives tracking of AI activity for free)
- Allow spins to provision the docker container with different tools
