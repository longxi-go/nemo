# AGENTS.md

## Project Purpose and Scope

- nemo implements configuration environments against gosr shared contracts.
- gosr owns portable property, profile, snapshot and static-binding semantics; nemo owns source loading, codecs, composition, resolution, atomic refresh and generation.
- Separate transport from decoding. File and remote providers supply immutable source snapshots; Nacos SDK details never enter gosr or nemo core APIs.
- Property lookup performs no external I/O. Refresh prepares and validates a full candidate before atomic publication; failures retain the previous generation.
- Runtime binding MUST NOT use reflection or unsafe. Use typed converters and generated BindProperties methods. Format libraries may use their own parsers internally.
- No global mutable loader registry or event bus. Use instance-owned provider, codec and notification dependencies.
- Public code belongs in pkg, implementation in internal, thin root/cmd only.

## Contract-First Development

- Current source and explicit user decisions define the development baseline. Git history provides context, not authorization to replace the current design.
- Complete and verify the relevant gosr environment contracts before implementing dependent nemo behavior.
- Prior implementations may inform capability and defect analysis; they are not the authoritative design.
- Do not add runtime reflection, unrelated utility APIs or compatibility aliases to satisfy obsolete consumers. Handle breaking integrations as explicitly scoped migrations.
- Follow the stage roadmap and complete usable capability bands with tests, examples, diagnostics and quality gates.

## Engineering Mandates

### Forbidden

1. MVP-only delivery: presenting temporary, partial or demonstration-grade work as complete production capability.
2. Untested production behavior or placeholder implementations presented as finished work.
3. Significant implementation without a clear design, package boundaries and acceptance criteria.
4. Feature accumulation that creates oversized, tangled or difficult-to-maintain packages.
5. Mixing unrelated responsibilities or leaking implementation details across architectural boundaries.
6. Isolated feature-by-feature delivery without considering the stage roadmap and known dependencies.
7. Designing each next feature only after finishing the previous one.
8. Speculative interfaces, factories, providers, registries or design patterns without an actual architectural purpose.
9. Silent changes to public contracts, configuration, errors or lifecycle behavior without assessing compatibility.
10. False completion claims, hidden verification gaps or weakened tests/linters used to obtain a passing result.

### Expected

- Prioritize correctness, maintainability, architectural coherence, testability and necessary extensibility.
- Prefer the simplest design that fully satisfies known requirements and preserves real extension boundaries.
- Encapsulate state and implementation details. Use composition, interfaces and design patterns where they solve an identified problem.
- Keep dependencies, responsibilities, naming, ownership and side effects explicit.
- Build complete, verifiable capability slices; do not substitute superficial feature counts for usable behavior.
- Simplicity must not excuse incomplete engineering; extensibility must not excuse unnecessary abstraction.

## Framework Design Expectations

- Each package MUST have one cohesive responsibility and a minimal intentional API.
- Before creating or expanding a package, justify its responsibility and dependency direction.
- Avoid dumping grounds such as `common`, `misc`, `helpers`, `utils` or vaguely named managers.
- Define small behavioral interfaces at real shared or consumer boundaries. Do not merely mirror concrete types.
- Prefer composition over inheritance-like embedding hierarchies. Adopt useful Spring IoC ideas through Go-compatible contracts and ownership.
- Shared APIs must remain independent of unnecessary transports, storage engines, runtimes and compiler internals.
- Runtime selection, registration, activation, bootstrapping and cleanup must have explicit owners in the responsible implementation.
- Extension points must specify initialization, lifecycle, failure and cleanup semantics where applicable.
- Default implementations must permit documented alternatives; registration, selection and configuration precedence must be deterministic.
- Registration must be discoverable and documented. Generated `init` may contribute definitions; it MUST NOT construct Beans, perform external I/O or start background work.
- Prefer standard-library solutions when adequate. Justify third-party dependencies by their value, compatibility and maintenance costs.
- Foundational contracts must not depend on higher-level implementations. Avoid import cycles and unnecessary transitive dependencies.
- Do not hide trivial dependencies behind artificial abstraction layers or transplant Java class-loading/inheritance machinery into Go.

## Package Layout and Go Conventions

