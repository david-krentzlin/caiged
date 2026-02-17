---
name: Discover Intent 
description:  Discover the intent of a change or task
---

## Objective
Verify that the change implements the **intended behavior** correctly and completely, as defined by explicit or derived acceptance criteria, and that it meets a clear definition of done.
This skill focuses on *what the change is supposed to do*, not just what the code currently does.

---

## Core Principle
If intent is unclear, incomplete, or contradictory, the review must pause and request clarification before proceeding to detailed testing.

## What to do

* Identify the **primary goal** of the change
- Identify **in-scope behavior**
- Identify **explicitly out-of-scope behavior**
- Detect ambiguous, underspecified, or conflicting intent


### Methods
- Read the linked issue or specification
- If no issue exists:
  - Infer intent from PR title and description
  - Treat inferred intent as *tentative*

### Mandatory Clarification
If any of the following are true, you MUST ask clarifying questions **before** finalizing acceptance criteria:
- Expected behavior is implicit or vague
- Success conditions are not measurable
- Edge cases are not addressed
- Backward compatibility expectations are unclear


## When to use me 

- When user asks you to implement a change
- When you are tasked to work on a specific problem in an underdefined way
* As part of the review process of for a change

---

## Inputs (in priority order)
1. Linked issue / ticket (preferred)
2. PR title and description
3. Commit messages
4. Code diffs (as a fallback signal only)


