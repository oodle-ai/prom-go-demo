#!/bin/sh

# Replace environment variables in prometheus.yml
sed -e "s|\${REMOTE_WRITE_URL}|$REMOTE_WRITE_URL|g" \
    -e "s|\${X_API_KEY}|$X_API_KEY|g" \
    -e "s|\${APP_PORT}|${APP_PORT:-6767}|g" \
    -e "s|\${PROMETHEUS_PORT}|${PROMETHEUS_PORT:-9797}|g" \
    /etc/prometheus/prometheus.yml > /tmp/prometheus.yml

# Start Prometheus with the processed config file
/bin/prometheus \
    --config.file=/tmp/prometheus.yml \
    --storage.tsdb.path=/prometheus \
    --web.console.libraries=/usr/share/prometheus/console_libraries \
    --web.console.templates=/usr/share/prometheus/consoles \
    --web.enable-lifecycle \
    --web.enable-admin-api \
    --web.listen-address=:${PROMETHEUS_PORT:-9797}