- Put public APIs and contracts in `pkg/<capability>`; put private implementation in `internal/<responsibility>`.
- Keep root Go code deliberately small, such as package documentation or a thin facade. Do not collect domain logic and helpers there.
- Executables belong in `cmd/<tool>` with thin entry points delegating to cohesive packages.
- Do not create empty directories to fill a template. Respect explicitly deferred layout changes; do not restore superseded packages merely to match a layout.
- Every owned Go package MUST contain `doc.go` and `example_test.go`, with meaningful examples that run under `go test`.
- Use precise package names, minimal exports, consistent receiver naming, straightforward control flow and focused functions.
- Follow the Go version in `go.mod`; verify language/tool support rather than assuming older-version restrictions.
- Use generics and code generation when they improve static safety or wiring. Keep their complexity contained; neither is prohibited merely as metaprogramming.
- Format owned source with the pinned golayout implementation through `make format`; gofmt or gofumpt alone is not the formatting gate. Generated Go must meet the same compilation, formatting and quality expectations as handwritten Go.
- Handle errors explicitly; preserve causes with `%w` where appropriate and classify with `errors.Is`/`errors.As`.
- Do not discard meaningful errors, rely on error strings for control flow or use panic for ordinary operational failure.
- Document cleanup failures and provide useful diagnostics without exposing sensitive data.
- Pass `context.Context` through context-aware operations; respect cancellation and deadlines. Do not retain request contexts in long-lived objects.
- Define ownership of goroutines, connections, timers and other resources. Release them on success, failure, cancellation and shutdown as applicable.
- Minimize mutable globals, hidden dependencies and initialization side effects.
- Make concurrency safety and limits explicit. Use synchronization intentionally; avoid races, leaks, unbounded work and concurrency without a useful purpose.

## Go Code Layout and Readability

- Keep each `.go` file focused on one cohesive responsibility. Split distinct
  responsibilities into accurately named files within the same package; package
  visibility does not require putting related capabilities into one large file.
- In compiler packages, `pipeline.go` owns orchestration only. Keep invocation state, step
  execution, each individual `*_step.go`, semantic modeling, rendering and output
  persistence in their respective files. Do not collect step implementations in
  the orchestration file or introduce an unnecessary package for each step.
- Every nonempty executable block MUST use multiple lines. This includes named functions,
  methods, anonymous functions/closures, callbacks, forwarding methods, lifecycle
  hooks, and deferred or goroutine function bodies, even when they contain one statement.
- Nonempty control-flow bodies MUST use multiple lines: `if`/`else if`/`else`, `for`/
  `range`, expression/type `switch`, `select`, and standalone/labeled blocks.
  Put each case/default body on following indented lines; do not place executable
  statements after `case ...:` or `default:` on the same line.
- For nonempty blocks, keep opening braces on their normal Go header line and
  closing braces on a separate line. Standard `} else if ... {` / `} else {` header placement stays
  valid. Do not compress multiple body statements with semicolons; semicolons in
  conventional `if` / `for` headers remain valid.
- These rules apply equally to handwritten source, tests, examples, generated Go
  and generator templates. Canonical empty type/literal forms such as `struct{}`
  and `Config{}` remain compact. Empty functions and control-flow blocks, including `select {}`, may also retain canonical single-line forms.
- Separate top-level declarations with a blank line, keeping doc comments attached
  to their declaration. Separate distinct types inside grouped declarations too.
- Every named, non-empty struct declaration MUST use multiple lines, even when
  it contains only one field. Empty structs retain the canonical `struct{}` form
  required by gofmt/gofumpt; do not fight the formatter with artificial comments.
- Non-empty struct literals with multiple fields MUST use multiple lines, with
  one field per line. Prefer that layout for single-field literals when it makes
  construction intent easier to scan.
- Non-trivial map and slice literals MUST use multiple lines when that improves
  readability, especially for nested values or collections of definitions.
- Avoid excessively long lines in function declarations/calls, composite
  literals, boolean expressions and nested expressions. Split at logical
  boundaries rather than relying on a fixed-width formatter.
- Prefer splitting long or multi-step fluent/method chains at call boundaries,
  usually one chained call per line, when it makes the sequence easier to follow.
  Short, clear chains may stay on one line. This is readability advice, not a
  mandatory line-width or call-count rule and not an automatic formatter gate.
  Keep the continuation dot on the preceding line in Go; consider named
  intermediate values when they read better. Apply the same advice to tests
  and generator templates, without mechanically splitting every selector chain.
- Prefer explicit, readable layouts over compact expressions. Do not collapse
  deliberately structured multiline code into less readable single-line forms.
- The shared `github.com/longxi-go/golayout` project is the sole owner of structural layout and gofumpt composition. Do not copy its parser, CLI or scripts locally.
- `make format`, `make check-format` and generation MUST use the same pinned golayout module and its public `pkg/format` API or CLI.
- Empty structs, interfaces and literals retain canonical compact forms. A
  standalone gofumpt invocation, including `-extra`, does not enforce this layout.
