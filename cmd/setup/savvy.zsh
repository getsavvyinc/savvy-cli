# Source this in your ~/.zshrc
SAVVY_INPUT_FILE=/tmp/savvy-socket

autoload -Uz add-zsh-hook

step_id=""

# This function fixes the prompt via a precmd hook.
 function __savvy_record_pre_cmd__() {
   local exit_code=$?
  if [[ "${SAVVY_CONTEXT}" == "record" && "$PS1" != *'recording'*  ]] ; then
      PS1+=$'%F{red}recording%f \U1f60e '
  fi

  # if return code is not 0, send the return code to the server
  if [[ "${SAVVY_CONTEXT}" == "record" && "${exit_code}" != "0" ]] ; then
    SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send --step-id="${step_id}" --exit-code="${exit_code}"
  fi
 }

function __savvy_record_pre_exec__() {
  # $2 is the command with all the aliases expanded
  local cmd=$3
  local prompt=$(print -P "$PROMPT")
  # clear step_id
  step_id=""
  if [[ "${SAVVY_CONTEXT}" == "record" ]] ; then
    step_id=$(SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send --prompt="${prompt}" $cmd)
  fi
}

function __savvy_run_pre_exec__() {
  local cmd=$3
  if [[ "${SAVVY_CONTEXT}" == "run" ]] ; then
    if [[ "${cmd}" -eq "${SAVVY_COMMANDS[SAVVY_NEXT_STEP]}" ]] ; then
      export SAVVY_NEXT_STEP=$((SAVVY_NEXT_STEP+1))
    fi
  fi
}

# SAVVY_RUNBOOK_COMMANDS is a list of commands that savvy should run in the run context

SAVVY_COMMANDS="(${(s:,:)SAVVY_RUNBOOK_COMMANDS}")
num_commands=${#SAVVY_COMMANDS}
function __savvy_runbook_runner__() {
  next_step_idx=${SAVVY_NEXT_STEP:1}
  BUFFER=${SAVVY_COMMANDS[next_step_idx]}  # Initial text to be edited by the user
  zle end-of-line  # Accept the line for editing
}



# NOTE: If you change any function names, you must also change the corresponding check in shell/check_setup.go, shell/zsh.go
#
# TODO: use templates to avoid the need to manually change shell checks

if [[ "${SAVVY_CONTEXT}" == "run" ]] ; then
  zle -N zle-line-init __savvy_runbook_runner__
  add-zle-hook-widget zle-line-init
fi

add-zsh-hook preexec __savvy_record_pre_exec__
add-zsh-hook preexec __savvy_run_pre_exec__

add-zsh-hook precmd __savvy_record_pre_cmd__
