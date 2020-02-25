// recharge is a reloading tool for Go HTTP services for use during
// development. It watches a project for changes and rebuilds the
// application when the code is modified. It provides an HTTP proxy
// to fascilitate in allowing the client to seemlessly access the
// target application.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/tebeka/atexit"
)

// Config holds the configuration values loaded from the configuration file.
type Config struct {
	// The RootDir specifies the directory, and its children, that should
	// be watched. If no value if given the root directory will default
	// to the current directory.
	RootDir string

	// Build is the command used to build the target application. This
	// command is run at startup and after each time a change is detected.
	Build string

	// Run is the command used to run the target application after building.
	// This command is executed after each successful build.
	Run string

	// The SourceAddr defines the listen address of the proxy HTTP server.
	// If this value is empty it defaults to :3000.
	SourceAddr string

	// The TargetAddr defines the address of the target HTTP application.
	TargetAddr string
}

func main() {
	// Load the application configuration and provide default
	// configuration values where necessary.
	var conf Config
	_, err := toml.DecodeFile("recharge.conf", &conf)
	if err != nil {
		log.Fatalf("unable to read configuration: %v", err)
	}
	if conf.RootDir == "" {
		conf.RootDir = "."
	}
	if conf.SourceAddr == "" {
		conf.SourceAddr = ":3000"
	}
	if conf.TargetAddr == "" {
		conf.TargetAddr = "http://localhost:3001"
	}

	// When the application exits the context is cancelled to provide
	// an oppurunity for downstream functions to perform any necessary
	// cleanup operations.
	ctx, cancel := context.WithCancel(context.Background())
	atexit.Register(cancel)

	// For consistency the target URL may be provided in the same format
	// as the listen address, without a protocol scheme. When that is the
	// case the address is assumed to be an HTTP address.
retry:
	target, err := url.Parse(conf.TargetAddr)
	if err != nil {
		if strings.Contains(err.Error(), "missing protocol scheme") {
			conf.TargetAddr = "http://" + conf.TargetAddr
			goto retry
		}
		log.Fatalf("failed to parse target address: %v", err)
	}

	ch, err := watch(ctx, conf.RootDir)
	if err != nil {
		log.Fatalf("unable to watch project: %v", err)
	}
	server := http.Server{
		Addr: conf.SourceAddr,
		Handler: &Proxy{
			Context:      ctx,
			BuildCommand: conf.Build,
			RunCommand:   conf.Run,
			TargetURL:    target,
			Change:       ch,
		},
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("http error: %v", err)
	}
}
