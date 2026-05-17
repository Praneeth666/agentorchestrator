---
name: go-code-review
description: Review Go code for correctness, idioms, and performance. Trigger on any request to review, audit, or improve Go code, or when Go code is pasted with "what do you think" / "any issues".
---

Review the Go code for: goroutine leaks, unsynchronized shared state, swallowed errors (`_, _`), missing `defer` on `Body.Close()`, non-idiomatic patterns (error strings, fat interfaces, `context` in structs), and unnecessary allocations. Run `go vet ./...` if bash is available. Output: 🔴 Must Fix → 🟠 Should Fix → 🟡 Consider, each with file:line, issue, and a concrete fix. Be direct and skip categories with nothing to say.