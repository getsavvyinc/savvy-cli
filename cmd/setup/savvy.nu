# Savvy initialization for Nushell

# Function to record command history
def savvy_record_history_skip_execution [cmd: string] {
  $env.SAVVY_SOCKET_PATH = $env.SAVVY_INPUT_FILE
  savvy send $cmd | ignore
  exec nu
}

# Hook to capture commands before execution
$env.config = ($env.config | upsert hooks.pre_execution {
  [
    { condition: true, code: { |cmd| savvy_record_history_skip_execution $cmd } }
  ]
})

# Function to get next command from runbook
def savvy_next [] {
  $env.SAVVY_SOCKET_PATH = $env.SAVVY_INPUT_FILE
  savvy next
}

# Alias for convenience
alias sn = savvy_next
