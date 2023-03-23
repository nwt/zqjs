package main

import (
	"bytes"
	"context"
	"strings"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/compiler"
	"github.com/brimdata/zed/pkg/storage"
	"github.com/brimdata/zed/runtime"
	"github.com/brimdata/zed/zbuf"
	"github.com/brimdata/zed/zio"
	"github.com/brimdata/zed/zio/anyio"
	"github.com/teamortix/golang-wasm/wasm"
)

func main() {
	wasm.Expose("zq", zq)
	wasm.Ready()
	<-make(chan struct{})
}

type opts struct {
	Program      string `wasm:"program"`
	Input        string `wasm:"input"`
	InputFormat  string `wasm:"inputFormat"`
	OutputFormat string `wasm:"outputFormat"`
}

func zq(opts opts) (string, error) {
	flowgraph, err := compiler.Parse(opts.Program)
	if err != nil {
		return "", err
	}

	zctx := zed.NewContext()
	zr, err := anyio.NewReaderWithOpts(zctx, strings.NewReader(opts.Input), anyio.ReaderOpts{
		Format: opts.InputFormat,
	})
	if err != nil {
		return "", err
	}
	defer zr.Close()

	var buf bytes.Buffer
	zwc, err := anyio.NewWriter(zio.NopCloser(&buf), anyio.WriterOpts{Format: opts.OutputFormat})
	if err != nil {
		return "", err
	}
	defer zwc.Close()

	local := storage.NewLocalEngine()
	comp := compiler.NewFileSystemCompiler(local)
	query, err := runtime.CompileQuery(context.Background(), zctx, comp, flowgraph, []zio.Reader{zr})
	if err != nil {
		return "", err
	}
	defer query.Pull(true)

	if err := zbuf.CopyPuller(zwc, query); err != nil {
		return "", err
	}
	if err := zwc.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
