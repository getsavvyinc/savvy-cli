package markdown

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/atotto/clipboard"
	"github.com/getsavvyinc/savvy-cli/display"
)

const MdTemplate = `I used [Savvy's CLI]({{ .URL }}) to record these commands:

{{- printf "\n" -}}
{{- range $i, $command := .Commands }}

 ~~~sh
 {{ add $i 1 }}. {{ $command }}
 ~~~

{{- printf "\n" -}}
{{- end -}}
`

type Service interface {
	ToMarkdownFile(ctx context.Context, commands []string) error
}

var mdTemplate *template.Template

func init() {
	mdTemplate = template.Must(template.New("md").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		}}).Parse(MdTemplate))
}

type svc struct {
	url string
}

func NewService() Service {
	return &svc{
		url: "https://github.com/getsavvyinc/savvy-cli",
	}
}

func (s *svc) ToMarkdownFile(ctx context.Context, commands []string) error {
	data := struct {
		Commands []string
		URL      string
	}{
		URL:      s.url,
		Commands: commands,
	}

	var buf bytes.Buffer
	if err := mdTemplate.Execute(&buf, data); err != nil {
		err = fmt.Errorf("error executing markdown template: %w", err)
		return err
	}

	mdContent := buf.String()

	defer func() {
		if cerr := clipboard.WriteAll(mdContent); cerr != nil {
			display.Info("Wrote md contents to clipboard")
		}
	}()

	humanReadableTime := time.Now().Format("2006_01_02_15:04:05")
	fileName := fmt.Sprintf("savvy_%s.md", humanReadableTime)
	f, err := os.Create(fileName)
	if err != nil {
		err = fmt.Errorf("failed to create mardown file: %w", err)
		return err
	}

	defer f.Close()
	if _, err := f.WriteString(mdContent); err != nil {
		err = fmt.Errorf("failed to write md to file: %w", err)
		return err
	}

	display.Info(fmt.Sprintf("Markdown file created: %s", fileName))
	return nil
}
