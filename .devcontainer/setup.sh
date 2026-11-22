#!/usr/bin/env bash
set -euo pipefail

# --- Shell setup first, so a failed tool install below never skips it ---

sudo chown "$(id -u):$(id -g)" /commandhistory
sudo chown "$(id -u):$(id -g)" /home/vscode/.gemini
if [ -d /mnt/mise-data ]; then
  sudo chown -R "$(id -u):$(id -g)" /mnt/mise-data
fi
touch /commandhistory/.bash_history

sed -i '/# >>> domus shell >>>/,/# <<< domus shell <<</d' ~/.bashrc
cat >> ~/.bashrc <<'BASHRC'
# >>> domus shell >>>
if [ -w /commandhistory ]; then
  export HISTFILE=/commandhistory/.bash_history
  export PROMPT_COMMAND='history -a'
fi
if [ -d "$HOME/.bun/bin" ]; then
  export PATH="$HOME/.bun/bin:$PATH"
fi
if command -v mise >/dev/null 2>&1; then
  eval "$(mise activate bash)"
fi
# <<< domus shell <<<
BASHRC

# --- Tools via mise & bun ---

mise trust --yes || true
mise install

if ! command -v bun >/dev/null 2>&1 && [ ! -x "$HOME/.bun/bin/bun" ]; then
  curl -fsSL https://bun.sh/install | bash
fi
export PATH="$HOME/.bun/bin:$PATH"

bun install
bun install --cwd ui
bun install --cwd e2e
