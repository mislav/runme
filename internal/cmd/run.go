package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/stateful/runme/internal/document"

	"github.com/pkg/errors"
	"github.com/rwtodd/Go.Sed/sed"
	"github.com/spf13/cobra"
	"github.com/stateful/runme/internal/runner"
)

func runCmd() *cobra.Command {
	var (
		dryRun         bool
		replaceScripts []string
	)

	cmd := cobra.Command{
		Use:               "run",
		Aliases:           []string{"exec"},
		Short:             "Run a selected command.",
		Long:              "Run a selected command identified based on its unique parsed name.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: validCmdNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			blocks, err := getCodeBlocks()
			if err != nil {
				return err
			}

			block, err := lookupCodeBlock(blocks, args[0])
			if err != nil {
				return err
			}

			if err := replace(replaceScripts, block.Lines()); err != nil {
				return err
			}

			executable, err := newExecutable(cmd, block)
			if err != nil {
				return err
			}

			ctx, cancel := sigCtxCancel(cmd.Context())
			defer cancel()

			if dryRun {
				executable.DryRun(ctx, cmd.ErrOrStderr())
				return nil
			}

			return errors.WithStack(executable.Run(ctx))
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the final command without executing.")
	cmd.Flags().StringArrayVarP(&replaceScripts, "replace", "r", nil, "Replace instructions using sed.")

	return &cmd
}

func newExecutable(cmd *cobra.Command, block *document.CodeBlock) (runner.Executable, error) {
	base := &runner.Base{
		Dir:    chdir,
		Stdin:  cmd.InOrStdin(),
		Stdout: cmd.OutOrStdout(),
		Stderr: cmd.ErrOrStderr(),
	}

	switch block.Language() {
	case "sh", "shell":
		return &runner.Shell{
			Cmds: block.Lines(),
			Base: base,
		}, nil
	case "sh-raw":
		return &runner.ShellRaw{
			Cmds: block.Lines(),
			Base: base,
		}, nil
	case "go":
		log.Printf("source: %s", block.Content())
		return &runner.Go{
			Source: string(block.Content()),
			Base:   base,
		}, nil
	default:
		return nil, errors.Errorf("unknown executable: %q", block.Language())
	}
}

func sigCtxCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	return ctx, cancel
}

func replace(scripts []string, lines []string) error {
	if len(scripts) == 0 {
		return nil
	}

	for _, script := range scripts {
		engine, err := sed.New(strings.NewReader(script))
		if err != nil {
			return errors.Wrapf(err, "failed to compile sed script %q", script)
		}

		for idx, line := range lines {
			var err error
			lines[idx], err = engine.RunString(line)
			if err != nil {
				return errors.Wrapf(err, "failed to run sed script %q on line %q", script, line)
			}
		}
	}

	return nil
}
