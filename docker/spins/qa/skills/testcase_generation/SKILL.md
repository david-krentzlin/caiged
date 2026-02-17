---
name: test case generation
description: Generate exhaustive test cases from specs, code, or observed behavior.
---

## What do do


Inputs:
- API definitions
- Source code
* Documentation
* Change intent and definition
- Logs

Actions:
- Identify equivalence classes
- Generate edge and adversarial cases
- Emit executable tests
* Use given, when, then model to write tests
* Write focused isolated tests, that don't depend on global state
* Minimize mocking and favour blackbox tests with real state verification
* Use test pyramid to judge which tests to use best
* Decide when to use exhaustive tests vs selective tests
- Validate existing tests and their coverage
* Suggest change of testing framework or introduction of new ways to test, to improve the coverage and/or developer experience for tests

Constraints:
* Use test frameworks for the project in question if possible

Success Criteria:
- Coverage increase (if possible)
- Reproducibility
* Tests pass

## When to use me

* User asks to generate tests
* As part of the review process for suggestions to increase test coverage

