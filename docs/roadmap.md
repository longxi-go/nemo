# Nemo roadmap

[English](roadmap.md) | [简体中文](roadmap.zh_CN.md)

| Stage | Capabilities | Available behavior | Status |
| --- | --- | --- | --- |
| Configuration sources | Immutable values, file/OS/map sources, JSON/YAML/TOML/INI/properties codecs, profile-based discovery and application decoder registries | Compose explicit sources without coupling transport to decoding | Implemented |
| Environment | Source precedence, placeholder resolution, profiles, immutable snapshots, validation and refresh | Prepare a complete candidate snapshot before atomic publication; retain the prior snapshot on failure | Implemented |
| Typed binding | Scalar/object collections, named/anonymous nesting, bounded recursion, annotation compiler and CLI | Generate typed `BindProperties` methods and validate output before writing | Implemented |
| Update integration | Provider watch and change notification interfaces | Connect an application-owned update transport to refresh and cancellation | Implemented |
| Resource and concurrency limits | Candidate budgets, bounded file input, cancellable refresh admission and coalesced invalidation | Reject oversized candidates before publication and retain an update arriving during refresh for a subsequent pass | Implemented |
| Remote adapters | Nacos and other configuration-center transports | Define SDK ownership, version freshness, reconnect and cancellation behavior for each adapter | Planned |
| Framework integration | Application startup and shutdown integration | Specify initial-load failure handling and lifecycle ownership in the consuming framework | Planned |

The implemented stages include tests for supported behavior and failure paths.
They do not establish compatibility with a concrete remote service. Remote
adapters and framework integration require separate designs and integration tests;
no remote SDK adapter is included in the current module.
