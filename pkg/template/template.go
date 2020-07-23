package template

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/drone/envsubst"
	"github.com/rancher/dolly/pkg/gotemplate"
	v1 "github.com/rancher/dolly/pkg/types"
	"github.com/rancher/wrangler/pkg/merr"
	"github.com/rancher/wrangler/pkg/yaml"
	"github.com/sirupsen/logrus"
)

type Template struct {
	Content     []byte
	BuiltinVars []string
}

type AnswerCallback func(key string) (string, error)

func AnswersFromMap(answers map[string]string) AnswerCallback {
	return func(key string) (string, error) {
		return answers[key], nil
	}
}

type templateFile struct {
	Meta v1.TemplateMeta `json:"template"`
}

func (t *Template) RequiredEnv() ([]string, error) {
	names := map[string]bool{}
	_, err := envsubst.Eval(string(t.Content), func(in string) string {
		names[in] = true
		return in
	})
	if err != nil {
		return nil, err
	}

	for _, b := range t.BuiltinVars {
		delete(names, b)
	}

	var result []string
	for key := range names {
		result = append(result, key)
	}

	return result, nil
}

func (t *Template) readTemplateFile(content []byte) (*templateFile, error) {
	templateFile := &templateFile{}
	if err := yaml.Unmarshal(content, templateFile); err != nil {
		logrus.Debugf("Error unmarshalling template: %v", err)
	}
	return templateFile, nil
}

func (t *Template) Parse(answers AnswerCallback) ([]byte, error) {
	return t.parseContent(answers)
}

func (t *Template) afterTemplate(content []byte) []byte {
	found := false

	result := bytes.Buffer{}
	scan := bufio.NewScanner(bytes.NewReader(content))
	for scan.Scan() {
		if strings.HasPrefix(string(scan.Bytes()), "template:") {
			found = true
		}
		if found {
			result.Write(scan.Bytes())
			result.WriteRune('\n')
		}
	}

	if found {
		return result.Bytes()
	}
	return content
}

func (t *Template) readTemplate() (*templateFile, error) {
	content, err := gotemplate.Apply(t.afterTemplate(t.Content), nil)
	if err != nil {
		return nil, nil
	}

	return t.readTemplateFile(content)
}

func (t *Template) parseContent(answersCB AnswerCallback) ([]byte, error) {
	template, err := t.readTemplate()
	if err != nil {
		return nil, err
	}
	if template == nil {
		return t.Content, nil
	}

	var (
		callbackErrs []error
		answers      = map[string]string{}
		evaled       = string(t.Content)
	)

	if template.Meta.EnvSubst {
		evaled, err = envsubst.Eval(evaled, func(key string) string {
			if answersCB == nil {
				return ""
			}
			val, err := answersCB(key)
			if err != nil {
				callbackErrs = append(callbackErrs, err)
			}
			answers[key] = val
			return val
		})
	}

	for _, key := range template.Meta.Variables {
		val, err := answersCB(key)
		if err != nil {
			return nil, err
		}
		answers[key] = val
	}

	if err != nil {
		return nil, err
	} else if len(callbackErrs) > 0 {
		return nil, merr.NewErrors(callbackErrs...)
	}

	if template.Meta.GoTemplate {
		return gotemplate.Apply([]byte(evaled), answers)
	}

	return []byte(evaled), nil
}
