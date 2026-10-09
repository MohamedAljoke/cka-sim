# What every CKA exam host gives you: the k alias with completion, and a prompt naming the host.
if [ -n "$BASH_VERSION" ] && [ -z "$CKA_EXAM_PROFILE" ]; then
  CKA_EXAM_PROFILE=1
  [ -f /usr/share/bash-completion/bash_completion ] && . /usr/share/bash-completion/bash_completion
  if command -v kubectl >/dev/null; then
    source <(kubectl completion bash)
    alias k=kubectl
    complete -o default -F __start_kubectl k
  fi
  PS1='\u@\h:\w\$ '
fi
