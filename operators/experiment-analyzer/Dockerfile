FROM node:22-slim

RUN npm install -g @anthropic-ai/claude-code && \
    apt-get update && apt-get install -y curl jq && \
    rm -rf /var/lib/apt/lists/*

COPY analyze.sh /usr/local/bin/analyze.sh
RUN chmod +x /usr/local/bin/analyze.sh

ENTRYPOINT ["/usr/local/bin/analyze.sh"]
