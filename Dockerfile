FROM docker:17.12.0-ce-dind
ADD drone-docker-image-promote /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker-image-promote"]
