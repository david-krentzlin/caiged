---
name: security review
description: Evaluate the change for security risks using layered controls and explicit threat modeling.
---

##  What to do

### Objective
Evaluate the change for security risks using layered controls and explicit threat modeling.

### Scope
- Input validation and encoding
- Authentication and authorization boundaries
- Secrets handling
- Logging and data exposure
- Dependency and supply-chain risk
- Abuse and adversarial scenarios

### Required Method
1. Identify trust boundaries
2. Enumerate relevant threat classes
3. Check for layered mitigations
4. Verify with tests, scans, or reasoning

### Constraints
- Prefer repo-native tools
- New security tools require user confirmation

### Outputs
- Threat model summary (short, explicit)
- Findings with exploitability assessment
- Conventional Comments with severity labels
- when you provide CVEs you must always provide evidence in the form of links, or executable commands that output those CVE findings.

### Severity Guidance
- Exploitable vulnerability → `blocking(security)`
- Missing layer / single point of failure → `issue(security)`
- Hardening opportunity → `suggestion(security)`

## When to use me

* When user asks for security review
* As part of review process for change or system
