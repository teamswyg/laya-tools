// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package claimsdemo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

// Run loads one explicit local trained .rsc artifact and serves until interrupted.
// Model inference is read-only. Optional Go profiles are written only locally.
// No downloads, training, or browser launch occur.
func Run(args []string, out, errOut io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return run(ctx, args, out, errOut)
}

func run(ctx context.Context, args []string, out, errOut io.Writer) (runErr error) {
	f := flag.NewFlagSet("riidolaya demo", flag.ContinueOnError)
	f.SetOutput(errOut)
	modelPath := f.String("model", "", "required local trained three-claim .rsc model")
	listen := f.String("listen", "127.0.0.1:8877", "local address: localhost, 127.0.0.1, or [::1] with a port")
	modelSHA := f.String("model-sha256", "", "optional required 64-digit SHA-256 checked before loading")
	cpuProfile := f.String("cpu-profile", "", "optional new local Go CPU profile file; 16 MiB limit")
	heapProfile := f.String("heap-profile", "", "optional new local post-GC Go heap profile on shutdown; 16 MiB limit")
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return errors.New("demo: invalid command options")
	}
	if f.NArg() != 0 || *modelPath == "" {
		return errors.New("demo: --model is required and positional arguments are not accepted")
	}
	listenAddress, err := normalizeListen(*listen)
	if err != nil {
		return err
	}
	model, digest, err := loadModel(*modelPath, *modelSHA)
	if err != nil {
		return err
	}
	if ctx.Err() != nil {
		return nil
	}
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return errors.New("demo: cannot listen on the requested local port")
	}
	defer listener.Close()
	bound, ok := listener.Addr().(*net.TCPAddr)
	if !ok || !bound.IP.IsLoopback() {
		return errors.New("demo: the listener did not bind to a loopback address")
	}
	server := &http.Server{
		Handler:           NewHandler(model, digest),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    8 << 10,
		ErrorLog:          log.New(io.Discard, "", 0),
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	profiles, err := startProfiles(*cpuProfile, *heapProfile)
	if err != nil {
		return err
	}
	defer func() {
		if err := profiles.finish(server); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()
	if _, err := fmt.Fprintf(out, "Local research preview: http://%s/\nClaim hints are not verified task state. Press Ctrl+C to stop.\n", listener.Addr().String()); err != nil {
		return errors.New("demo: cannot write startup status")
	}
	finished := make(chan error, 1)
	go func() { finished <- server.Serve(listener) }()
	select {
	case err := <-finished:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return errors.New("demo: local HTTP server stopped unexpectedly")
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
		}
		<-finished
		return nil
	}
}

func validPort(port string) bool {
	if port == "" {
		return false
	}
	for _, r := range port {
		if r < '0' || r > '9' {
			return false
		}
	}
	n, err := strconv.ParseUint(port, 10, 16)
	return err == nil && n <= 65535
}

func normalizeListen(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || !loopbackHost(host) || !validPort(port) {
		return "", errors.New("demo: --listen must use localhost, 127.0.0.1, or [::1] with a numeric port")
	}
	// Bind an IP literal so custom hosts/DNS cannot resolve localhost externally.
	if strings.EqualFold(host, "localhost") {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port), nil
}

func loadModel(path, expectedSHA string) (*statehintclaims.Model, string, error) {
	if expectedSHA != "" {
		decoded, err := hex.DecodeString(expectedSHA)
		if err != nil || len(decoded) != sha256.Size {
			return nil, "", errors.New("demo: --model-sha256 must contain exactly 64 hexadecimal digits")
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", errors.New("demo: cannot open the explicit local model")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() != statehintclaims.ArtifactBytes {
		return nil, "", errors.New("demo: model must be a regular .rsc file with the exact artifact size")
	}
	data, err := io.ReadAll(io.LimitReader(f, int64(statehintclaims.ArtifactBytes)+1))
	if err != nil || len(data) != statehintclaims.ArtifactBytes {
		return nil, "", errors.New("demo: cannot read a complete .rsc model")
	}
	digest := sha256.Sum256(data)
	actualSHA := hex.EncodeToString(digest[:])
	if expectedSHA != "" {
		expected, _ := hex.DecodeString(expectedSHA)
		if !bytes.Equal(expected, digest[:]) {
			return nil, "", errors.New("demo: model SHA-256 mismatch")
		}
	}
	model, err := statehintclaims.Load(bytes.NewReader(data))
	if err != nil {
		return nil, "", errors.New("demo: invalid or corrupted three-claim model")
	}
	if model.TrainingSteps() == 0 {
		return nil, "", errors.New("demo: the local model must have trained weights; an untrained scaffold cannot serve the demo")
	}
	return model, actualSHA, nil
}
