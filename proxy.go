package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"sync"
	"time"

	"github.com/bep/debounce"
)

// The Proxy manages passing requests to the target application. Duties include
// building and running the target application, when necessary.
type Proxy struct {
	// Context holds a function that returns a base context to the caller.
	// The context can be used to shut down any running operations.
	Context context.Context

	// The BuildCommand is the command that is executed when the target
	// is ready to be built.
	BuildCommand string

	// The RunCommand is the command that is executed when the target has finished
	// building and is ready to go.
	RunCommand string

	// TargetURL is the URL of the target application.
	TargetURL *url.URL

	// Change listens for changes on the filesystem to signal when a build
	// should be triggered.
	Change <-chan string

	// initOnce guards initialization of the Proxy struct from happening
	// more than once.
	initOnce sync.Once

	// mu protects the following attributes.
	mu           sync.RWMutex
	reverseProxy *httputil.ReverseProxy // The reverse proxy handler.
	cmd          *exec.Cmd              // The currently running command.
	err          error                  // The last error encountered.
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.init()

	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.err != nil {
		http.Error(w, p.err.Error(), http.StatusInternalServerError)
		return
	}

	// The target may not become immediately available after it is started.
	// The request will be attempted a few times to see if we can get a successful
	// response.
	bw := badGatewayWriter{
		ResponseWriter: w,
		MaxRetries:     3,
	}
	for !bw.OK() {
		p.reverseProxy.ServeHTTP(&bw, r)
		time.Sleep(100 * time.Millisecond)
	}
}

// init is called upon the first access to ServeHTTP to setup the environment
// required for the proxy to operate.
func (p *Proxy) init() {
	p.initOnce.Do(func() {
		p.reverseProxy = httputil.NewSingleHostReverseProxy(p.TargetURL)
		p.compile()
		go func() {
			// Reloads are debounced to prevent successive saves from triggering many builds to
			// kick off unnecessarily. This is especially useful when using formatting tools
			// (gofmt, goimports, etc.) after saving a file as they will save the file again after
			// formatting and it is not useful to run a build each time.
			debouncer := debounce.New(100 * time.Millisecond)
			for {
				select {
				case <-p.Change:
					debouncer(p.compile)
				case <-p.Context.Done():
					return
				}
			}
		}()
	})
}

// compile builds and runs the target application. If the target application is running
// when a new compile is triggered it will be killed.
func (p *Proxy) compile() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := build(p.Context, p.BuildCommand); err != nil {
		p.err = fmt.Errorf("build: %w", err)
		return
	}
	if p.cmd != nil {
		if err := p.cmd.Process.Kill(); err != nil {
			p.err = err
			return
		}
	}
	c, err := run(p.Context, p.RunCommand)
	if err != nil {
		p.err = fmt.Errorf("run: %w", err)
		return
	}
	p.cmd = c
	p.err = nil
}

// badGatewayWriter wraps a ResponseWriter to allow handling of bad gateway errors from the
// upstream proxy HTTP handler.
type badGatewayWriter struct {
	http.ResponseWriter

	// MaxRetries holds the number of times the bad gateway writer should hold back the response.
	// After MaxRetries number of failures the bad gateway writer will revert to sending the content
	// back to the client unmodified. This allows the client to see a response when we have given up.
	MaxRetries int

	failed  int
	written bool
}

// WriteHeader watches for a bad gateway error and sets the appropriate internal state and eats
// the header write. Any other status will be sent as normal.
func (b *badGatewayWriter) WriteHeader(status int) {
	if b.failed < b.MaxRetries && status == http.StatusBadGateway {
		b.failed++
		return
	}
	b.ResponseWriter.WriteHeader(status)
	b.written = true
}

// Write takes the content to be written and eats it if a bad gateway error has been detected.
// This pevents sending unwanted data to the client when we want to retry. Under normal conditions
// the data is passed up to the ResponseWriter.
func (b *badGatewayWriter) Write(p []byte) (int, error) {
	if !b.OK() {
		return 0, nil
	}
	return b.ResponseWriter.Write(p)
}

// OK returns true if the reponse was anything other than a bad gateway error or the number of retries
// exceeded the limit.
func (b *badGatewayWriter) OK() bool {
	return b.written
}
