FROM debian:bookworm-slim AS base

ARG DEBIAN_FRONTEND=noninteractive
ARG MISE_VERSION=2026.2.13
ARG GH_VERSION=2.86.0
ARG OP_VERSION=2.32.1
ARG OPENCODE_VERSION=latest
ARG ARCH=arm64

ENV AGENT_WORKDIR=/workspace
ENV MISE_DATA_DIR=/opt/mise
ENV PATH="/opt/mise/shims:/root/.bun/bin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV SHELL=/bin/zsh
ENV STARSHIP_CONFIG=/etc/starship.toml
ENV OPENCODE_CONFIG_DIR=/root/.config/opencode
ENV OPENCODE_VERSION=${OPENCODE_VERSION}
ENV TERM=xterm-256color
ENV COLORTERM=truecolor

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    git \
    gnupg \
    jq \
    ncurses-term \
    openssh-client \
    tmux \
    docker.io \
    unzip \
    xz-utils \
    zsh \
  && rm -rf /var/lib/apt/lists/*

RUN case "$ARCH" in \
    amd64) \
      GH_PKG_ARCH="amd64"; \
      OP_PKG_ARCH="amd64"; \
      MISE_PKG_ARCH="linux-x64"; \
      ;; \
    arm64) \
      GH_PKG_ARCH="arm64"; \
      OP_PKG_ARCH="arm64"; \
      MISE_PKG_ARCH="linux-arm64"; \
      ;; \
    *) echo "Unsupported arch: $ARCH (supported: amd64, arm64)" >&2; exit 1 ;; \
  esac \
  && curl -sSLo /tmp/gh.deb \
    "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${GH_PKG_ARCH}.deb" \
  && curl -sSLo /tmp/1password-cli.deb \
    "https://downloads.1password.com/linux/debian/${OP_PKG_ARCH}/stable/1password-cli-${OP_PKG_ARCH}-latest.deb" \
  && apt-get update \
  && apt-get install -y --no-install-recommends /tmp/gh.deb /tmp/1password-cli.deb \
  && OP_INSTALLED="$(op --version)" \
  && case "$OP_INSTALLED" in \
    ${OP_VERSION}*) ;; \
    *) echo "Expected op ${OP_VERSION}, got ${OP_INSTALLED}" >&2; exit 1 ;; \
  esac \
  && rm -f /tmp/gh.deb /tmp/1password-cli.deb \
  && rm -rf /var/lib/apt/lists/* \
  && curl -sSLo /usr/local/bin/mise \
    "https://github.com/jdx/mise/releases/download/v${MISE_VERSION}/mise-v${MISE_VERSION}-${MISE_PKG_ARCH}" \
  && chmod +x /usr/local/bin/mise

COPY config/target_mise.toml /etc/mise.toml
COPY config/starship.toml /etc/starship.toml
COPY config/tmux.conf /etc/tmux.conf
COPY config/zshrc /etc/zsh/zshrc
COPY config/zprofile /etc/zsh/zprofile

RUN mkdir -p /root/.config/mise \
  && cp /etc/mise.toml /root/.config/mise/config.toml \
  && mkdir -p /root/.local/share/opencode \
  && mkdir -p "${MISE_DATA_DIR}" \
  && MISE_YES=1 mise install \
  && mise reshim \
  && if [ "$OPENCODE_VERSION" = "latest" ]; then bun add -g opencode-ai; else bun add -g "opencode-ai@${OPENCODE_VERSION}"; fi

WORKDIR /workspace

COPY entrypoint.sh /usr/local/bin/agent-entrypoint
COPY scripts/start-opencode.sh /usr/local/bin/start-opencode
COPY scripts/comma-help.sh /usr/local/bin/,help
COPY scripts/comma-auth-tools.sh /usr/local/bin/,auth-tools
RUN chmod +x /usr/local/bin/agent-entrypoint \
  /usr/local/bin/start-opencode \
  /usr/local/bin/,help \
  /usr/local/bin/,auth-tools

ENTRYPOINT ["/usr/local/bin/agent-entrypoint"]

FROM base AS spin

ARG SPIN=qa
ENV AGENT_SPIN=${SPIN}
ENV AGENT_SPIN_DIR=/opt/agent/spin

COPY spins/${SPIN}/ /opt/agent/spin/
RUN mkdir -p "$OPENCODE_CONFIG_DIR" \
  && if [ -f /opt/agent/spin/AGENTS.md ]; then cp /opt/agent/spin/AGENTS.md "$OPENCODE_CONFIG_DIR/AGENTS.md"; fi \
  && if [ -f /opt/agent/spin/AGENT.md ] && [ ! -f "$OPENCODE_CONFIG_DIR/AGENTS.md" ]; then cp /opt/agent/spin/AGENT.md "$OPENCODE_CONFIG_DIR/AGENTS.md"; fi \
  && if [ -d /opt/agent/spin/skills ]; then cp -R /opt/agent/spin/skills "$OPENCODE_CONFIG_DIR/"; fi \
  && if [ -d /opt/agent/spin/mcp ]; then cp -R /opt/agent/spin/mcp "$OPENCODE_CONFIG_DIR/"; fi
