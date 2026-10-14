#!/bin/sh
# Installs the prea CLI and puts it on the PATH:
#
#   curl -fsSL https://raw.githubusercontent.com/prea-framework/prea/main/install.sh | sh
#
# PREA_VERSION picks a version (default latest): ... | PREA_VERSION=v0.0.1 sh
# The framework itself comes with go get in each project; this installs the CLI.
set -eu

version="${PREA_VERSION:-latest}"

fail() {
	printf 'prea: %s\n' "$1" >&2
	exit 1
}

command -v go >/dev/null 2>&1 || fail "Go is required (1.21 or newer, it fetches the toolchain prea needs): https://go.dev/dl"

printf 'Installing prea %s...\n' "$version"
go install "github.com/prea-framework/prea/cmd/prea@$version" || fail "go install failed"

bin="$(go env GOBIN)"
[ -n "$bin" ] || bin="$(go env GOPATH)/bin"

case ":$PATH:" in
*":$bin:"*)
	printf 'prea installed: %s/prea\n' "$bin"
	exit 0
	;;
esac

# Not on the PATH: add it to the shell's startup file, once.
case "$(basename "${SHELL:-sh}")" in
zsh)
	rc="${ZDOTDIR:-$HOME}/.zshrc"
	line="export PATH=\"\$PATH:$bin\""
	;;
bash)
	rc="$HOME/.bashrc"
	[ "$(uname)" = Darwin ] && rc="$HOME/.bash_profile" # macOS terminals open login shells
	line="export PATH=\"\$PATH:$bin\""
	;;
fish)
	rc="$HOME/.config/fish/config.fish"
	line="fish_add_path \"$bin\""
	;;
*)
	rc="$HOME/.profile"
	line="export PATH=\"\$PATH:$bin\""
	;;
esac

if ! grep -qsF "$line" "$rc"; then
	mkdir -p "$(dirname "$rc")"
	printf '\n# prea\n%s\n' "$line" >>"$rc"
	printf 'Added %s to the PATH in %s\n' "$bin" "$rc"
fi
printf 'prea installed: %s/prea\nOpen a new terminal, or run: %s\n' "$bin" "$line"
