#!/bin/sh
set -eu

repo=${CKA_SIM_REPO:-MohamedAljoke/cka-sim}
ref=${CKA_SIM_REF:-main}
dir=${CKA_SIM_HOME:-$HOME/.cka-sim}
raw=https://raw.githubusercontent.com/$repo/$ref/infra

command -v curl >/dev/null 2>&1 || { echo "install.sh: needs curl" >&2; exit 1; }

mkdir -p "$dir"
for f in compose.yml cka-sim; do
  curl -fsSL "$raw/$f" -o "$dir/$f.new"
  mv "$dir/$f.new" "$dir/$f"
done
chmod +x "$dir/cka-sim"

bin=$HOME/.local/bin
mkdir -p "$bin"
ln -sf "$dir/cka-sim" "$bin/cka-sim"
case ":$PATH:" in
*":$bin:"*) ;;
*) echo "add $bin to your PATH to run cka-sim from anywhere, or run $dir/cka-sim" ;;
esac

[ -n "${CKA_SIM_INSTALL_ONLY:-}" ] || exec "$dir/cka-sim" up
