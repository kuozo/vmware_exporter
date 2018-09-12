FROM gcr.io/distroless/base

COPY dist/vmware_exporter /usr/local/bin/
CMD ['vmware_exporter']
EXPOSE 9272