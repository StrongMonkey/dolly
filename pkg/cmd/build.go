package cmd

import (
	"github.com/rancher/dolly/pkg/dollyfile"
	"github.com/rancher/dolly/pkg/template"
	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
)

func NewBuildCommand() *cobra.Command {
	build := cli.Command(&Build{}, cobra.Command{
		Short: "Run docker build using riofile syntax",
	})
	return build
}

type Build struct {
	File       string `name:"file" usage:"Path to riofile, can point to local file path, https links or stdin(-)" default:"DollyFile" short:"f"`
	AnswerFile string `name:"answer-file" usage:"Answer file set for riofile" default:"DollyFile-answers" short:"a"`
}

func (b Build) Run(cmd *cobra.Command, args []string) error {
	content, answers, err := dollyfile.LoadFileAndAnswer(b.File, b.AnswerFile)
	if err != nil {
		return err
	}

	rf, err := dollyfile.Parse(content, "", template.AnswersFromMap(answers))
	if err != nil {
		return err
	}

	return rf.Build(false)
}
