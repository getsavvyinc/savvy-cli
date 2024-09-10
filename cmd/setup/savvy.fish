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
__savvy_modify_prompt



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
