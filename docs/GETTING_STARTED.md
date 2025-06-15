# Starting Your First `go-cyberbrain` Project

This guide explains the foundational steps to set up and run a `go-cyberbrain` instance, using the `cmd/example/example.go` file as a blueprint. This covers system initialization, configuration, action registration, and data ingestion.

---

## The Basic Project Structure

A typical `go-cyberbrain` project, like the example, will often have a structure similar to this:

```

my-cyberbrain-project/
├── cmd/
│   └── myproject/        \# Your main application entry point (e.g., cmd/example/example.go)
│       └── main.go
├── src/
│   └── myactions/        \# Your custom Action implementations
│       └── myaction.go
├── go.mod
├── go.sum
└── ...

````

Your primary interaction will be within the `main.go` file of your `cmd` directory.

---

## Step-by-Step Initialization (`cmd/example/example.go` Breakdown)

Let's examine the key parts of `cmd/example/example.go` to understand how to set up `go-cyberbrain`.

### 1. Setting Up the Logger

```go
package main

import (
    "github.com/voodooEntity/gits/src/storage"
    "os"

    "github.com/voodooEntity/gits/src/transport"
    "github.com/voodooEntity/go-cyberbrain"
    "github.com/voodooEntity/go-cyberbrain/src/example"
    "github.com/voodooEntity/go-cyberbrain/src/system/archivist"
    "github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
    "log" // Standard Go log package
)

