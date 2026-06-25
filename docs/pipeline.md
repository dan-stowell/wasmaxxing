# Data pipeline

The pipeline turns a handful of community "awesome" lists into one normalized,
deduplicated catalog of the WebAssembly ecosystem: `data/catalog.json`.

## Seed sources

Committed under [`data/sources/`](../data/sources/) so the pipeline runs without
network access:

| File | Upstream | Format |
|------|----------|--------|
| `wasm-langs.md` | [appcypher/awesome-wasm-langs](https://github.com/appcypher/awesome-wasm-langs) | `langs` |
| `wasmlang-org.md` | [wasmlang.org](https://wasmlang.org/) (fork of the above) | `langs` |
| `wasm-runtimes.md` | [appcypher/awesome-wasm-runtimes](https://github.com/appcypher/awesome-wasm-runtimes) | `runtimes` |
| `awesome-wasm.md` | [mbasso/awesome-wasm](https://github.com/mbasso/awesome-wasm) | `awesome` |

To refresh the seeds, re-download the upstream `README.md` files into
`data/sources/` (the pipeline is tolerant of layout drift but parsers may need
updates).

## Stages

### 1. Build the catalog (offline)

```sh
bazel run //pipeline/cmd/build-catalog
```

Parses every seed via the matching parser in `pipeline/parse`, then
`catalog.Build` normalizes IDs, extracts GitHub `owner/repo` from URLs,
deduplicates (by repo, then URL, then name+kind), merges descriptions/sources,
and writes sorted pretty JSON to `data/catalog.json`.

### 2. Enrich with GitHub metadata (online)

```sh
GITHUB_TOKEN=$(gh auth token) bazel run //pipeline/cmd/enrich-catalog
```

For each entry with a GitHub repo, fetches stars, forks, primary language,
license, last-push time, archived status, topics, and homepage. Failures (404,
rate-limit) are recorded per-entry in `github.error` rather than aborting.
Flags: `-limit N`, `-only-missing` (default true), `-delay`.

## Catalog schema

`data/catalog.json` is `{ "generated_at": ..., "entries": [Entry, ...] }`.

```jsonc
{
  "id": "wazero",                 // stable slug from name
  "name": "wazero",
  "kind": "runtime",              // language|compiler|runtime|tool|project|resource
  "description": "...",
  "url": "https://wazero.io",
  "repo": "tetratelabs/wazero",   // owner/name if GitHub, else empty
  "category": "runtime",          // originating section/heading
  "related_language": "Rust",     // for compilers listed under a language
  "sources": ["awesome-wasm-runtimes"],
  "github": {                      // present only after enrichment
    "full_name": "tetratelabs/wazero",
    "stars": 5300, "forks": 300, "open_issues": 40,
    "language": "Go", "license": "Apache-2.0",
    "archived": false, "pushed_at": "2026-...",
    "topics": ["webassembly", "wasm"]
  }
}
```

### Kinds

- `language` — a source language that compiles to / has a VM in wasm.
- `compiler` — a tool that emits wasm.
- `runtime` — an engine that executes wasm modules.
- `tool` — operates on wasm (inspect/optimize/convert).
- `project` — an app or library built with/on wasm.
- `resource` — docs, tutorials, playgrounds, articles.

## Tests

```sh
bazel test //pipeline/...
```

Parsers and the catalog model are unit-tested against inline fixtures, so the
pipeline's core logic is verified without network or the large seed files.
