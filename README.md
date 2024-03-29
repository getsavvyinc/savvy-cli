# Savvy

<h3 align="left">
  | <a href="https://getsavvy.so/discord">Discord</a> |
  <a href="https://twitter.com/savvyoncall">Twitter</a> |
  <a href="https://www.getsavvy.so/">Website</a> |
</h3>

Savvy helps your create and share runbooks directly from your terminal.

## Demo

![Savvy Runbook](https://vhs.charm.sh/vhs-1UmW0o6uSztF6b76y92K2K.gif)

## Quick Start

Follow these steps to get started:

1. **Install Savvy CLI**

Run the following command in your terminal

```sh
curl -fsSL https://install.getsavvy.so | sh
```

Follow the on-screen instructions to complete the installation.

2. **Login**

Before you can create runbooks using the CLI you need to login. Use the following command:

```sh
savvy login
```

3. **Create a Runbook**

Run the following command to start creating a runbook from your terminal:

```sh
savvy record
```

Perform the tasks you wish to record in your terminal. When you're done, you can stop the recording by typing `exit` or pressing `ctrl-D`.

4. **Upgrade CLI**

```sh
savvy upgrade
```

## Limitations

Currently, Savvy does not support the following:

* Windows
* Savvy supports `zsh` and `bash`. Please [create an issue](https://github.com/getsavvyinc/savvy-cli/issues/new) if you'd like us to support another shell.

## Getting Help

If you need assistance or have questions:

* [Create an issue](https://github.com/getsavvyinc/savvy-cli/issues/new) on our GitHub repository.
* Join our [Discord](https://getsavvy.so/discord) server
