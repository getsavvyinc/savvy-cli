# Source this in your ~/.zshrc
SAVVY_INPUT_FILE=/tmp/savvy-socket

autoload -Uz add-zsh-hook

step_id=""

# This function fixes the prompt via a precmd hook.
 function __savvy_cmd_pre_cmd__() {
   local exit_code=$?
  if [[ "${SAVVY_CONTEXT}" == "1" && "$PS1" != *'recording'*  ]] ; then
      PS1+=$'%F{red}recording%f \U1f60e '
  fi

  # if return code is not 0, send the return code to the server
  if [[ "${SAVVY_CONTEXT}" == "1" && "${exit_code}" != "0" ]] ; then
    SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send --step-id="${step_id}" --exit-code="${exit_code}"
  fi
 }

function __savvy_cmd_pre_exec__() {
  # $2 is the command with all the aliases expanded
  local cmd=$3
  # clear step_id
  step_id=""
  if [[ "${SAVVY_CONTEXT}" == "1" ]] ; then
    step_id=$(SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send $cmd)
  fi
}
add-zsh-hook preexec __savvy_cmd_pre_exec__
# NOTE: If you change this function name, you must also change the corresponding check in shell/check_setup.go
# TODO: use templates to avoid the need to manually change shell checks
add-zsh-hook precmd __savvy_cmd_pre_cmd__
