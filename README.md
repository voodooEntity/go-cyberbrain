## `go-cyberbrain`

A reactive, graph-driven system for knowledge processing and automated task execution.

---

### Documentation

* [Getting Started](docs/GETTING_STARTED.md)
* [Core Concepts & Architecture](docs/CONCEPTS.md)
* [Developing Custom Actions](docs/ACTIONS.md)

---

### What is `go-cyberbrain`?

`go-cyberbrain` is a Go-based framework designed to manage and process information as a **knowledge graph**. It uses a decentralized approach to identify new data, apply defined **Actions**, and continuously expand its understanding or perform tasks. The system is designed for scenarios where data relationships are key, and automated responses to new information are required.

---

### Key Features

* **Graph-based Memory:** Uses **`gits`** (Graph-based In-memory Transactional Storage) to store all information, including system state and learned data, as an interconnected graph.
* **Decentralized Work Units:** **`Neurons`** act as independent workers that discover, claim, and execute **`Jobs`**.
* **Reactive Scheduling:** The system autonomously creates new **`Jobs`** by detecting patterns in newly acquired or modified knowledge. This process is driven by the **`Scheduler`**, which reacts to changes in the graph. **Crucially, scheduling is decentralized; `Neurons` contribute to the scheduling process directly after completing their tasks.**
* **Pluggable Actions:** Developers can extend the system's capabilities by implementing custom **`Actions`** that perform specific tasks or integrate with external systems.
* **Self-Awareness:** `go-cyberbrain` stores its own operational state and configuration within its `gits` memory, allowing for introspection and dynamic management.

---

### Possible Use Cases

`go-cyberbrain`'s reactive, graph-driven nature lends itself to scenarios requiring automated information processing and response. Examples include:

* **Cybersecurity Analysis:** Identifying related security incidents, mapping network assets and vulnerabilities, and triggering automated remediation steps based on discovered relationships (e.g., "if this IP is malicious and connected to this internal host, then block it").
* **Infrastructure Automation:** Managing the state of dynamic infrastructure, where changes in one component (e.g., a new service deployment) can automatically trigger configuration updates or health checks in related components.
* **Data Lineage and Transformation:** Tracking how data flows through various systems, identifying data quality issues, and triggering transformation or enrichment processes based on data characteristics or new inputs.
* **Supply Chain Monitoring:** Mapping complex supply chain relationships, detecting anomalies (e.g., delays from a specific supplier affecting downstream processes), and initiating alerts or alternative routing actions.
* **IoT Device Management:** Monitoring state changes in interconnected devices, processing sensor data, and automating control actions or alerts based on predefined rules and contextual information derived from the graph.

---

### Getting Started

To run the example demonstrating how `go-cyberbrain` resolves IP addresses for a domain, refer to the [Getting Started Documentation](docs/GETTING_STARTED.md).

---

### Project Status

`go-cyberbrain` is currently in a **Proof-of-Concept (POC) / pre-alpha state**. Its primary goal is to demonstrate the core mechanics of a reactive, graph-driven system. While functional, it is not yet production-ready and is undergoing active development and refinement.

Due to its volatile state of development, contributions are generally **not being sought or applied at this time**. The project's structure and APIs are subject to change without prior notice.

---

### License

`go-cyberbrain` is open-source software licensed under the [Apache License](LICENSE).