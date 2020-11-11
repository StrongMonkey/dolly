package cmd

import (
	"github.com/rancher/dolly/pkg/dollyfile"
	"github.com/rancher/dolly/pkg/template"
	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
)

func NewPushCommand() *cobra.Command {
	push := cli.Command(&Push{}, cobra.Command{
		Short: "Run docker build and push using dollyfile syntax",
	})
	return push
}

type Push struct {
	File       string `name:"file" usage:"Path to dollyfile, can point to local file path, https links or stdin(-)" default:"DollyFile" short:"f"`
	AnswerFile string `name:"answer-file" usage:"Answer file set for dollyfile" default:"DollyFile-answers" short:"a"`
}

func (p Push) Run(cmd *cobra.Command, args []string) error {
	content, answers, err := dollyfile.LoadFileAndAnswer(p.File, p.AnswerFile)
	if err != nil {
		return err
	}

	rf, err := dollyfile.Parse(content, "", template.AnswersFromMap(answers))
	if err != nil {
		return err
	}

	return rf.Build(true)
}