- File responsibilities and logical line breaks remain source-design/review
  requirements; a formatter cannot infer appropriate architectural boundaries.
- Apply the same readability requirements to generated Go. Templates own
  structural layout; the formatting stage runs before validation/output.
- After changing Go source, run `make format`, then `make check-format`.

For a chain whose setup sequence is easier to scan across lines, prefer:

```go
t, err := template.New("codegen").
    Option("missingkey=error").
    Funcs(functions).
    ParseFS(source, patterns...)
```

Avoid:

```go
type Config struct{ Port int }
config := Config{Port: 8080}
```

Prefer:

```go
type Config struct {
    Port int
}
config := Config{
    Port: 8080,
}
```

Avoid compact executable bodies:

```go
func (s *Service) Ready() bool { return s.ready }
if err != nil { return err }
defer func() { close(done) }()
```

Use the same expanded layout for every nonempty executable block:

```go
func (s *Service) Ready() bool {
    return s.ready
}

if err != nil {
    return err
}

defer func() {
    close(done)
}()

switch state {
case Ready:
    start()
default:
    return errNotReady
}
```

## Interface and Constraint Naming

- Use named interfaces for behavioral contracts in API signatures, fields, type
  assertions and generic constraints instead of inline non-empty `interface {...}`.
  A single use is not a reason to hide the contract inside an expression.
- Reuse a suitable existing interface, including standard-library contracts,
  before defining another. Keep consumer-specific constraints and optional
  capabilities local; promote only genuinely shared contracts into gosr.
- Prefer named structs for API and persistent data shapes. Local anonymous
  structs in test tables or small examples remain useful; empty marker types,
  `any` and intentional compiler fixtures are not violations by themselves.
- Review source, tests and generation templates together. Distinguish actual
  anonymous contracts from serialized Go syntax and deliberate test inputs.
- Keep independent signature assertions and legacy compatibility fixtures
  explicit; replacing their expected shapes with production types weakens them.

## Function Types and Constant Ownership

- Prefer semantic function type aliases (`type DecodeFunc = func(...)`) in
  parameter and callback-field declarations instead of repeating inline
  `func(...)` signatures. This also applies to readable signatures used once;
  repeated use is not a prerequisite. Choose names by responsibility rather
  than forcing all callbacks into one generic name such as `ValueFunc`.
- A function type alias uses `=`. Choose a defined function type instead only
  when distinct type identity or methods serve an explicit design purpose.
- At call sites, prefer named functions or clearly named local callbacks over
  long inline function literals. Short, clear or necessary capturing closures
  remain allowed; this is a readability preference, not an absolute ban.
- Place shared function contracts beside the types/capability they describe in
  gosr. Keep implementation-specific callbacks in their owning project/package;
  matching signatures alone do not establish shared semantics.
- Extract literals when a semantic name clarifies intent or one owner must keep
  repeated protocol, syntax or policy values consistent. Prefer the standard
  library's existing constant when it already expresses the intended meaning.
- Prefer named constants over hardcoded protocol tokens, marker names, format
  identifiers, generated artifact names and meaningful policy defaults. Keep one
  owner for each meaning; derive related lengths and values from that owner.
- Put ecosystem contract constants in the relevant gosr capability package;
  keep project-specific constants local to that project. Avoid a broad
  `constants` package that mixes unrelated domains.
- Do not merge literals solely because their spelling matches or replace
  self-explanatory values mechanically. Keep ordinary zero values, loop steps,
  local syntax and independent test fixtures literal when clearer.
- When parsing shared syntax, derive cursor offsets from token lengths instead
  of duplicating magic widths. Generated source and templates follow the same
  function-type and constant-ownership rules.

## Shared Compiler Infrastructure

- Reuse gosr/pkg/pipeline and cohesive gosr/pkg/codegen mechanisms for compiler execution, imports, artifacts, diagnostics, strict templates and staged output.
- Reuse gosr/tools/pkg/packageload and annotation for tool-dependent adapters; do not copy their mechanism implementations.
- Keep annotation schemas, semantic models, module/isolation policies and publication/report data local. Keep golayout as the sole Go formatter.
- The nested gosr/tools module must be resolved explicitly in local consumer fixtures; parent module replace directives do not propagate.

## Multi-Repository Development

