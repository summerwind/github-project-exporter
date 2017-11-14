FROM quay.io/prometheus/busybox:latest

COPY ./github-project-exporter /bin/github-project-exporter

EXPOSE 9410
ENTRYPOINT ["/bin/github-project-exporter"]
