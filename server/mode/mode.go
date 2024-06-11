package mode

type Mode int

const (
	Record Mode = iota
	Run
)

func (mode Mode) String() string {
	switch mode {
	case Record:
		return "recording"
	case Run:
		return "run"
	default:
		return ""
	}
}
