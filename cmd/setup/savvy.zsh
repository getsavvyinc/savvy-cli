# Source this in your ~/.zshrc
SAVVY_INPUT_FILE=/tmp/savvy-socket

autoload -Uz add-zsh-hook

step_id=0

# This function fixes the prompt via a precmd hook.
 function __savvy_cmd_pre_cmd__() {
    echo "$step_id commands recorded. Press exit to stop recording."
  if [[ "${SAVVY_CONTEXT}" == "1" && "$PS1" != *'recording'*  ]] ; then
      PS1+=$'%F{red}recording%f \U1f60e '
  fi

  # if return code is not 0, send the return code to the server
  if [[ "${SAVVY_CONTEXT}" == "1" && "$?" != "0" ]] ; then
    SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send "{'step_id': \"${step_id}\", 'exit_status': \"$?\"}"
  fi

  # increment the step id
  ((step_id++))
 }

function __savvy_cmd_pre_exec__() {
  # $2 is the command with all the aliases expanded
  local cmd=$3
  if [[ "${SAVVY_CONTEXT}" == "1" ]] ; then
     SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send "{'step_id': \"${step_id}\", 'command': \"$cmd\"}"
  fi
}
add-zsh-hook preexec __savvy_cmd_pre_exec__
# NOTE: If you change this function name, you must also change the corresponding check in shell/check_setup.go
# TODO: use templates to avoid the need to manually change shell checks
add-zsh-hook precmd __savvy_cmd_pre_cmd__
