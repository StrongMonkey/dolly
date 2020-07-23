package gotemplate

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/rancher/dolly/pkg/gotemplate/funcs"
)

func Apply(contents []byte, variables map[string]string) ([]byte, error) {
	// Skip templating if contents begin with '# notemplating'
	trimmedContents := strings.TrimSpace(string(contents))
	if strings.HasPrefix(trimmedContents, "#notemplating") || strings.HasPrefix(trimmedContents, "# notemplating") {
		return contents, nil
	}

	funcMaps := sprig.TxtFuncMap()
	funcMaps["splitPreserveQuotes"] = funcs.SplitPreserveQuotes
	funcMaps["flat"] = funcs.Flat

	t, err := template.New("template").Funcs(funcMaps).Parse(string(contents))
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	err = t.Execute(&buf, map[string]interface{}{
		"Values": variables,
	})
	return buf.Bytes(), err
}
