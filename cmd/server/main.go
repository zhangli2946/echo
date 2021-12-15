package main

import (
    "context"
    "fmt"
    "github.com/go4s/configuration"
    "github.com/zhangli2946/echo"
    "golang.org/x/sync/errgroup"
    "io"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
)

var (
    initializer = sync.Once{}
    env         configuration.Configuration
    service     echo.Service

    workers *errgroup.Group
    rootCtx context.Context
    cancel  context.CancelFunc
)

func init() {
    initializer.Do(initialize)
}

func initialize() {
    env = configuration.FromEnv()
    service = echo.NewService(env)
    rootCtx, cancel = context.WithCancel(context.Background())
    workers, rootCtx = errgroup.WithContext(rootCtx)
}

func main() {
    defer cancel()
    workers.Go(handleSignal(rootCtx, service))
    workers.Go(service.HandlePkt(rootCtx))
    log.Printf("final recv : %v\n", workers.Wait())
}

func handleSignal(ctx context.Context, closer io.Closer) func() error {
    return func() error {
        defer closer.Close()
        signals := make(chan os.Signal, 1)
        defer close(signals)
        signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
        defer signal.Stop(signals)
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case sig := <-signals:
                return fmt.Errorf("signal recv : %v", sig)
            }
        }
    }
}
