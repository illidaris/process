package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/illidaris/core"
)

func NewCommand() *Command {
	return &Command{
		Timeout: time.Second * 5,
	}
}

type Command struct {
	WorkSpace string
	CmdStr    string
	Args      []string
	Timeout   time.Duration
	Notify    NotifyCallback
	Before    []func()
	After     []func()
	log       core.ICtxLogger
}

func (c *Command) WithWorkSpace(v string) *Command {
	c.WorkSpace = v
	return c
}

func (c *Command) WithCmdStr(v string) *Command {
	c.CmdStr = v
	return c
}

func (c *Command) WithArgs(args ...string) *Command {
	if c.Args == nil {
		c.Args = make([]string, 0)
	}
	c.Args = append(c.Args, args...)
	return c
}

func (c *Command) WithTimeout(v time.Duration) *Command {
	c.Timeout = v
	return c
}

func (c *Command) WithNotify(v NotifyCallback) *Command {
	c.Notify = v
	return c
}

func (c *Command) WithBefore(fs ...func()) *Command {
	if c.Before == nil {
		c.Before = make([]func(), 0)
	}
	c.Before = append(c.Before, fs...)
	return c
}

func (c *Command) WithAfter(fs ...func()) *Command {
	if c.After == nil {
		c.After = make([]func(), 0)
	}
	c.After = append(c.After, fs...)
	return c
}

func (c *Command) WithLogger(v core.ICtxLogger) *Command {
	c.log = v
	return c
}

func (c *Command) Run(ctx context.Context) error {
	if c.Before != nil {
		for _, f := range c.Before {
			if f != nil {
				f()
			}
		}
	}
	defer func() {
		if c.After != nil {
			for _, f := range c.After {
				if f != nil {
					f()
				}
			}
		}
	}()
	timeoutCtx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	return c.invoke(timeoutCtx)
}

func (c *Command) invoke(ctx context.Context) error {
	app := exec.CommandContext(ctx, c.CmdStr, c.Args...)
	app.Dir = c.WorkSpace
	if c.Notify != nil {
		stdout, err := app.StdoutPipe()
		if err != nil {
			return err
		}
		c.Notify(stdout)
	} else {
		app.Stdout = os.Stdout
		app.Stderr = os.Stderr
	}
	if c.log != nil {
		c.log.InfoCtxf(ctx, "%s %v timeout=%fs", c.CmdStr, c.Args, c.Timeout.Seconds())
	}
	if execErr := app.Start(); execErr != nil {
		return execErr
	}
	defer func() {
		_ = app.Process.Kill()
	}()
	errCh := make(chan ExecResult)
	go func() {
		defer close(errCh)
		execErr := app.Wait()
		res := ExecResult{}
		if execErr != nil {
			res.Error = execErr
		} else {
			res.Ok = true
		}
		errCh <- res
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("PID[%d]%s %v,timeout=%fs,error=%s", app.Process.Pid, c.CmdStr, c.Args, c.Timeout.Seconds(), ctx.Err())
	case c := <-errCh:
		if c.Ok {
			return nil
		} else {
			return c.Error
		}
	}
}

type ExecResult struct {
	Ok    bool
	Error error
}

type NotifyCallback func(reader io.ReadCloser)

func ExecApp(ctx context.Context, cmd string, workspace string, timeout time.Duration, notify NotifyCallback, args ...string) error {
	proc, err := os.StartProcess(cmd, args, nil)
	if err != nil {
		return err
	}
	_, err = proc.Wait()
	if err != nil {
		return err
	}
	return nil
}
