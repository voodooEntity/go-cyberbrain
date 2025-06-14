# Developing Custom Actions for `go-cyberbrain`

`go-cyberbrain`'s flexibility comes from its **pluggable `Actions`**. These are custom Go modules that define what the system can actually *do*. By implementing the `ActionInterface`, you can extend `go-cyberbrain` to perform specific tasks, interact with external services, or generate new knowledge.

---

## The `ActionInterface`

Every `Action` must implement the `interfaces.ActionInterface`, which looks like this:

```go
type ActionInterface interface {
    Execute(transport.TransportEntity, string, string) ([]transport.TransportEntity, error)
    GetConfig() transport.TransportEntity
}
````

While `GetConfig()` is crucial for registering your action (and is covered in the [Concepts Documentation](https://www.google.com/search?q=docs/CONCEPTS.md)), we'll focus here on the **`Execute` method**, which contains your action's core logic.

-----

## Implementing `Execute(input transport.TransportEntity, requirement string, context string)`

The `Execute` method is where your `Action` performs its work. A `Neuron` calls it when a `Job` is assigned.

* **`input transport.TransportEntity`**: This is the primary data payload for your action. It's the `TransportEntity` (or a combination of related entities, depending on how the `Scheduler` and `Demultiplexer` processed the trigger) that matched your action's dependency. Its `Type`, `Value`, `Properties`, and `ChildRelations` contain the specific knowledge your action needs to operate on.
* **`requirement string`**: This string indicates the specific dependency within your action's configuration that was met to trigger this job. It can be useful for actions with multiple dependencies to differentiate their behavior.
* **`context string`**: This provides additional contextual information, often derived from the `Context` property of the `input` entity itself, or the `Job` that triggered the action.

Your `Execute` method must return:

* **`[]transport.TransportEntity`**: A slice of `TransportEntity` objects. These represent any **new or modified knowledge** your action generates. After your `Action` completes, the `Neuron` will pass these results to the `Mapper`, which will integrate them back into `gits`. This is how your action expands `go-cyberbrain`'s knowledge graph and potentially triggers subsequent actions.
* **`error`**: If your action encounters an error, return it. The `Neuron` will mark the `Job` as "Error" in `gits`.

### Example Walkthrough: `resolveIPFromDomain`

Let's look at a simplified version of the `example.Example` action to illustrate the `Execute` method:

```go
package example

import (
    "github.com/voodooEntity/gits/src/transport"
    "net"
    "strconv"
    "time"
    // ... other imports for gits, archivist, etc.
)

type Example struct {
    // These are injected by the Neuron if your action implements the
    // respective interfaces (ActionExtendGitsInterface, etc.)
    Gits   *gits.Gits
    Mapper *cerebrum.Mapper
    log    *archivist.Archivist // Assuming you add a logger field and SetLogger method
}

// ... New() and Set* methods omitted for brevity

