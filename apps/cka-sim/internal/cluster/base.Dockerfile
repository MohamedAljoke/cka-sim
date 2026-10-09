FROM ubuntu:24.04

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      openssh-client vim less bash-completion curl ca-certificates iputils-ping \
 && rm -rf /var/lib/apt/lists/* \
 && userdel --remove ubuntu \
 && useradd --create-home --shell /bin/bash candidate

# The nodes are recreated by every up, so their host keys change; never stop on that.
COPY <<'EOF' /etc/ssh/ssh_config.d/cka.conf
Host *
  User candidate
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  LogLevel ERROR
EOF

CMD ["sleep", "infinity"]
