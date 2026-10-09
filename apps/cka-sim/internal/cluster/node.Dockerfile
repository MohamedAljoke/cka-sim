FROM kindest/node:v1.37.0

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends vim less \
 && rm -rf /var/lib/apt/lists/*
