//go:build dev

package config

func init() {
	apiHost = "http://localhost:8080"
	dashboardHost = "http://localhost:5173"
}
