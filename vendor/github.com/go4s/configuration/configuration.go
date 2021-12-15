package configuration

import (
    "bufio"
    "os"
    "strings"
    "sync"
)

const (
    Prefix          = "SERVICE"
    Sep             = "="
    ExtendedEnvFile = "service.extended.config"
)

type Configuration = map[string]interface{}

type Modifier func(Configuration) error

var (
    mutable Configuration
    once    = new(sync.Once)
)

func fromEnv() {
    mutable = Configuration{}
    for _, eq := range os.Environ() {
        tmp := strings.SplitN(eq, Sep, 2)
        if len(tmp) == 2 {
            apply(tmp)
        }
    }
    if path, found := mutable[ExtendedEnvFile]; found {
        loadFromFile(path.(string))
    }
}

func apply(tmp []string) {
    if strings.HasPrefix(tmp[0], Prefix) {
        mutable[strings.ReplaceAll(strings.ToLower(tmp[0]), "_", ".")] = tmp[1]
    } else {
        mutable[tmp[0]] = tmp[1]
    }
}

func loadFromFile(s string) {
    f, err := os.Open(s)
    if err != nil {
        return
    }
    defer f.Close()
    lines := bufio.NewScanner(f)
    lines.Split(bufio.ScanLines)
    for lines.Scan() {
        tmp := strings.SplitN(lines.Text(), Sep, 2)
        if len(tmp) == 2 {
            apply(tmp)
        }
    }
}

func init() {
    once.Do(func() {
        fromEnv()
    })
}

func FromEnv() Configuration {
    return mutable
}
