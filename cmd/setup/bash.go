package setup

import (
	"embed"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed bash-preexec.sh bash-hooks.sh
var bashFiles embed.FS

const bashSetupScript = `
# Savvy initialization for bash

# Function to source a file and check for errors
source_file() {
    if [[ -n "$1" ]]; then
        if ! source /dev/stdin <<< "$1"; then
            echo "Error sourcing script" >&2
            return 1
        fi
    else
        echo "Empty script content" >&2
        return 1
    fi
}

# Embed bash-preexec.sh content
read -r -d '' BASH_PREEXEC_CONTENT << 'EOF'
{{.BashPreexecContent}}
EOF

# Embed bash-hooks.sh content
read -r -d '' BASH_HOOKS_CONTENT << 'EOF'
{{.BashHooksContent}}
EOF

# Source bash-preexec.sh
if ! source_file "$BASH_PREEXEC_CONTENT"; then
    echo "Failed to source bash-preexec.sh" >&2
    return 1
fi

# Source bash-hooks.sh
if ! source_file "$BASH_HOOKS_CONTENT"; then
    echo "Failed to source bash-hooks.sh" >&2
    return 1
fi
`

// initCmd represents the init command
var BashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Output shell setup for bash",
	Long:  `Output shell setup for bash`,
	RunE:  runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {

	bashPreexecContent, err := bashFiles.ReadFile("bash-preexec.sh")
	if err != nil {
		return fmt.Errorf("failed to read bash-preexec.sh: %w", err)
	}

	// Read the content of bash-hooks.sh
	bashHooksContent, err := bashFiles.ReadFile("bash-hooks.sh")
	if err != nil {
		return fmt.Errorf("failed to read bash-hooks.sh: %w", err)
	}

	// Prepare the template data
	data := struct {
		BashPreexecContent string
		BashHooksContent   string
	}{
		BashPreexecContent: string(bashPreexecContent),
		BashHooksContent:   string(bashHooksContent),
	}

	// Parse the template
	tmpl, err := template.New("bash_setup").Parse(bashSetupScript)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template and write to stdout
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

var DashCmd = &cobra.Command{
	Use:   "dash",
	Short: "Output shell setup for dash",
	Long:  `Output shell setup for dash`,
	RunE:  runCmd,
}
