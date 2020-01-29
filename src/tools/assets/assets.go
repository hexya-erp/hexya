// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package assets

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hexya-erp/hexya/src/tools/logging"
)

// A Compiler can compile to CSS
type Compiler interface {
	// Compile CSS from the in stream to out.
	Compile(in io.Reader, out io.Writer, includePaths ...string) error
}

var log logging.Logger

const rxPreprocessImports = `(@import\s?['"]([^'"]+)['"](;?))`

// LessCompiler can compile Less to CSS
type LessCompiler struct{}

// Compile Less to CSS
func (lc LessCompiler) Compile(in io.Reader, out io.Writer, includePaths ...string) error {
	var includeDirective string
	if len(includePaths) > 0 {
		includeDirective = fmt.Sprintf("--include-path=%s", strings.Join(includePaths, ":"))
	}
	lessCmd := exec.Command("lessc", "-", "--no-js", "--no-color", includeDirective)
	return executeCmd(lessCmd, in, out)
}

var _ Compiler = LessCompiler{}

// ScssCompiler can compile Sass to CSS
type ScssCompiler struct{}

// Compile Sass/Scss to CSS
func (sc ScssCompiler) Compile(in io.Reader, out io.Writer, includePaths ...string) error {
	var (
		includeDirectives []string
		sanitized         bytes.Buffer
	)
	sc.sanitize(in, &sanitized)
	for _, includePath := range includePaths {
		includeDirectives = append(includeDirectives, "--load-path", includePath)
	}
	// DEBUG
	tmpFile, err := os.OpenFile("/tmp/sass", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	tr := io.TeeReader(&sanitized, tmpFile)
	defer tmpFile.Close()
	// ======
	scssCmd := exec.Command("sass", "--stdin", "--no-quiet")
	scssCmd.Args = append(scssCmd.Args, includeDirectives...)
	return executeCmd(scssCmd, tr, out)
}

// sanitize the import directives of the input SCSS
func (sc ScssCompiler) sanitize(in io.Reader, out io.Writer) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("error while reading input CSS: %v", err)
	}
	re, err := regexp.Compile(rxPreprocessImports)
	if err != nil {
		return err
	}
	imports := make(map[string]bool)
	re.ReplaceAllFunc(data, func(match []byte) []byte {
		line := bytes.ReplaceAll(match, []byte("'"), []byte("\""))
		if bytes.Contains(line, []byte(".")) || bytes.HasPrefix(line, []byte("@import \"/")) || bytes.HasPrefix(line, []byte("@import \"~")) {
			panic(fmt.Errorf("local import '%s' is forbidden for security reasons. "+
				"Please remove all @import {your_file} imports in your custom files. "+
				"In Hexya you have to import all files in the assets, and not through the @import statement", match))
		}
		if imports[string(line)] {
			return []byte{}
		}
		imports[string(line)] = true
		return line
	})
	out.Write(data)
	return nil
}

var _ Compiler = ScssCompiler{}

func executeCmd(cmd *exec.Cmd, in io.Reader, out io.Writer) error {
	log.Debug("Compiling assets", "cmd", cmd.String())
	lessIn, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to get compile command stdin pipe: %v", err)
	}
	lessOut, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to get compile command stdout pipe: %v", err)
	}
	lessErr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to get compile command stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start compile command: %v", err)
	}
	if _, err := io.Copy(lessIn, in); err != nil {
		return fmt.Errorf("unable to read input: %v", err)
	}
	if err := lessIn.Close(); err != nil {
		return err
	}
	var outPut bytes.Buffer
	teeReader := io.TeeReader(lessOut, &outPut)
	if _, err := io.Copy(out, teeReader); err != nil {
		return fmt.Errorf("unable to write output: %v", err)
	}
	var msg bytes.Buffer
	if _, err := io.Copy(&msg, lessErr); err != nil {
		return fmt.Errorf("unable to write on err pipe: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		msgString := msg.String()
		if msgString == "" {
			msgString = outPut.String()
		}
		return fmt.Errorf("error while executing compile command %s. Error: %s: %s\n", cmd.String(), err, msgString)
	}
	return nil
}

func init() {
	log = logging.GetLogger("assets")
}
