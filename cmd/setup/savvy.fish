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
          echo -n (set_color green)"recording"(set_color normal)" 😎>"
          echo -n "  "
        else
          echo -n $original_prompt
        end
    end
end

# Call the function to set up the modified prompt
__savvy_record_prompt



# Initialize variables
set -g SAVVY_COMMANDS ()
set -g SAVVY_RUN_CURR ""
set -g SAVVY_NEXT_STEP 0

function __savvy_run_pre_exec__ --on-event fish_preexec
    if not test "$SAVVY_CONTEXT" = "run"
        return
    end

    set -l cmd $argv[1]

    if test "$SAVVY_CONTEXT" = "run"
        set -g SAVVY_NEXT_STEP (savvy internal next --cmd="$cmd")
    end
end

function __savvy_run_prompt --description "Modify prompt for Savvy run"
    # Save the original prompt function if not already saved
    if not functions -q __pre_savvy_run_prompt
        set -g SAVVY_COMMANDS (string split "COMMA" $SAVVY_RUNBOOK_COMMANDS)
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
          echo -n (set_color --bold)"[ctrl-n: get next step]"(set_color normal)
          echo -n (set_color green)"(savvy run" (set_color normal) "$SAVVY_RUN_CURR) "
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
            set -l count (count $SAVVY_COMMANDS)
            echo -n (set_color green)"($num/$count)" (set_color normal)
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

    set -l cmd $argv[1]

    # Clear step_id
    set -g step_id ""

    if test "$SAVVY_CONTEXT" = "record"
        set -g step_id (
            env SAVVY_SOCKET_PATH=$SAVVY_INPUT_FILE \
            savvy send $cmd
        )
    end
end


function check_savvy_record_setup
    if not test "$SAVVY_CONTEXT" = "record"
        return
    end

    if not functions -q __savvy_record_pre_exec__
    set_color red
    echo "Your recording shell is not configured to use Savvy. Please run the following commands: "
    set_color normal
    echo
    set_color red echo "> echo 'savvy init fish | source' >> ~/.config/fish/config.fish" set_color normal
    set_color red echo "> source ~/.config/fish/config.fish" set_color normal
    echo
    exit 1
    end
end

check_savvy_record_setup


function __savvy_run_completion__
    if not test "$SAVVY_CONTEXT" = "run"
        return
    end

    set -l run_cmd (savvy internal current)
    set -l cmd (commandline -o)

    if test -z "$cmd"
        # Completions for empty command line
        commandline --replace $run_cmd
        return
    end
end

function trigger_run_autocomplete --on-event fish_complete_command
    if not test "$SAVVY_CONTEXT" = "run"
        return
    end
    bind \cn '__savvy_run_completion__'
end

trigger_run_autocomplete
