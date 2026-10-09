# The exam's "base" host: where you land, with ssh and an editor but no Kubernetes tools.
FROM ubuntu:24.04

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      openssh-client vim less bash-completion curl ca-certificates iputils-ping \
 && rm -rf /var/lib/apt/lists/* \
 && userdel --remove ubuntu 2>/dev/null; useradd --create-home --shell /bin/bash candidate

COPY ssh_config /etc/ssh/ssh_config.d/cka.conf
RUN echo "PS1='\\u@\\h:\\w\\$ '" >> /home/candidate/.bashrc

CMD ["sleep", "infinity"]
