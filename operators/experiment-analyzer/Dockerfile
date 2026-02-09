FROM node:22-slim

RUN npm install -g @anthropic-ai/claude-code && \
    apt-get update && apt-get install -y curl jq python3-pip && \
    pip3 install --break-system-packages awscli && \
    rm -rf /var/lib/apt/lists/* /root/.cache

COPY analyze.sh /usr/local/bin/analyze.sh
RUN chmod +x /usr/local/bin/analyze.sh

ENTRYPOINT ["/usr/local/bin/analyze.sh"]
