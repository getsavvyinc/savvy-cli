package kind

// Kind represents the type of shell.
type Kind string

// Define constants for ShellKind here, based on your needs, e.g., Bash, Zsh, etc.
const (
	Bash    Kind = "bash"
	Zsh     Kind = "zsh"
	Dash    Kind = "dash"
	Fish    Kind = "fish"
	Nu      Kind = "nu"
	Unknown Kind = "unknown"
)

// ShellKindFromString tries to match a string to a shell Kind.
func ShellKindFromString(name string) (Kind, bool) {
	switch name {
	case "bash":
		return Bash, true
	case "zsh":
		return Zsh, true
	case "dash":
		return Dash, true
	case "fish":
		return Fish, true
	case "nu":
		return Nu, true
	}
	return Unknown, false
}
