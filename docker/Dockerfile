FROM alpine:3.20 AS base

ARG MISE_VERSION=2026.2.13
ARG GH_VERSION=2.86.0
ARG OPENCODE_VERSION=latest
ARG ARCH=arm64

ENV AGENT_WORKDIR=/workspace
ENV MISE_DATA_DIR=/opt/mise
ENV PATH="/opt/mise/shims:/root/.bun/bin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ENV SHELL=/bin/zsh
ENV OPENCODE_CONFIG_DIR=/root/.config/opencode
ENV OPENCODE_VERSION=${OPENCODE_VERSION}
ENV TERM=xterm-256color
ENV COLORTERM=truecolor

# Install runtime dependencies and build dependencies in separate layers
RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    git \
    jq \
    libgcc \
    libstdc++ \
    ncurses-terminfo \
    openssh-client \
    ripgrep \
    tmux \
    unzip \
    xz \
    zsh \
  && apk add --no-cache --virtual .docker-deps docker-cli \
  && apk add --no-cache --virtual .build-deps \
    make \
    autoconf

RUN case "$ARCH" in \
    amd64) \
      GH_PKG_ARCH="amd64"; \
      MISE_PKG_ARCH="linux-x64-musl"; \
      ;; \
    arm64) \
      GH_PKG_ARCH="arm64"; \
      MISE_PKG_ARCH="linux-arm64-musl"; \
      ;; \
    *) echo "Unsupported arch: $ARCH (supported: amd64, arm64)" >&2; exit 1 ;; \
  esac \
  && curl -sSLo /tmp/gh.tar.gz \
    "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${GH_PKG_ARCH}.tar.gz" \
  && tar -xzf /tmp/gh.tar.gz -C /tmp \
  && mv /tmp/gh_${GH_VERSION}_linux_${GH_PKG_ARCH}/bin/gh /usr/local/bin/ \
  && rm -rf /tmp/gh* \
  && curl -sSLo /usr/local/bin/mise \
    "https://github.com/jdx/mise/releases/download/v${MISE_VERSION}/mise-v${MISE_VERSION}-${MISE_PKG_ARCH}" \
  && chmod +x /usr/local/bin/mise

COPY config/target_mise.toml /etc/mise.toml
COPY config/tmux.conf /etc/tmux.conf
COPY config/zshrc /etc/zsh/zshrc
COPY config/zprofile /etc/zsh/zprofile

RUN mkdir -p /root/.config/mise \
  && cp /etc/mise.toml /root/.config/mise/config.toml \
  && mkdir -p /root/.local/share/opencode \
  && mkdir -p "${MISE_DATA_DIR}" \
  && MISE_YES=1 mise install \
  && mise reshim \
  && if [ "$OPENCODE_VERSION" = "latest" ]; then bun add -g opencode-ai; else bun add -g "opencode-ai@${OPENCODE_VERSION}"; fi \
  && apk del .build-deps

WORKDIR /workspace

COPY entrypoint.sh /usr/local/bin/agent-entrypoint
COPY scripts/start-opencode.sh /usr/local/bin/start-opencode
COPY scripts/comma-help.sh /usr/local/bin/,help
RUN chmod +x /usr/local/bin/agent-entrypoint \
  /usr/local/bin/start-opencode \
  /usr/local/bin/,help

ENTRYPOINT ["/usr/local/bin/agent-entrypoint"]

FROM base AS spin

ARG SPIN=qa
ENV AGENT_SPIN=${SPIN}
ENV AGENT_SPIN_DIR=/opt/agent/spin

COPY spins/${SPIN}/ /opt/agent/spin/
RUN mkdir -p "$OPENCODE_CONFIG_DIR/agents" \
  && SPIN_NAME="${SPIN}" \
  && if [ -f /opt/agent/spin/AGENTS.md ]; then cp /opt/agent/spin/AGENTS.md "$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md"; fi \
  && if [ -f /opt/agent/spin/AGENT.md ] && [ ! -f "$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md" ]; then cp /opt/agent/spin/AGENT.md "$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md"; fi \
  && if [ -d /opt/agent/spin/skills ]; then cp -R /opt/agent/spin/skills "$OPENCODE_CONFIG_DIR/"; fi \
  && if [ -d /opt/agent/spin/mcp ]; then cp -R /opt/agent/spin/mcp "$OPENCODE_CONFIG_DIR/"; fi \
  && cat > "$OPENCODE_CONFIG_DIR/opencode.json" <<EOF
{
  "agent": {
    "${SPIN_NAME}": {
      "description": "Spin-specific agent: ${SPIN_NAME}",
      "mode": "primary",
      "prompt": "{file:$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md}"
    }
  },
  "default_agent": "${SPIN_NAME}"
}
EOF

