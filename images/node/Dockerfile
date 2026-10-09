# A kind node that looks like a CKA exam host: ssh in as candidate, sudo -i for root,
# kubectl with the k alias and completion, yq, curl, wget, man, vim, and etcdctl/etcdutl.
ARG K8S_VERSION=v1.35.8
FROM kindest/node:${K8S_VERSION}

ARG ETCD_VERSION=v3.6.6
ARG YQ_VERSION=v4.54.1

# The slim base image strips man pages; the exam hosts have them.
RUN rm -f /etc/dpkg/dpkg.cfg.d/docker \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      openssh-server sudo vim less man-db manpages bash-completion curl wget jq iputils-ping dnsutils netcat-openbsd \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends --reinstall coreutils grep sed tar systemd util-linux mount procps iproute2 \
 && mandb -q \
 && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -o /usr/local/bin/yq \
 && chmod +x /usr/local/bin/yq \
 && curl -fsSL "https://github.com/etcd-io/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz" \
      | tar -xz -C /usr/local/bin --strip-components=1 "etcd-${ETCD_VERSION}-linux-amd64/etcdctl" "etcd-${ETCD_VERSION}-linux-amd64/etcdutl"

RUN useradd --create-home --shell /bin/bash candidate \
 && echo 'candidate ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/candidate \
 && chmod 0440 /etc/sudoers.d/candidate \
 && mkdir -p /run/sshd \
 && sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config \
 && systemctl enable ssh

COPY exam-profile.sh /etc/profile.d/cka-exam.sh
RUN echo '. /etc/profile.d/cka-exam.sh' >> /etc/bash.bashrc
