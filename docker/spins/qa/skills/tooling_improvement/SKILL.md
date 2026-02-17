---
name: tooling and process improvement
description: Improve confidence and signal quality of QA systems without touching production code.

---
## What to do

### Objective
Improve confidence and signal quality of QA systems without touching production code.

### Allowed Changes
- Tests and fixtures
- Test utilities
- CI configuration
- Benchmark harnesses
- Scanning configuration
- Reporting and observability for tests
* Automation of local QA setups and general tooling to increase DevEx

### Confirmation Required For
- New external dependencies
- New scanners or services
- New mandatory CI gates
- Significant workflow changes

### Decision Rule

Prefer the smallest change that:
- Increases confidence
- Reduces flakiness
- Improves reproducibility
* Increases simplicity

### Outputs
- Description of tooling change
- Rationale
- How to run locally and in CI
* Documentation


## When to use me

* As part of the review process for a change or system
* When user asks for tool improvement in a codebase
