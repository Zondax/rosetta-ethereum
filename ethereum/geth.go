// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ethereum

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

const (
	ethLogger       = "erigon"
	ethStdErrLogger = "erigon err"
	rpcLogger       = "rpcDaemon"
	rpcStdErrLogger = "rpcDaemon err"
)

// logPipe prints out logs from geth. We don't end when context
// is canceled beacause there are often logs printed after this.
func logPipe(pipe io.ReadCloser, identifier string) error {
	reader := bufio.NewReader(pipe)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Println("closing", identifier, err)
			return err
		}

		message := strings.ReplaceAll(str, "\n", "")
		log.Println(identifier, message)
	}
}

// StartNode starts a geth daemon in another goroutine
// and logs the results to the console.
func StartNode(ctx context.Context, arguments string, g *errgroup.Group) error {
	//parsedArgs := strings.Split(arguments, " ")
	cmd := exec.Command(
		"/app/erigon", "datadir /data", "chain goerli",
		) // #nosec G204

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	g.Go(func() error {
		return logPipe(stdout, ethLogger)
	})

	g.Go(func() error {
		return logPipe(stderr, ethStdErrLogger)
	})

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: unable to start geth", err)
	}

	g.Go(func() error {
		<-ctx.Done()

		log.Println("sending interrupt to erigon")
		return cmd.Process.Signal(os.Interrupt)
	})

	return cmd.Wait()
}

// StartRPCDaemon starts an Erigon rpc daemon in another goroutine
// and logs the results to the console.
func StartRPCDaemon(ctx context.Context, arguments string, g *errgroup.Group) error {
	//parsedArgs := strings.Split(arguments, " ")
	cmd := exec.Command(
		"/app/rpcdaemon", "datadir /data", "private.api.addr=localhost:9090",
		"http.api=eth,erigon,web3,net,debug,trace,txpool",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	g.Go(func() error {
		return logPipe(stdout, rpcLogger)
	})

	g.Go(func() error {
		return logPipe(stderr, rpcStdErrLogger)
	})

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: unable to start rpc daemon", err)
	}

	g.Go(func() error {
		<-ctx.Done()

		log.Println("sending interrupt to geth")
		return cmd.Process.Signal(os.Interrupt)
	})

	return cmd.Wait()
}