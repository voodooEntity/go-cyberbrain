# Core Concepts & Architecture of `go-cyberbrain`

`go-cyberbrain` is a Go-based system designed to manage and process knowledge in a graph structure, enabling autonomous and reactive behavior. It aims to continuously learn, reason, and act based on changes in its internal knowledge base.

---

## The System Metaphor: A Distributed Brain

`go-cyberbrain` adopts a "brain" metaphor to describe its architecture:

* **Memory (`Gits`):** The persistent, central store of all knowledge and system state.
* **Cortex (`Cortex`):** Registers the system's capabilities (Actions) and their dependencies.
* **Neurons (`Neuron`):** The distributed workers that perform tasks.
* **Scheduler (`Scheduler`):** Identifies patterns in new information and triggers subsequent tasks.
* **Demultiplexer (`Demultiplexer`):** Processes new data into all relevant perspectives for the scheduler.
* **Observer (`Observer`):** Monitors the system's overall activity and manages its lifecycle.

---

## Core Components Deep Dive

### `Gits` (Graph-based In-memory Transactional Storage)

**`Gits`** serves as the **central memory and knowledge graph** for `go-cyberbrain`. All information, including learned data, action configurations, and the system's operational state (like Job status and Neuron states), is stored as interconnected **`TransportEntity`** objects.

* **In-Memory Focus:** `gits` is designed as an in-memory hashtable, prioritizing high-speed data access and manipulation.
* **Data Representation:** Data is stored as entities and relations, forming a graph.
* **`Unsafe` Methods:** `gits` exposes specific "Unsafe" methods for direct, high-performance interaction with its internal storage mechanisms. These methods are used by `go-cyberbrain`'s `Mapper` to ensure efficient data writes, with locking handled explicitly by `go-cyberbrain` at a higher level.

### `Cortex` (Action Registry)

The `Cortex` is responsible for **registering and managing the `Actions`** that `go-cyberbrain` can perform.

* It translates the `Action`'s `GetConfig()` output into a graph structure within `gits`. This includes defining the `Action` entity itself, its `Dependencies` (what data it needs to run), and its `Categories`.
* This mapping allows the `Scheduler` to query `gits` to discover which actions can be triggered by specific data patterns.

### `Neuron` (Worker Unit)

`Neurons` are the **distributed worker units** within `go-cyberbrain`. They operate in a continuous loop to find, claim, and execute `Jobs`.

* **Decentralized Scheduling Contribution:** After a `Neuron` successfully executes a `Job` and produces new data, it directly feeds this `result` back into the `Scheduler`. This makes the scheduling process decentralized, with each `Neuron` contributing to identifying and creating follow-up work.
* **Job Lifecycle Management:** A `Neuron` changes the state of `Job` entities in `gits` (e.g., from "Open" to "Assigned" to "Done" or "Error") as it processes them.
* **Dependency Injection:** `Neurons` inject references to `Gits`, `Mapper`, and `Archivist` into `Actions` that implement specific interfaces (e.g., `ActionExtendGitsInterface`), allowing actions to interact with core system components.

### `Scheduler` (Reactive Engine)

The `Scheduler` is the core of `go-cyberbrain`'s reactive behavior. Its primary role is to **identify potential `Jobs` based on new or updated information** and then create those jobs in `gits` for `Neurons` to pick up.

* **Trigger Mechanism:** It reacts to data that has been newly mapped into `gits` (identified by the `bMap` property on `TransportEntity` instances) and new relation structures.
* **Pattern Matching:** It compares the structure and types of the new data against the `Dependencies` registered by `Actions` in the `Cortex`.
* **Query Building:** It dynamically constructs `gits` queries to gather all necessary context and related data to satisfy an `Action`'s `Dependency`, forming the `input` for a new `Job`.

### `Demultiplexer` (Combinatorial Parser)

The `Demultiplexer` works closely with the `Scheduler` to ensure that all relevant interpretations of complex input data are considered.

* **Purpose:** It takes a single `transport.TransportEntity` (potentially with nested `ChildRelations`) and generates a slice of new `transport.TransportEntity` instances. Each resulting entity represents a unique combination of the original entity with one instance from each of its distinct child entity types.
* **Mechanism:** It recursively traverses the input graph segment and applies a combinatorial approach (effectively a Cartesian product) to produce all valid "perspectives" or "contexts" from the data.
* **Scheduler Input:** This detailed breakdown allows the `Scheduler` to match actions that might depend on specific combinations of related entities, ensuring no potential trigger is overlooked.

