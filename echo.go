package echo

import (
    "bytes"
    "context"
    "fmt"
    "github.com/go4s/configuration"
    "log"
    "net"
)

type (
    Service interface {
        HandlePkt(context.Context) func() error
        Close() error
    }
)

type service struct {
    conn     *net.UDPConn
    hostName []byte
}

func (s *service) Close() error { return s.conn.Close() }

func (s *service) HandlePkt(ctx context.Context) func() error {
    return func() (err error) {
        var (
            size  int
            rAddr *net.UDPAddr
            buf   = [(1 << 31) - 1]byte{0}
        )
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                if size, rAddr, err = s.conn.ReadFromUDP(buf[:]); err != nil {
                    return err
                }
                if wSize, err := s.conn.WriteToUDP(
                    bytes.Join(
                        [][]byte{s.hostName, buf[:size]},
                        []byte(" : "),
                    ),
                    rAddr,
                ); err != nil {
                    log.Printf("write to %s failed (size : %d) : %s", rAddr, wSize, err)
                }
            }
        }
    }
}

func NewService(env configuration.Configuration) Service {
    var (
        s   = service{}
        err error
    )
    if s.hostName, err = getHostName(env); err != nil {
        panic(err)
    }
    if s.conn, err = dial(env); err != nil {
        panic(err)
    }
    return &s
}

func getHostName(env configuration.Configuration) ([]byte, error) {
    if hostName, found := env["HOSTNAME"]; found {
        return []byte(hostName.(string)), nil
    }
    return []byte{}, fmt.Errorf("err HostName not found")
}

func dial(env configuration.Configuration) (conn *net.UDPConn, err error) {
    var port = ":"
    var lAddr *net.UDPAddr

    if portStr, found := env["PORT"]; found {
        port += portStr.(string)
    } else {
        port += "80"
    }
    if lAddr, err = net.ResolveUDPAddr("udp", port); err != nil {
        log.Printf("resolve error : %v\n", err)
        return
    }
    if conn, err = net.ListenUDP("udp", lAddr); err != nil {
        log.Printf("listen error : %v\n", err)
    }
    return
}
