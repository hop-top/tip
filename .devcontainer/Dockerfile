# tip benchmark/test container
# Includes all 5 adapter target tools (git, gh, docker, wrangler, mmdc)
# plus tip + ben for real-workflow benchmarking.
# Uses 1.24 base + GOTOOLCHAIN=auto to fetch go1.26.
FROM golang:1.24-trixie

ARG USERNAME=dev
ARG USER_UID=1000
ARG USER_GID=${USER_UID}

# --- Base packages ---
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    jq \
    zsh \
    bash \
    sudo \
    sqlite3 \
    gnupg \
    lsb-release \
    && rm -rf /var/lib/apt/lists/*

# --- gh CLI ---
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) \
    signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] \
    https://cli.github.com/packages stable main" \
    > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# --- Docker CLI (client only — talks to host daemon via socket) ---
RUN curl -fsSL https://download.docker.com/linux/debian/gpg \
    | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) \
    signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
    https://download.docker.com/linux/debian trixie stable" \
    > /etc/apt/sources.list.d/docker.list \
    && apt-get update \
    && apt-get install -y docker-ce-cli \
    && rm -rf /var/lib/apt/lists/*

# --- Node.js (for wrangler + mmdc) ---
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# --- wrangler + mmdc (mermaid-cli) ---
RUN npm install -g wrangler @mermaid-js/mermaid-cli

# --- Non-root user ---
RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} \
       -m -s /usr/bin/zsh ${USERNAME} \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" \
       >> /etc/sudoers.d/${USERNAME}

USER ${USERNAME}
ENV GOPATH=/home/${USERNAME}/go
ENV GOTOOLCHAIN=auto
ENV PATH=/home/dev/.local/bin:${GOPATH}/bin:${PATH}

WORKDIR /workspace

# --- Go deps (cache layer; GOTOOLCHAIN=auto fetches go1.26) ---
COPY --chown=${USERNAME} go.mod go.sum ./
RUN go mod download

# --- Copy source + build tip ---
COPY --chown=${USERNAME} . .
RUN make build \
    && mkdir -p ${GOPATH}/bin \
    && cp bin/tip ${GOPATH}/bin/tip

# --- Install ben ---
RUN GONOSUMCHECK="hop.top/*" GOPRIVATE="hop.top/*" \
    go install hop.top/ben/cmd/ben@latest 2>/dev/null || true

# --- Install rtk (CLI proxy for token compression) ---
RUN mkdir -p /home/dev/.local/bin \
    && curl -fsSL \
    https://raw.githubusercontent.com/rtk-ai/rtk/refs/heads/master/install.sh \
    | sh

# --- Verify tools ---
RUN echo "=== Tool versions ===" \
    && git --version \
    && gh --version | head -1 \
    && docker --version \
    && wrangler --version \
    && mmdc --version \
    && tip version \
    && rtk --version \
    && echo "=== All tools present ==="

CMD ["zsh"]