func main() {
    // A standard Go logger is used for internal system messages.
    logger := log.New(os.Stdout, "", 0) // Logs to standard output
    // logger := log.New(io.Discard, "", 0) // Uncomment to disable logging
````

`go-cyberbrain` integrates with the standard Go `log` package. You provide your own `log.Logger` instance, allowing you to control where logs go (e.g., `os.Stdout`, a file, or `io.Discard` to mute them) and how they are formatted. Cyberbrain does a certain level of formatting of its log strings itself so its recommended to use 0 as logflag if not intended different.

### 2\. Creating the `Cyberbrain` Instance

The core of your project starts by creating a `cyberbrain.Cyberbrain` instance:

```go
    // create base instance. ident is required.
    // NeuronAmount will default back to runtime.NumCPU == num logical cpu's
    cb := cyberbrain.New(cyberbrain.Settings{
       NeuronAmount: 1,
       Ident:        "GreatName",
       LogLevel:     archivist.LEVEL_INFO,
       Logger:       logger,
    })
```

The `cyberbrain.New()` function takes a `cyberbrain.Settings` struct to configure the system.

#### `cyberbrain.Settings` Configuration Options

* **`NeuronAmount int`**:

    * Specifies the number of `Neuron` (worker) goroutines `go-cyberbrain` should create and run concurrently.
    * If set to `0`, it defaults to `runtime.NumCPU()`, which is the number of logical CPUs available on the system.
    * **Example:** `NeuronAmount: 1` (for simple testing), `NeuronAmount: 0` (for max concurrency).

* **`Ident string`**:

    * A **required** unique identifier for this `Cyberbrain` instance. This `Ident` is used internally in `gits` to tag entities related to this specific `Cyberbrain`'s operation (e.g., its own `Neuron` entities, `Job` entities).
    * **Example:** `Ident: "MyFirstBrain"`, `Ident: "WebCrawlerInstance"`

* **`LogLevel archivist.LogLevel`**:

    * Controls the verbosity of `go-cyberbrain`'s internal logging.
    * Uses constants from the `archivist` package:
        * `archivist.LEVEL_DEBUG`: Most verbose, includes detailed internal process messages.
        * `archivist.LEVEL_INFO`: General informational messages, good for understanding flow.
        * `archivist.LEVEL_WARN`: Potential issues or non-critical problems.
        * `archivist.LEVEL_ERROR`: Critical errors that prevent normal operation.
        * `archivist.LEVEL_FATAL`: System-critical errors leading to shutdown.
        * `archivist.LEVEL_NONE`: Disables all logging (if `Logger` is not `io.Discard`).
    * **Example:** `LogLevel: archivist.LEVEL_INFO`

* **`Logger *log.Logger`**:

    * The standard Go `log.Logger` instance that `go-cyberbrain` will use for its output.
    * **Example:** `Logger: logger` (referencing the logger initialized earlier).

### 3\. Registering Your Actions

Before starting the system, you must register the `Actions` that your `Cyberbrain` instance will be able to execute:

```go
    // register actions
    cb.RegisterAction("resolveIPFromDomain", example.New)
```

* `cb.RegisterAction(actionName string, actionFactoryFunc func() interfaces.ActionInterface)`:
    * The `actionName` (e.g., `"resolveIPFromDomain"`) is a logical string name that identifies this action. It must match the `Value` property of the `Action` entity in its `GetConfig()`.
    * The `actionFactoryFunc` is a function that returns a new instance of your action, typically a `New()` method from your action's package (e.g., `example.New` for `example.Example`).

For a detailed explanation on how to **implement** a custom `Action` and its `GetConfig()` method, refer to the [Developing Custom Actions Documentation](https://www.google.com/search?q=docs/ACTIONS.md).

### 4\. Starting the `Cyberbrain`

Once configured and actions are registered, start the `Cyberbrain`'s internal processes:

```go
    // start the neurons
    cb.Start()
```

This call initiates the `Neuron` goroutines and other background services. The `Cyberbrain` is now active and ready to process information.

### 5\. Learning Initial Data and Scheduling Jobs

The `go-cyberbrain` system is reactive. It starts working when you feed it initial data:

```go
    // Learn data and schedule based on it
    cb.LearnAndSchedule(transport.TransportEntity{
       ID:         storage.MAP_FORCE_CREATE,
       Type:       "Domain",
       Value:      "laughingman.dev",
       Context:    "example code",
       Properties: map[string]string{},
    })
```

* `cb.LearnAndSchedule(entity transport.TransportEntity)`:
    * This is how you inject initial knowledge into the `Cyberbrain`. The provided `transport.TransportEntity` will be mapped into `gits` by the `Mapper`.
    * The `Mapper` will mark this data as "new" (using the `bMap` flag).
    * The `Scheduler` will then analyze this newly added data. If it finds any registered `Actions` whose dependencies match the structure of the learned `entity`, it will create `Job` entities for the `Neurons` to pick up.
    * **`ID: storage.MAP_FORCE_CREATE`**: This special constant instructs the `Mapper` to always create a new entity in `gits` for this data, even if an entity with the same `Type` and `Value` might already exist. This is useful for initial seeding or when you explicitly want duplicates. Other ID values like `-1` (create new, assign ID) or `-2` (find/create by type/value/parent) can also be used, depending on the mapping desired.

### 6\. Observing System Activity and Managing Shutdown

For applications that need to run until all work is done (e.g., batch processing, one-shot analysis), the `Observer` is used:

```go
    // get an observer instance. provide a callback
    // to be executed at the end and lethal=true
    // which stops the cyberbrain at the end
    obsi := cb.GetObserverInstance(func(mi *cerebrum.Memory) {
       qry := mi.Gits.Query().New().Read("IP")
       ret := mi.Gits.Query().Execute(qry)
       logger.Println("Result:", ret)
    }, true)

    // blocking while neurons are
    // working & non-finished jobs exist
    obsi.Loop()
```

* `cb.GetObserverInstance(callback func(*cerebrum.Memory), lethal bool)`:

    * Retrieves an `Observer` instance that monitors the system.
    * `callback func(*cerebrum.Memory)`: A function that will be executed once the `Observer` detects that the `Cyberbrain` has entered an "endgame" (idle) state. This callback provides access to the `Memory` (which contains `Gits`) for final result queries or cleanup.
    * `lethal bool`: If `true`, the `Observer` will initiate a full shutdown of the `Cyberbrain` system after the `endgame` callback has completed. If `false`, the `Cyberbrain` will remain running in an idle state.

* `obsi.Loop()`:

    * This is a **blocking call** that keeps your `main` goroutine alive. The `Observer` continuously monitors the state of `Jobs` and `Neurons` in `gits`.
    * It will only return (and allow `main` to exit) when the `Cyberbrain` is determined to be inactive (no open jobs, all neurons searching for work, and no new activity detected for a period).

By following these steps, you can set up and control your `go-cyberbrain` instance for various knowledge processing and automation tasks.

-----
