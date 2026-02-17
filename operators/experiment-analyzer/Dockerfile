FROM node:22-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      curl jq python3-pip chromium fonts-liberation && \
    pip3 install --break-system-packages awscli && \
    npm install -g @anthropic-ai/claude-code @mermaid-js/mermaid-cli && \
    rm -rf /var/lib/apt/lists/* /root/.cache

ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium

COPY analyze.sh /usr/local/bin/analyze.sh
RUN chmod +x /usr/local/bin/analyze.sh

ENTRYPOINT ["/usr/local/bin/analyze.sh"]
