#!/bin/sh
#
# Adapted from the pressly/goose installer: Copyright 2021. MIT License.
# Ref: https://github.com/pressly/goose/blob/master/install.sh
#
# Adapted from the Deno installer: Copyright 2019 the Deno authors. All rights reserved. MIT license.
# Ref: https://github.com/denoland/deno_install
#
# TODO(everyone): Keep this script simple and easily auditable.

# Not intended for Windows.

set -e

# source: colors.sh file on render managed envs
# source: https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux
BLUE=$(tput setaf 4)
CYAN=$(tput setaf 6)
BOLD=$(tput bold)
RED=$(tput setaf 1)
RESET=$(tput sgr0)

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

if [ "$arch" = "aarch64" ]; then
	arch="arm64"
fi

if [ $# -eq 0 ]; then
	savvy_uri="https://github.com/getsavvyinc/savvy-cli/releases/latest/download/savvy_${os}_${arch}"
else
	savvy_uri="https://github.com/getsavvyinc/savvy-cli/releases/download/${1}/savvy_${os}_${arch}"
fi

savvy_install="${SAVVY_INSTALL:-$HOME}"
bin_dir="${savvy_install}/bin"
exe="${bin_dir}/savvy"

if [ ! -d "${bin_dir}" ]; then
	mkdir -p "${bin_dir}"
fi

curl --silent --show-error --fail --location --output "${exe}" "$savvy_uri"
chmod +x "${exe}"

echo
echo "savvy was installed successfully to ${exe}"
echo



echo
echo "${BLUE}${BOLD}Run the following commands to finish setting up savvy:${RESET}"
echo

# defaults to zsh
shell="zsh"
if [ -n "$ZSH_VERSION" ]; then
  shell="zsh"
elif [ -n "$BASH_VERSION" ]; then
  shell="bash"
fi

case :$PATH:
  in *:${bin_dir}*) ;; # do nothing
     *) echo "${BLUE}> echo 'export PATH=\"\$HOME/bin:\$PATH\"' >> ~/.${shell}rc${RESET}";;
esac

case :$shell:
  in  *zsh*) echo "${BLUE}> echo 'eval \"\$(savvy init zsh)\"' >> ~/.zshrc${RESET}";;
      *bash*) echo "${BLUE}> echo 'eval \"\$(savvy init bash)\"' >> ~/.bashrc${RESET}";;
esac

echo "${BLUE}> source ~/.${shell}rc # to pick up the new changes${RESET}"
echo
echo "Run 'savvy help' to learn more or checkout our docs at https://github.com/getsavvyinc/savvy-cli"
echo
echo "${RED}${BOLD}Stuck?${RESET} ${RED}We'd love to help. Just join our Discord https://getsavvy.so/discord${RESET}"

