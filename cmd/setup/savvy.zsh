# Source this in your ~/.zshrc
SAVVY_INPUT_FILE=/tmp/savvy-socket

autoload -Uz add-zsh-hook

# This function fixes the prompt via a precmd hook.
 function __savvy_cmd_pre_cmd__() {
  if [[ "${SAVVY_CONTEXT}" == "1" && "$PS1" != *'recording'*  ]] ; then
      PS1+=$'%F{red}recording%f \U1f60e '
  fi
 }

function __savvy_cmd_pre_exec__() {
  # $2 is the command with all the aliases expanded
  local cmd=$3
  if [[ "${SAVVY_CONTEXT}" == "1" ]] ; then
     SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send "$cmd"
  fi
}
add-zsh-hook preexec __savvy_cmd_pre_exec__
add-zsh-hook precmd __savvy_cmd_pre_cmd__