### `Mapper` (Knowledge Integrator)

The `Mapper` is responsible for **integrating `TransportEntity` data into the `gits` knowledge graph**.

* **Core Function:** `MapTransportData` is its main method, recursively walking `TransportEntity` structures and creating or updating corresponding entities and relations in `gits`.
* **ID Handling:** It uses special `ID` values (e.g., `storage.MAP_FORCE_CREATE`, `-1`, `-2`) to determine how entities should be handled (force creation, map by type/value, map by type/value/parent).
* **`bMap` Flag:** It sets a `bMap` property on newly created or modified entities and relations. This flag serves as a signal to the `Scheduler` that these are "newly learned" data points that need to be analyzed for potential job creation.
* **Concurrency:** It employs global mutexes (like `EntityTypeMutex`, `EntityStorageMutex`, `RelationStorageMutex`) to manage concurrent writes to the underlying `gits` memory. While global, these locks operate on fast in-memory hashtables, which is a deliberate choice for current performance characteristics.

### `Job` (Unit of Work)

A `Job` represents a **single unit of work** within `go-cyberbrain`. It's an entity stored in `gits` that captures all the necessary information for a `Neuron` to execute an `Action`.

* **Components:** A `Job` entity typically relates to an `Action`, a specific `Dependency`, and the `input` data required for execution.
* **Lifecycle:** `Jobs` transition through various states (e.g., "Open," "Assigned," "Done," "Error") which are also represented as entities and relations in `gits`.

### `Observer` (System Watchdog)

The `Observer` continuously **monitors the activity of the `go-cyberbrain` system**.

* **Idle Detection:** It queries `gits` to check the number of open `Jobs` and the state of `Neurons` (e.g., "Searching" for new work).
* **Version Tracking:** It tracks `Neuron` entity versions to detect recent activity, preventing premature system shutdowns.
* **Lifecycle Management:** Once the system is consistently idle (no open jobs, all neurons searching, and no recent activity for a set duration), the `Observer` can trigger a user-defined callback or initiate a system shutdown.

### `Archivist` (Logging)

The `Archivist` provides a standardized **logging mechanism** for `go-cyberbrain` components. It wraps a standard Go `log.Logger` and supports different log levels.

### `Util` (Helper Functions)

The `Util` package contains various **general-purpose helper functions** used throughout the system, such as generating unique IDs, manipulating maps, and checking the system's "Alive" state in `gits`.

---

## Data Flow & Lifecycle: The Sense-Think-Act Loop

The core operation of `go-cyberbrain` follows a continuous sense-think-act loop, driven by data flowing through its components:

1.  **Sense (Data Ingestion):** External data is provided to `cb.LearnAndSchedule()`.
2.  **Integrate (Mapping):** The `Mapper` ingests this data, converting it into `TransportEntity` objects and storing it in `gits`. Crucially, new or changed entities/relations are flagged with `bMap`.
3.  **Process Perspectives (Demultiplexing):** The `Demultiplexer` takes the newly mapped data and breaks it down into all possible relevant combinations or "perspectives" (one entity for each unique child type combination).
4.  **Think (Scheduling):** The `Scheduler` then analyzes these demultiplexed data entities. It matches them against the `Dependencies` of registered `Actions` (known via the `Cortex`). If a match is found and all required data can be gathered from `gits`, the `Scheduler` creates a new `Job` entity in `gits`.
5.  **Act (Execution):** An idle `Neuron` discovers an "Open" `Job` in `gits`, claims it (changes its state to "Assigned"), and executes the associated `Action` with the prepared input data.
6.  **Results & Reactivity:** The `Action` executes its logic and returns `result` data (also `TransportEntity` objects).
7.  **Loop Back:** The `Neuron` then immediately feeds these `result` data back into the `Mapper` (closing the `Job` beforehand), restarting the cycle from step 2. This creates a continuous, reactive flow.

---

## Key Design Principles

* **Graph-centric Memory:** All information resides in a flexible, interconnected graph.
* **Decentralized Processing:** Work is distributed among `Neurons`, enhancing potential parallelism.
* **Event-Driven Reactivity:** The system autonomously responds to changes in its knowledge graph.
* **Plausible Capabilities:** `Actions` provide a clear mechanism for extending functionality.
* **Self-Awareness:** The system's internal state is part of its own knowledge graph, enabling introspection.
* **In-Memory Performance:** Design choices like direct `gits` interaction prioritize speed for its intended in-memory operational model.