func (self *Example) Execute(input transport.TransportEntity, requirement string, context string) ([]transport.TransportEntity, error) {
    // 1. Access the input data
    // For the resolveIPFromDomain action, input.Value is expected to be a domain name
    domain := input.Value
    self.log.Debug("Executing resolveIPFromDomain for domain: ", domain)

    // 2. Perform the action's core logic (e.g., external API call, computation)
    ips, err := net.LookupIP(domain)
    if err != nil {
        self.log.Error("Failed to lookup IP for ", domain, ": ", err)
        return []transport.TransportEntity{}, err // Return empty slice and the error
    }

    // 3. Prepare new knowledge based on the action's results
    // We'll modify the original input entity by adding child relations
    // representing the resolved IPs.
    var ipChildren []transport.TransportRelation

    for _, ip := range ips {
        if ipv4 := ip.To4(); ipv4 != nil { // Only process IPv4 for this example
            ipChildren = append(ipChildren, transport.TransportRelation{
                Target: transport.TransportEntity{
                    // ID: -2 signals the Mapper to find or create this entity
                    // based on Type, Value, and its relation to the parent.
                    ID:         -2,
                    Type:       "IP",
                    Value:      ipv4.String(),
                    Context:    context, // Inherit context
                    Properties: map[string]string{"protocol": "V4", "created": strconv.FormatInt(time.Now().Unix(), 10)},
                },
            })
            self.log.Debug("Resolved IP: ", ipv4.String())
        }
    }

    // 4. Return the new/modified knowledge
    // We append the new IP relations to the original input entity.
    // When this entity is returned, the Mapper will persist these new relations
    // and the associated IP entities into gits.
    input.ChildRelations = append(input.ChildRelations, ipChildren...)

    // Note: The example code had input.Properties = make(map[string]string) here.
    // For many actions, you might want to preserve or modify original properties,
    // not overwrite them. This depends on your specific action's requirements.

    return []transport.TransportEntity{input}, nil // Return the enriched input entity
}
```

### Key Concepts in `Execute`

* **Input Consumption:** Your action receives `input` as a `transport.TransportEntity`. You extract the necessary data from its `Value`, `Properties`, or `ChildRelations`.
* **Result Generation:** Your action's primary goal is to generate new data or modify existing data. This is typically done by:
    * Creating entirely new `transport.TransportEntity` instances.
    * Adding `ChildRelations` or `ParentRelations` to existing entities (like adding `IP` children to a `Domain` entity).
    * Modifying `Properties` of existing entities.
* **Mapping to `Gits`:** The `Neuron` will automatically pass the `[]transport.TransportEntity` you return to the `Mapper`. The `Mapper` then handles the complex task of persisting this data into `gits`, creating new entities or updating existing ones based on their `ID`s and relationships.
* **Special `ID` Values:** Pay attention to `ID` values like `-1` (create new entity, assign new ID) and `-2` (find or create based on Type, Value, and Parent relation). These are crucial for the `Mapper` to correctly integrate your new knowledge into the graph.
* **Error Handling:** Proper error handling within your `Execute` method is vital. Returning an `error` signals to `go-cyberbrain` that the `Job` failed.

-----

## Dependency Injection for Actions

To allow your actions to interact with core system components, `go-cyberbrain` uses **dependency injection**. If your `Action` struct implements specific interfaces, the `Neuron` will automatically provide the necessary instances:

* **`type ActionExtendGitsInterface interface { SetGits(gitsInstance *gits.Gits) }`**: For accessing the `gits` database directly.
* **`type ActionExtendMapperInterface interface { SetMapper(mapper *Mapper) }`**: For direct access to the `Mapper` (less common if `Execute` returns entities for automatic mapping).
* **`type ActionExtendLoggerInterface interface { SetLogger(log *Archivist) }`**: For logging within your action.

By implementing these `Set` methods on your `Action` struct, you ensure your action has the tools it needs.

-----

## Registering Your Action

To make your custom `Action` available to `go-cyberbrain`, you must register it with the `Cortex`. This involves providing an instance of your action to the system during initialization. It's a common Go practice for packages that provide an interface implementation to also expose a **`New` method**, which acts as an instance factory. This `New` method should return a pointer to your `Action` struct.

For example, if your action is defined in the `example` package and its type is `Example`, you would typically have a `NewExample()` function:

```go
package example

// ... imports and Example struct definition ...

func NewExample() *Example {
    return &Example{}
}
```

This `NewExample()` function then gets provided to the `Cyberbrain` during its setup, typically when you initialize the `Cortex` and register actions:

```go
// In your main setup code (e.g., cmd/example/main.go)
import (
    "github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
    "github.com/voodooEntity/go-cyberbrain/src/example" // Import your action package
)

func main() {
    // ... setup gits, archivist ...

    cortex := cerebrum.NewCortex(gitsInstance, archivistInstance)
    // Register your action using its New method
    cortex.RegisterAction(example.NewExample())

    // ... proceed with Cyberbrain initialization and start ...
}
```

This factory method (`NewExample`) ensures that the `Cortex` can reliably obtain a new, correctly initialized instance of your action when it's needed, allowing `go-cyberbrain` to manage and use your custom capabilities.

