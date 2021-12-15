package main

import (
    "context"
    "fmt"
    "github.com/go4s/configuration"
    "golang.org/x/sync/errgroup"
    "io"
    "log"
    "net"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
)

var (
    initializer = sync.Once{}
    env         configuration.Configuration
    workers     *errgroup.Group
    rootCtx     context.Context
    cancel      context.CancelFunc
)

func init() {
    initializer.Do(initialize)
}

func initialize() {
    env = configuration.FromEnv()
    rootCtx, cancel = context.WithCancel(context.Background())
    workers, rootCtx = errgroup.WithContext(rootCtx)
}

func main() {
    defer cancel()
    conn, err := net.Dial("udp", "192.168.99.100:30591")
    if err != nil {
        panic(err)
    }
    workers.Go(handleSend(rootCtx, conn))
    workers.Go(handleRecv(rootCtx, conn))
    workers.Go(handleSignal(rootCtx, conn))
    log.Printf("final recv : %v\n", workers.Wait())
}

func handleSend(ctx context.Context, conn net.Conn) func() error {
    return func() error {
        ticker := time.NewTimer(time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case then := <-ticker.C:
                if _, err := conn.Write([]byte(then.Format(time.RFC3339))); err != nil {
                    log.Printf("write err : %v\n", err)
                    return err
                }
                ticker.Reset(time.Second)
            }
        }
    }
}

func handleRecv(ctx context.Context, conn net.Conn) func() error {
    return func() error {
        var buf = [1000]byte{0}
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                if size, err := conn.Read(buf[:]); err != nil {
                    log.Printf("recv err : %v\n", err)
                    return err
                } else {
                    fmt.Printf("recv msg : %s\n", string(buf[:size]))
                }
            }
        }
    }
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
