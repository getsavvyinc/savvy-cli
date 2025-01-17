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
	"github.com/getsavvyinc/savvy-cli/extension"
	"github.com/getsavvyinc/savvy-cli/slice"
)

const MdTemplate = `I used [Savvy's CLI]({{ .URL }}) to record these{{ if gt (len .Links) 0 }} commands and links{{ else }} commands{{ end }}:

{{- printf "\n" -}}
{{- if gt (len .Links) 0 }}

### Relevant Links
{{- range .Links }}
* [{{ .Title }}]({{ .URL }})
{{- end }}
{{- end -}}
----

### Relevant Links
{{- range $i, $command := .Commands }}

 ~~~sh
 {{ add $i 1 }}. {{ $command }}
 ~~~

{{- printf "\n" -}}
{{- end -}}
`

type Service interface {
	ToMarkdownFile(ctx context.Context, commands []string, links []extension.HistoryItem) error
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

type TitleURL struct {
	Title string
	URL   string
}

func (s *svc) ToMarkdownFile(ctx context.Context, commands []string, links []extension.HistoryItem) error {
	data := struct {
		Commands []string
		URL      string
		Links    []TitleURL
	}{
		URL:      s.url,
		Commands: commands,
		Links: slice.Map(links, func(item extension.HistoryItem) TitleURL {
			return TitleURL{Title: item.Title, URL: item.URL}
		}),
	}

	var buf bytes.Buffer
	if err := mdTemplate.Execute(&buf, data); err != nil {
		err = fmt.Errorf("error executing markdown template: %w", err)
		return err
	}

	mdContent := buf.String()

	defer func() {
		if cerr := clipboard.WriteAll(mdContent); cerr != nil {
			err := fmt.Errorf("failed to write md contents to clipboard: %w", cerr)
			display.Error(err)
		}
		display.Info("Wrote md contents to clipboard")
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
