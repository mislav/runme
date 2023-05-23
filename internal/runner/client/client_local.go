package client

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/stateful/runme/internal/document"
	"github.com/stateful/runme/internal/env"
	runnerv1 "github.com/stateful/runme/internal/gen/proto/go/runme/runner/v1"
	"github.com/stateful/runme/internal/project"
	"github.com/stateful/runme/internal/runner"
	"go.uber.org/zap"
)

type LocalRunner struct {
	dir    string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	shellID int

	session *runner.Session
	project project.Project

	customShell string

	logger *zap.Logger
}

func (r *LocalRunner) Clone() Runner {
	return &LocalRunner{
		dir: r.dir,

		stdin:  r.stdin,
		stdout: r.stdout,
		stderr: r.stderr,

		shellID: r.shellID,
		session: r.session,

		logger: r.logger,
	}
}

func (r *LocalRunner) setSession(s *runner.Session) error {
	r.session = s
	return nil
}

func (r *LocalRunner) setProject(p project.Project) error {
	r.project = p
	return nil
}

func (r *LocalRunner) setSessionID(sessionID string) error {
	return nil
}

func (r *LocalRunner) setCleanupSession(cleanup bool) error {
	return nil
}

func (r *LocalRunner) setSessionStrategy(runnerv1.SessionStrategy) error {
	return nil
}

func (r *LocalRunner) setWithinShell() error {
	id, ok := shellID()
	if !ok {
		return nil
	}
	r.shellID = id
	return nil
}

func (r *LocalRunner) setDir(dir string) error {
	r.dir = dir
	return nil
}

func (r *LocalRunner) setStdin(stdin io.Reader) error {
	r.stdin = stdin
	return nil
}

func (r *LocalRunner) setStdout(stdout io.Writer) error {
	r.stdout = stdout
	return nil
}

func (r *LocalRunner) setStderr(stderr io.Writer) error {
	r.stderr = stderr
	return nil
}

func (r *LocalRunner) getStdin() io.Reader {
	return r.stdin
}

func (r *LocalRunner) getStdout() io.Writer {
	return r.stdout
}

func (r *LocalRunner) getStderr() io.Writer {
	return r.stderr
}

func (r *LocalRunner) setLogger(logger *zap.Logger) error {
	r.logger = logger
	return nil
}

func (r *LocalRunner) setInsecure(bool) error {
	return nil
}

func (r *LocalRunner) setTLSDir(string) error {
	return nil
}

func (r *LocalRunner) setEnableBackgroundProcesses(bool) error {
	return nil
}

func (r *LocalRunner) setCustomShell(shell string) error {
	r.customShell = shell
	return nil
}

func NewLocalRunner(opts ...RunnerOption) (*LocalRunner, error) {
	r := &LocalRunner{}
	if err := ApplyOptions(r, opts...); err != nil {
		return nil, err
	}

	if r.logger == nil {
		r.logger = zap.NewNop()
	}

	r.session = runner.NewSession(os.Environ(), r.logger)

	return r, nil
}

func (r *LocalRunner) newExecutable(fileBlock project.FileCodeBlock) (runner.Executable, error) {
	block := fileBlock.GetBlock()
	fmtr := fileBlock.GetFrontmatter()

	customShell := r.customShell
	if fmtr.Shell != "" {
		customShell = fmtr.Shell
	}

	cfg := &runner.ExecutableConfig{
		Name:    block.Name(),
		Dir:     r.dir,
		Tty:     block.Interactive(),
		Stdout:  r.stdout,
		Stderr:  r.stderr,
		Session: r.session,
		Logger:  r.logger,
	}

	if r.project != nil {
		projEnvs, err := r.project.LoadEnvs()
		if err != nil {
			return nil, err
		}

		cfg.PreEnv = env.ConvertMapEnv(projEnvs)
	}

	mdFile := fileBlock.GetFile()
	if mdFile != "" {
		cfg.Dir = filepath.Join(r.dir, filepath.Dir(mdFile))
	}

	if block.Interactive() {
		cfg.Stdin = r.stdin
	}

	switch block.Language() {
	// TODO(mxs): empty string should return nil when guesslang model is implemented
	case "bash", "bat", "sh", "shell", "zsh", "":
		return &runner.Shell{
			ExecutableConfig: cfg,
			Cmds:             block.Lines(),
			CustomShell:      customShell,
		}, nil
	case "sh-raw":
		return &runner.ShellRaw{
			Shell: &runner.Shell{
				ExecutableConfig: cfg,
				Cmds:             block.Lines(),
			},
		}, nil
	case "go":
		return &runner.Go{
			ExecutableConfig: cfg,
			Source:           string(block.Content()),
		}, nil
	default:
		return nil, nil
	}
}

func (r *LocalRunner) RunBlock(ctx context.Context, fileBlock project.FileCodeBlock) error {
	block := fileBlock.GetBlock()

	if r.shellID > 0 {
		return r.runBlockInShell(ctx, block)
	}

	executable, err := r.newExecutable(fileBlock)
	if err != nil {
		return err
	}

	if executable == nil {
		return errors.Errorf("unknown executable: %q", block.Language())
	}

	// poll for exit
	// TODO(mxs): we probably want to use `StdinPipe` eventually
	if block.Interactive() {
		go func() {
			for {
				if executable.ExitCode() > -1 {
					if closer, ok := r.stdin.(io.ReadCloser); ok {
						_ = closer.Close()
					}

					return
				}

				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	return errors.WithStack(executable.Run(ctx))
}

func (r *LocalRunner) runBlockInShell(ctx context.Context, block *document.CodeBlock) error {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", "/tmp/runme-"+strconv.Itoa(r.shellID)+".sock")
	if err != nil {
		return errors.WithStack(err)
	}
	for _, line := range block.Lines() {
		line = strings.TrimSpace(line)

		if _, err := conn.Write([]byte(line)); err != nil {
			return errors.WithStack(err)
		}
		if _, err := conn.Write([]byte("\n")); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (r *LocalRunner) DryRunBlock(ctx context.Context, fileBlock project.FileCodeBlock, w io.Writer, opts ...RunnerOption) error {
	block := fileBlock.GetBlock()

	executable, err := r.newExecutable(block)
	if err != nil {
		return err
	}

	executable.DryRun(ctx, w)

	return nil
}

func (r *LocalRunner) Cleanup(ctx context.Context) error {
	return nil
}

func shellID() (int, bool) {
	id := os.Getenv("RUNMESHELL")
	if id == "" {
		return 0, false
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return -1, false
	}
	return i, true
}
