set SAVVY_INPUT_FILE /tmp/savvy-socket


# Fish automatically loads completions, so no need for 'autoload' or 'compinit'

# Load savvy completions
savvy completion fish | source

set -g step_id ""


# NOTE: If you change any function names, you must also change the corresponding check in shell/check_setup.go, shell/fish.go
#
# TODO: use templates to avoid the need to manually change shell checks

function __savvy_record_prompt --description "Modify prompt for Savvy recording"
    # Save the original prompt function if not already saved
    if not functions -q __pre_savvy_record_prompt
        functions -c fish_prompt __pre_savvy_record_prompt
    end

    # Define new fish_prompt function
    function fish_prompt
        # Call the original prompt function
        set -l original_prompt (__pre_savvy_record_prompt)

        if test "$SAVVY_CONTEXT" = "record"
          and not string match -q '*recording*' "$fish_prompt"
          echo -n $original_prompt
          echo -n (set_color green)"recording"(set_color normal)" 😎 "
        else
          echo -n $original_prompt
        end
    end
end

# Call the function to set up the modified prompt
__savvy_record_prompt



function __savvy_runbook_runner__ --on-event fish_prompt
    if test "$SAVVY_CONTEXT" = "run"
        set -l run_cmd (savvy internal current)
        commandline -r $run_cmd
        commandline -f end-of-line
    end
end


# Initialize variables
set -g SAVVY_COMMANDS ()
set -g SAVVY_RUN_CURR ""
set -g SAVVY_NEXT_STEP 0


function __savvy_run_prompt --description "Modify prompt for Savvy run"
    # Save the original prompt function if not already saved
    if not functions -q __pre_savvy_run_prompt
        set -g SAVVY_COMMANDS (string split ":COMMA:" $SAVVY_RUNBOOK_COMMANDS)
        functions -c fish_prompt __pre_savvy_run_prompt
    end

    if not functions -q fish_right_prompt
    # If fish_right_prompt doesn't exist, create an empty one
      function fish_right_prompt
        # Empty function
      end
    end

    if not functions -q __pre_savvy_run_right_prompt
      functions -c fish_right_prompt __pre_savvy_run_right_prompt
    end

    # Define new fish_prompt function
    function fish_prompt
        # Call the original prompt function
        set -l original_prompt (__pre_savvy_run_prompt)
        set -g SAVVY_RUN_CURR "$SAVVY_RUNBOOK_ALIAS"

        echo -n $original_prompt
        if test "$SAVVY_CONTEXT" = "run"
          echo -n (set_color green)"savvy run $SAVVY_RUN_CURR"(set_color normal)
        end

        if test "$SAVVY_CONTEXT" = "run"
          and test "$SAVVY_NEXT_STEP" -ge (count $SAVVY_COMMANDS)
            echo -n (set_color green)" done 😎"(set_color normal)
            echo -n (set_color red)" ctrl-d/exit to exit"(set_color normal)
        end
    end

    function fish_right_prompt
        set -l original_right_prompt (__pre_savvy_run_right_prompt)

        if test "$SAVVY_CONTEXT" = "run"
          and test (count $SAVVY_COMMANDS) -gt 0
          and test "$SAVVY_NEXT_STEP" -lt (count $SAVVY_COMMANDS)
            set -l num (math $SAVVY_NEXT_STEP + 1)
            echo -n (set_color green)"($num/(count $SAVVY_COMMANDS))"(set_color normal)
        end
        echo -n $original_right_prompt
    end
end

__savvy_run_prompt


function __savvy_record_post_exec --on-event fish_postexec
    set -l exit_code $status

    if not test "$SAVVY_CONTEXT" = "record"
        return
    end

    # Send the return code to the server if it's not 0
    if test "$SAVVY_CONTEXT" = "record"
      and test "$exit_code" -ne 0
        set -x SAVVY_SOCKET_PATH $SAVVY_INPUT_FILE
        savvy send --step-id="$step_id" --exit-code="$status"
    end
end



function __savvy_record_pre_exec__ --on-event fish_preexec
    if not test "$SAVVY_CONTEXT" = "record"
        return
    end

    # $argv[2] is the full command line in Fish
    set -l cmd $argv[1]

    # Clear step_id
    set -g step_id ""

    if test "$SAVVY_CONTEXT" = "record"
        # Get the current prompt
        #set -l prompt (fish_prompt)

        # Remove color codes and other formatting from the prompt
        #set -l clean_prompt (string replace -ra '\e\[[^m]*m' "" -- $prompt)

        # Send command to savvy and get step_id
        #savvy send --prompt="$clean_prompt" $cmd
        set -g step_id (
            env SAVVY_SOCKET_PATH=$SAVVY_INPUT_FILE \
            savvy send $cmd
        )
    end
end
