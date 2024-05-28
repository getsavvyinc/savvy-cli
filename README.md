# Savvy

<h3 align="left">
  | <a href="https://getsavvy.so/discord">Discord</a> |
  <a href="https://twitter.com/savvyoncall">Twitter</a> |
  <a href="https://www.getsavvy.so/">Website</a> |
</h3>

Savvy is the easiest way to create, share and run runbooks from your terminal.


Savvy's CLI generates runbooks with AI or from your provided commands.

## Install Savvy's CLI

```sh
curl -fsSL https://install.getsavvy.so | sh
```

## Generate Runbooks with AI

Generate runbooks using natural language or get help crafting a single command using `savvy ask`.

Any one can use it, there's no need to signup for an account or provide a credit card.

Just run `savvy ask` and provide a prompt.

### Examples

1. Ask Savvy to create a runbook for publishing a new go module.


![Ask Savvy to create a runbook for publishing a new go module.](demos/ask-runbook.gif)

2. Ask Savvy to help you with a tricky sequence of shell commands.

![Ask Savvy to help you with a tricky shell command](demos/ask-command.gif)


## Generate Runbooks using your commands

Use `savvy record` or `savvy record history` to create a runbook using commands you provide.

You don't have to change anything about your shell or aliases, savvy auto expands all aliases to make sure your runbook runs reliably on any machine.

### Create Runbooks from your Shell History

Use `savvy record history` to go back in time and create a runbooks by selecting just the commands you want.

Savvy will never execute any command you select.


![savvy record history](demos/savvy-history.gif)

### Sharing Runbooks

Runbooks are private by default, but you can share them using a public or unlisted link from Savvy's dashboard.

You can also export runbooks to markdown and paste them in your existing docs.

### Create Runbooks with Savvy Record

`savvy record` starts a new shell and all commands in this shell are recorded and sent to an LLM to generate a runbook.

![Savvy Runbook](https://vhs.charm.sh/vhs-1UmW0o6uSztF6b76y92K2K.gif)

> [!NOTE]
> Creating a runbook with savvy record requires you to signup for a free account.


##

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
