# Source this in your ~/.zshrc
SAVVY_INPUT_FILE=/tmp/savvy-socket

autoload -Uz add-zsh-hook

# This function fixes the prompt via a precmd hook.
# function __savvy_cmd_pre_cmd__() {
#  local status=$?
#  echo "exit_status: ${status}</command>" > $SAVVY_STATUS_FILE
# }

function __savvy_cmd_pre_exec__() {
  # $2 is the command with all the aliases expanded
  local cmd=$3
  if [[ "${SAVVY_CONTEXT}" == "1" ]] ; then
      echo "${cmd}" | nc -U $SAVVY_INPUT_FILE
  fi
}
add-zsh-hook preexec __savvy_cmd_pre_exec__
