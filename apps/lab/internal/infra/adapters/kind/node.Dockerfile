FROM kindest/node:v1.37.0

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      vim less openssh-server sudo bash-completion \
 && rm -rf /var/lib/apt/lists/*

# sshd is enabled, not started from CMD: kind's entrypoint has to stay PID 1.
RUN useradd --create-home --shell /bin/bash candidate \
 && echo 'candidate ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/candidate \
 && chmod 0440 /etc/sudoers.d/candidate \
 && sed -i 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config \
 && systemctl enable ssh

COPY <<'EOF' /etc/profile.d/cka-exam.sh
if [ -n "$BASH_VERSION" ] && [ -z "$CKA_EXAM_PROFILE" ]; then
  CKA_EXAM_PROFILE=1
  [ -f /usr/share/bash-completion/bash_completion ] && . /usr/share/bash-completion/bash_completion
  source <(kubectl completion bash)
  alias k=kubectl
  complete -o default -F __start_kubectl k
fi
EOF
RUN echo '. /etc/profile.d/cka-exam.sh' >> /etc/bash.bashrc
