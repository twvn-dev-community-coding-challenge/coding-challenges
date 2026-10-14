# Bakes config into the image so Docker never bind-mounts prometheus.yml (avoids file/dir mount bugs).
FROM prom/prometheus:v2.53.1
COPY config/prometheus.yml /etc/prometheus/prometheus.yml