- A parent `go.work` may select local ecosystem source for development. Include
  `gosr/tools` explicitly when needed; selecting the parent module does not include
  nested modules. Standalone cloning and published dependency checks must not require
  a sibling checkout or a developer-specific workspace.
- Production `go.mod` files declare dependencies and MUST NOT contain local
  sibling-directory replacements. Temporary consumer fixtures may retain explicit
  local replacements to exercise isolated module behavior.
- Keep go.work, go.work.sum, execution state, caches, backups, OpenSpec and
  superpowers planning artifacts outside the publishable change set.
- Repository checks and CI MUST resolve the published module graph with GOWORK
  disabled. Integration fixtures may replace nemo itself to test the current
  checkout, but MUST NOT require sibling repositories or their source trees.
- Review golayout upgrades intentionally and keep the formatter CLI and generation
  library on the same module version. Derive the target language from the owning
  package's module; cross-compilation checks must use the target architecture.
- Unpublished dependency placeholders are for workspace development only. Verify
  standalone module resolution after actual dependency versions are published;
  do not treat a workspace-only pass as proof of standalone release readiness.

## Design and Implementation Workflow

- Before a meaningful change, inspect current code, dependencies, relevant documents, extension seams and compatibility constraints.
- Current user instructions and current source are authoritative; stale documents and cached summaries do not override them.
- Establish the problem, scope/non-goals, architectural placement, public contracts, ownership, failure behavior, testing strategy and acceptance criteria.
- Consider relationships between known capabilities before implementation. Do not speculate beyond established needs.
- Keep design effort proportional; small fixes do not need unnecessary ceremony.
- Use existing suitable abstractions, but do not copy a demonstrably unsuitable pattern or retain a flawed design merely because it already exists.
- Implement against the agreed design with cohesive, reviewable changes. Keep code, tests and documents aligned.
- Reassess a flawed abstraction when evidence warrants it; do not continually patch around architectural defects.
- Distinguish required capability work, necessary supporting work, related optional work and unrelated changes.
- Do not add speculative features or omit essential correctness, testing or integration to reduce scope.
- Evaluate downstream effects of public API, configuration, error and lifecycle changes. Document intentional breaks and migration implications.
- Explicitly authorized breaking changes do not need repeated permission; do not resurrect obsolete APIs or add compatibility shims without a concrete requirement.

## Testing Expectations

- Every production package MUST be covered by focused tests. Contract-only packages require meaningful usage, semantic or compile-time examples rather than a production runtime built just for testing.
- New or modified behavior must have automated tests covering ordinary behavior, edge cases and meaningful failure paths.
- Prefer behavioral, deterministic and isolated tests; use table-driven cases when they improve clarity.
- Do not mirror implementation details, rely on incidental order/globals or add meaningless tests for coverage.
- Test new extension points with default and alternative implementations, registration/selection, lifecycle and relevant invalid-use cases.
- Integration tests must verify source loading, decoding/resolution/binding, refresh failure preservation, notification/resource ownership and generated output through nemo public boundaries.
- Configuration precedence and external integration tests are required when those capabilities are in the authorized scope, not merely because a generic checklist mentions them.
- Practical bug fixes SHOULD include a regression reproducing the original failure and asserting the corrected behavior.
- Use TDD for new behavior: observe the failing behavioral test before implementing its solution.
- Run race detection for applicable concurrent/shared-state changes; meet configured coverage thresholds without treating percentages as proof of quality.
- A conformance suite must name its version and exact coverage. Do not equate a limited catalog, test fixture or compiling interface with complete conformance.

## Test Go Source Fixtures

- Prefer per-scenario `testdata/<scenario>/*.go.tmpl` files for substantial Go
  programs compiled, generated or inspected by tests. Keep scenario names and
  expected diagnostics in the test table; keep declarations and runtime checks
  readable as normal Go source in the fixture.
- Read static fixtures directly. Use the existing strict template renderer only
  when substitutions are needed; keep view models named and substitutions small.
  Do not add a general fixture framework for simple file reads.
- Complete `.go.tmpl` files under `testdata` opt into the shared golayout
  `make format` / `make check-format` gates. They must parse as Go before template
  execution; keep template actions inside quoted strings or comments when possible.
- Short, clear inline snippets may remain. Intentionally malformed syntax or
  compact inputs that test formatting must retain their exact shape; use clearly
  named `.go.txt` files or brief inline samples rather than opting them into
  formatting. Never normalize away the behavior a negative test must exercise.
