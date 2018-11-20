// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package assets

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/hexya-erp/hexya/src/tools/logging"
)

// LessCompilerCommand is the executable lessc to be called to compile less files
const LessCompilerCommand string = "lessc"

var log logging.Logger

// CompileLess reads text/less input from the in Reader and
// writes compiled text/css to the out Writer
func CompileLess(in io.Reader, out io.Writer, includePaths ...string) error {
	var includeDirective string
	if len(includePaths) > 0 {
		includeDirective = fmt.Sprintf("--include-path=%s", strings.Join(includePaths, ":"))
	}
	lessCmd := exec.Command(LessCompilerCommand, "-", "--no-js", "--no-color", includeDirective)
	log.Debug("Compiling assets", "cmd", lessCmd.Path, "args", lessCmd.Args)
	lessIn, err := lessCmd.StdinPipe()
	if err != nil {
		return err
	}
	lessOut, err := lessCmd.StdoutPipe()
	if err != nil {
		return err
	}
	lessErr, err := lessCmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := lessCmd.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(lessIn, in); err != nil {
		return err
	}
	if err := lessIn.Close(); err != nil {
		return err
	}
	if _, err := io.Copy(out, lessOut); err != nil {
		return err
	}
	var msg bytes.Buffer
	if _, err := io.Copy(&msg, lessErr); err != nil {
		return err
	}
	if err := lessCmd.Wait(); err != nil {
		return fmt.Errorf("%s: %s\n", err, msg.String())
	}
	return nil
}

func init() {
	log = logging.GetLogger("assets")
}
