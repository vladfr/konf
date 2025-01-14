# The simplest config loader for Go

![Go Version](https://img.shields.io/github/go-mod/go-version/nil-go/konf)
[![Go Reference](https://pkg.go.dev/badge/github.com/nil-go/konf.svg)](https://pkg.go.dev/github.com/nil-go/konf)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/nil-go/konf)](https://goreportcard.com/report/github.com/nil-go/konf)
[![Build](https://github.com/nil-go/konf/actions/workflows/test.yml/badge.svg)](https://github.com/nil-go/konf/actions/workflows/test.yml)
[![Coverage](https://codecov.io/gh/nil-go/konf/branch/main/graph/badge.svg)](https://codecov.io/gh/nil-go/konf)

konf offers an(other) opinion on how Go programs can read configuration without
becoming coupled to a particular configuration source.

## Features

- [konf.Unmarshal](#usage) for reading configuration to any type of object.
- [konf.OnChange](#usage) for registering callbacks while configuration changes.
- [konf.Explain](#understand-the-configuration) for understanding where the configuration is loaded from.
- [Various providers](#configuration-providers) for loading configuration from major clouds,
  [AWS](examples/aws), [Azure](examples/azure), and [GCP](examples/gcp) with [Notifier](notifier) for notifying the changes of configuration.
- [Zero dependencies](go.mod) in core module which supports loading configuration
  from environment variables,flags, and embed file system.

## Benchmarks

The following [benchmarks](benchmark)  compare the performance of konf
with [spf13/viper](https://github.com/spf13/viper) and
[knadh/koanf](https://github.com/knadh/koanf), which are inspiration of konf.

|       | Unmarshal (ns/op) | Unmarshal (allocs/op) | Get[^1] (ns/op) | Get (allocs/op) |
|:------|------------------:|----------------------:|----------------:|----------------:|
| Konf  |         __41.09__ |                 __4__ |           16.71 |           __1__ |
| Viper |             614.8 |                    22 |           104.9 |               3 |
| Koanf |             15949 |                   657 |       __7.898__ |           __1__ |

[^1]: Comparing to Get in both Viper and Koanf only can get the primitive types, Get in Konf can get any type
since it's just a wrapper of Unmarshal. So for complex struct, it only needs single Konf Get
but may needs multiple Get(s) in Viper and Koanf.

## Usage

Somewhere, early in an application's life, it will make a decision about which
configuration source(s) (implementation) it actually wants to use. Something like:

```go
//go:embed config
var config embed.FS

func main() {
    var config konf.Config

    // Load configuration from embed file system.
    if err := config.Load(fs.New(config, "config/config.json")); err != nil {
        // Handle error here.
    }
    // Load configuration from environment variables.
    if err := config.Load(env.New(env.WithPrefix("server"))); err != nil {
        // Handle error here.
    }

    // Watch the changes of configuration.
    go func() {
      if err := config.Watch(ctx); err != nil {
        // Handle error here.
      }
    }

    konf.SetDefault(config)

    // ... other setup code ...
}
```

Outside of this early setup, no other packages need to know about the choice of
configuration source(s). They read configuration in terms of functions in package `konf`:

```go
func (app *appObject) Run() {
    // Server configuration with default values.
    serverConfig := struct {
        Host string
        Port int
    }{
        Host: "localhost",
        Port: "8080",
    }
    // Read the server configuration.
    if err := konf.Unmarshal("server", &serverConfig);  err != nil {
        // Handle error here.
    }

    // Register callbacks while server configuration changes.
    konf.OnChange(func() {
      // Reconfig the application object.
    }, "server")

    // ... use cfg in app code ...
}
```

## Design

It contains two APIs with two different sets of users:

- The `Config` type is intended for application authors. It provides a relatively
  small API which can be used everywhere you want to read configuration.
  It defers the actual configuration loading to the `Loader` interface.
- The `Loader` and `Watcher` interface is intended for configuration source library implementers.
  They are pure interfaces which can be implemented to provide the actual configuration.

This decoupling allows application developers to write code in terms of `*konf.Config`
while the configuration source(s) is managed "up stack" (e.g. in or near `main()`).
Application developers can then switch configuration sources(s) as necessary.

## Change Notification
The providers for loading configuration from clouds periodically poll the configuration source.
It also supports watching the changes of configuration using corresponding notifier.
For example, the `sns` notifier notifies the changes of `appconfig`  and `s3` provider:

```go
notifier := sns.NewNotifier("konf-test")
notifier.Register(s3Loader, appConfigLoader)
go func() {
  if err := notifier.Start(ctx); err != nil {
    // handle error
  }
}()
```

## Understand the configuration

While the configuration is loaded from multiple sources, static like environments or dynamic like AWS AppConfig,
it's hard to understand where a final value comes from. The `Config.Explain` method provides information
about how Config resolve each value from loaders for the given path. One example explanation is like:

```
config.nest has value [map] is loaded by map.
Here are other value(loader)s:
  - env(env)
```

Even more, the `Config.Explain` blurs sensitive information (e.g. password, secret, api keys).

## Observability

For watching the changes of configuration, it uses `slog.Default()` for logging. You can change the logger
via option `konf.WithLogHandler`. Furthermore, you also can register onStatus via option `konf.WithOnStatus`
to monitor the status of configuration loading/watching, e.g. recording metrics.

## Configuration Providers

There are providers for the following configuration sources.

| Loader                                    | Load From                                                                               | Watch Changes | Notifier                  |
|:------------------------------------------|:----------------------------------------------------------------------------------------|:-------------:|:--------------------------|
| [`env`](provider/env)                     | environment variables                                                                   |               |                           |
| [`fs`](provider/fs)                       | [fs.FS](https://pkg.go.dev/io/fs)                                                       |               |                           |
| [`file`](provider/file)                   | file                                                                                    |       ✓       |                           |
| [`flag`](provider/flag)                   | [flag](https://pkg.go.dev/flag)                                                         |               |                           |
| [`pflag`](provider/pflag)                 | [spf13/pflag](https://github.com/spf13/pflag)                                           |               |                           |
| [`appconfig`](provider/appconfig)         | [AWS AppConfig](https://aws.amazon.com/systems-manager/features/appconfig/)             |       ✓       | [sns](notifier/sns)       |
| [`s3`](provider/s3)                       | [AWS S3](https://aws.amazon.com/s3)                                                     |       ✓       | [sns](notifier/sns)       |
| [`parameterstore`](provider/parameterstore)           | [AWS ParameterStore](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)                                         |       ✓       | [sns](notifier/sns)       |
| [`azappconfig`](provider/azappconfig)     | [Azure App Configuration](https://azure.microsoft.com/en-us/products/app-configuration) |       ✓       | [azservicebus](notifier/azservicebus) |
| [`azblob`](provider/azblob)               | [Azure Blob Storage](https://azure.microsoft.com/en-us/products/storage/blobs)          |       ✓       | [azservicebus](notifier/azservicebus) |
| [`secretmanager`](provider/secretmanager) | [GCP Secret Manager](https://cloud.google.com/security/products/secret-manager)         |       ✓       | [pubsub](notifier/pubsub) |
| [`gcs`](provider/gcs)                     | [GCP Cloud Storage](https://cloud.google.com/storage)                                   |       ✓       | [pubsub](notifier/pubsub) |

## Custom Configuration Providers

You can Custom provider by implementing the `Loader` for static configuration loader (e.g [`fs`](provider/fs))
or both `Loader` and `Watcher` for dynamic configuration loader (e.g. [`appconfig`](provider/appconfig)).