- These are readability recommendations, not a requirement to externalize every
  literal. Preserve annotations, build constraints, tags and test assertions when
  moving source into fixtures.

## Documentation Standards

- Prefer diagrams before detailed prose when explaining architecture, dependency
  directions, workflows, states or interactions. Use Mermaid flowcharts, sequence
  diagrams or state diagrams as appropriate; do not force diagrams onto simple facts.
- Keep diagram labels and technical meaning aligned in English and Simplified
  Chinese. Distinguish existing mechanisms, normative obligations and planned work.
  Retain ownership, failure, concurrency and compatibility details in concise prose.
- Validate diagram rendering as well as Markdown builds; a successful static build
  does not prove Mermaid syntax or visual readability.
- Every package MUST have `doc.go` and `example_test.go`; this is not limited to core packages or optional discoverability improvements.
- Package documentation must explain purpose, responsibility, intended use, constraints and extension boundaries.
- Exported contracts, types and important functions MUST have meaningful Go doc comments.
- Source comments, package documentation and API comments MUST be in English.
- Explain why an abstraction exists, its invariants, design choices and prevented misuse. Do not restate source line by line.
- Describe relevant ownership, lifecycle, concurrency, cancellation, failure and compatibility semantics.
- Keep examples compilable, executable and representative of supported behavior. Do not add empty examples merely to satisfy filenames.
- Update documentation with contract changes. Keep comments precise, necessary and consistent with actual behavior.
- Source-level documentation must make architectural seams understandable without requiring every external design document.
- Maintain English and Simplified Chinese README and public manual pages together.
  Review both against source before committing, including commands, prerequisites,
  ownership, failure behavior, versions and links. A strict build does not prove
  semantic accuracy or translation parity.
- Normative requirements belong in tracked bilingual documentation, never only in
  ignored architecture notes. Keep historical plans and task records excluded.
- Public README and manual pages address developers using or extending nemo.
  Keep local task progress, review verdicts, acceptance records and personal
  migration explanations out of them. Preserve technical limits, failure behavior
  and ownership rules; distinguish implemented behavior from planned adapters.
- Root release tags use vX.Y.Z with maintained English notes. Before tagging,
  run the full source checks and live release preflight, including the repository's
  Pages build type and github-pages tag deployment policy. A main branch policy
  does not authorize tag deployments. Follow docs/maintainers/releases.md.
- Release workflows MUST check the exact tagged source, publish the stable Release
  before deploying its tagged documentation, and keep branch builds from deploying.
  Verify public installation with fresh module caches and no workspace/replacements;
  do not equate a source check or GitHub Release with public module availability.
- Never move a published tag or disable checksum verification to complete a release.

## Linting and Quality Gates

- `codequality/HEADER` is the authoritative plain-text Go source-file notice.
  Preserve its copyright owner and year policy; do not infer ownership from Git.
- Use `make license-headers` to refresh owned `.go` files and
  `make check-license-headers` for the read-only gate included in `make check`.
- Shared notice logic lives in `golayout/pkg/header`; generators apply the output
  owner's optional template before formatting/validation, keeping generated markers
  first. Do not duplicate notice-replacement logic in consumers.
- Exclude ignored files, backups, caches, vendor, testdata and symlinks from bulk
  refreshes. A different existing copyright owner requires an explicit decision.

- golayout owns the structural-layout policy and pinned gofumpt dependency. Pin its module version, review formatting-policy upgrades intentionally and do not enable optional `-extra` rules independently.
- `make setup` MUST verify the shared golayout tool module resolves; it must not install a consumer-local formatter binary or edit module files.
- `make check-format` MUST fail on structural-layout or gofumpt drift without rewriting source or module files. Invoke the shared CLI from the module declaring its tool dependency; do not maintain a private formatter copy. `make check` and CI MUST include this gate and resolve the same checksum-verified golayout version. A sibling tooling checkout is not required.
- Format/check only Git-owned or unignored Go source, including generated files explicitly. Exclude `.tmp`, `.cache`, dependency/vendor trees and ordinary intentional `testdata` samples; opt complete `testdata/**/*.go.tmpl` fixtures into the shared gate. Never recurse blindly over `.`.
- `golangci-lint` is mandatory for code delivery. Run `golangci-lint run --timeout=10m` with zero issues before declaring code complete.
- Satisfy all enabled linters, including `errcheck`, `staticcheck`, `unused` and applicable related checks.
- Do not suppress findings merely to pass validation. Any permitted suppression needs a specific justified reason.
- Migrate deprecated APIs/patterns when an appropriate replacement exists. Retention needs a documented intentional compatibility reason.
- Resolve introduced lint/test failures in the same capability slice. Investigate existing failures and disclose them accurately.
- For code changes, verify formatting of owned Go files, run `go test ./...`, `go vet ./...`, required race/integration/coverage checks and lint.
- Run `make check` when the project defines that target; satisfy any additional applicable project gates.
- Limit formatting/file scans to owned source; ignored snapshots and dependency caches are not current packages.
- Inspect actual output and exit status. Report skipped, unavailable and failed checks with reasons; do not make an unqualified completion claim while required gates fail.
- Do not weaken tests, disable linters or remove gates to obtain success.
- Instruction/documentation-only tasks require diff, requirement and reference validation; they do not require unrelated Go builds/tests.
- Never restore deleted source to make a gate run or claim that an empty source tree proves implementation correctness.

