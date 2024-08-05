#### SAVVY CUSTOMIZATIONS ####

# Enable experimental subshell support
export __bp_enable_subshells="true"


SAVVY_INPUT_FILE=/tmp/savvy-socket

# Save the original PS1
orignal_ps1=$PS1

get_user_prompt() {
  local user_prompt
  # P expansion is only available in bash 4.4+
  if [[ "${BASH_VERSINFO[0]}" -gt 4 ]] || (( BASH_VERSINFO[0] > 3  && BASH_VERSINFO[1] > 4)); then
    user_prompt=$(printf '%s' "${PS1@P}")
  else
    user_prompt=""
  fi
  echo "${user_prompt}"
}

step_id=""
savvy_cmd_pre_exec() {
  local expanded_command=""
  local spaced_command=$(echo $1 | sed -e 's/\$(\([^)]*\))/$( \1 )/g' -e 's/`\(.*\)`/` \1 `/g')
  local command_parts=( $spaced_command )
  for part in "${command_parts[@]}"; do
      if [[ "$part" =~ ^[a-zA-Z0-9_]+$ && $(type -t "$part") == "alias" ]]; then
        expanded_command+=$(alias "$part" | sed -e "s/^[[:space:]]*alias $part='//" -e "s/^$part='//" -e "s/'$//")" "
      else
          expanded_command+="$part "
      fi
  done
  local cmd="${expanded_command}"
  local prompt=$(get_user_prompt)
  step_id=""
  if [[ "${SAVVY_CONTEXT}" == "record" ]] ; then
    step_id=$(SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send --prompt="${prompt}" "$cmd")
  fi
}

savvy_cmd_pre_cmd() {
  local exit_code=$?

  if [[ "${SAVVY_CONTEXT}" == "record" && "$PS1" != *'recording'* ]]; then
    PS1+=$'\[\e[31m\]recording\[\e[0m\] \U1f60e '
  fi

  # if return code is not 0, send the return code to the server
  if [[ "${SAVVY_CONTEXT}" == "record" && "${exit_code}" != "0" ]] ; then
    SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send --step-id="${step_id}" --exit-code="${exit_code}"
  fi
}

SAVVY_COMMANDS=()
SAVVY_RUN_CURR=""
SAVVY_NEXT_STEP=0

# Set up a function to run the next command in the runbook when the user presses C-n
savvy_runbook_runner() {
  if [[ "${SAVVY_CONTEXT}" == "run"  && "${SAVVY_NEXT_STEP}" -le "${#SAVVY_COMMANDS[@]}" ]] ; then
    next_step=$(savvy internal current)
    READLINE_LINE="${next_step}"
    READLINE_POINT=${#READLINE_LINE}
  fi
}


savvy_run_pre_exec() {
  # we want the command as it was typed in.
  local cmd=$1
  if [[ "${SAVVY_CONTEXT}" == "run" && "${SAVVY_NEXT_STEP}" -lt "${#SAVVY_COMMANDS[@]}" ]] ; then
    SAVVY_NEXT_STEP=$(savvy internal next --cmd="${cmd}")
  fi
}

PROMPT_GREEN="\[$(tput setaf 2)\]"
PROMPT_BLUE="\[$(tput setaf 4)\]"
PROMPT_BOLD="\[$(tput bold)\]"
PROMPT_RED="\[$(tput setaf 1)\]"
PROMPT_RESET="\[$(tput sgr0)\]"

savvy_run_pre_cmd() {
  # transorm 0 based index to 1 based index
  local display_step=$((SAVVY_NEXT_STEP+1))
  local size=${#SAVVY_COMMANDS[@]}

  if [[ "${SAVVY_CONTEXT}" == "run" && "${SAVVY_NEXT_STEP}" -lt "${size}" && "${size}" -gt 0 ]] ; then
    PS1="${orignal_ps1}\n${PROMPT_GREEN}[ctrl+n:get next step]${PROMPT_RESET}(running ${PROMPT_BOLD}${SAVVY_RUN_CURR} ${display_step}/${size}${PROMPT_RESET}) "
  fi

  if [[ "${SAVVY_CONTEXT}" == "run" && "${SAVVY_NEXT_STEP}" -ge "${size}" ]] ; then
    # space at the end is important
    PS1="${orignal_ps1}\n(${PROMPT_GREEN}done${PROMPT_RESET}"$' \U1f60e '"${PROMPT_BOLD}${SAVVY_RUN_CURR}${PROMPT_RESET})${PROMPT_GREEN}[exit/ctrl+d to exit]${PROMPT_RESET} "
  fi

  if [[ "${SAVVY_CONTEXT}" == "run" && "${SAVVY_NEXT_STEP}" -lt "${size}" ]] ; then
    savvy internal set-param
  fi
}


if [[ "${SAVVY_CONTEXT}" == "run" ]] ; then
  mapfile -t SAVVY_COMMANDS < <(awk -F'COMMA' '{ for(i=1;i<=NF;i++) print $i }' <<< $SAVVY_RUNBOOK_COMMANDS)
  SAVVY_RUN_CURR="${SAVVY_RUNBOOK_ALIAS}"

  # Set up a keybinding to trigger the function
  bind 'set keyseq-timeout 0'
  bind -x '"\C-n":savvy_runbook_runner'

  precmd_functions+=(savvy_run_pre_cmd)
  preexec_functions+=(savvy_run_pre_exec)
fi;

preexec_functions+=(savvy_cmd_pre_exec)
# NOTE: If you change this function name, you must also change the corresponding check in shell/check_setup.go
# TODO: use templates to avoid the need to manually change shell checks
precmd_functions+=(savvy_cmd_pre_cmd)
