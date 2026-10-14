# Bakes config + provisioning + dashboards into the image (avoids bind-mount issues; rebuild after edits).
FROM grafana/grafana:10.4.6
COPY config/grafana.ini /etc/grafana/grafana.ini
COPY provisioning /etc/grafana/provisioning
COPY dashboards /etc/grafana/dashboards