## Delivery Standard

- Design, implementation, tests, examples and documentation must evolve together.
- Applicable OpenSpec artifacts must stay consistent with current code and requirements.
- Satisfy acceptance criteria with coherent boundaries, intentional APIs and documented compatibility implications.
- Do not leave placeholders, unexplained critical TODOs or unresolved material design conflicts presented as complete.
- Verify error handling, resource/lifecycle ownership and applicable concurrency behavior.
- Required unit, integration, extension, regression, race, coverage and lint checks must finish and their results must be inspected.
- Keep source easy to read, extend and replace in parts; do not introduce unrelated functionality or complexity.
- Completion requires current evidence for the delivered revision, not optimism, successful compilation alone or an agent's report.

## Roadmap Governance

- A stage-based roadmap MUST exist before large-scale iteration continues.
- Define each stage's capability bands, dependencies, usable outcome and acceptance criteria.
- OpenSpec changes MUST derive from the roadmap and group work by stage/capability, not isolated feature impulses.
- Complete cohesive stage-capability slices; do not invent one disconnected feature after another.
- If stages are marked complete but their promised usable outcome is absent, the roadmap/decomposition is wrong and MUST be corrected before further implementation.
- Usable nemo stages provide working source composition, configuration resolution, typed binding and refresh behavior for the supported scope, with failure diagnostics and extension boundaries.
- Roadmap updates are required development work. They do not authorize implementation or artifact restoration outside the current scope.

## Execution, Verification and Completion

- Work only within current authorization. A rules review does not authorize application code, scaffolding, reference engines, new specs or another repository's implementation.
- Preserve user changes, staging and deletions. Inspect current files before modifying them; never restore rejected designs to simplify a task.
- Continue authorized implementation through verification and completion. Do not substitute a plan, partial handoff or tool-recovery exercise for the deliverable.
- A failed tool/test, unfamiliar package or large task is not by itself a blocker. Inspect evidence, adjust the approach, retry when justified and continue.
- Do not repeat unchanged failures or busy-poll. A real blocker requires missing information, authority or an external dependency with no useful independent work remaining.
- Resolve genuine material ambiguities before dependent work; do not repeat approval questions for settled requirements.
- For permitted parallel work, assign clear path ownership and verification responsibility. An agent report does not establish integrated correctness.
- Keep changes/commits scoped. Do not stage unrelated user deletions, force history changes or publish externally without current authorization.
- Before completion, inspect the intended diff and map multi-area acceptance requirements to current implementation/artifacts and actual verification.
- Reassess verification after relevant changes; do not present stale passes as current evidence.
- Report actual changes, architectural impact, tests, commands/results, documentation and material remaining issues in proportion to the task.
- Reports MUST state completed changes, verification results, unverified work and known issues; use a format appropriate to the change.
- Distinguish verified, unverified, partial, unimplemented and genuinely blocked work. State blockers, impact and skipped checks; never invent results or silently lower acceptance criteria.

## Instruction Maintenance

- Keep project instructions focused on shared engineering and contribution requirements. Personal editor/agent configuration, task-state protocols and report templates belong in local configuration.
- Preserve applicable continuation, verification, completion and instruction-protection requirements when editing project rules. Do not weaken them without explicit authorization.
- When editing rules, compare requirements section by section against the originals; retaining headings alone is insufficient.

## Final Engineering Standard

- Define responsibilities and verify the design before expanding implementation.
- Use idiomatic Go and explicit contracts while preserving typed configuration binding, snapshot ownership and transactional refresh semantics.
- Deliver capabilities whose stated behavior has been verified.
- Maintainability, complete behavior and proportionate process are all required.
