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
echo "Run the following commands to finish setting up savvy:"

case :$PATH:
  in *:$HOME/bin/foo:*) ;; # do nothing
     *) echo "> echo 'export PATH=\"\$HOME/bin:\$PATH\"' >> ~/.zshrc";;
esac

echo "> echo 'eval \"\$(savvy init zsh)\"' >> ~/.zshrc"
echo "> source ~/.zshrc # to pick up the new changes"
echo
echo "Run 'savvy help' to learn more or checkout our docs at https://github.com/getsavvyinc/savvy-cli"
echo
echo "Stuck? Join our Discord https://getsavvy.so/discord"

