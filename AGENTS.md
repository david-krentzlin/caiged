# Hard rules

* invoke `make qa` after every meaningful change and ensure it passes
* minimize the number concepts in the codebase. maximize cohesion within modules
* modules / packages should have a small interface but a deep implementation

## Definition of Done (must be followed)

1. Quality gate
   - Run `make qa` for the final state of the change.
   - A task is not done until `make qa` passes.

2. Automated tests
   - Add or update tests when behavior changes or bugs are fixed.
   - Run relevant test commands and verify they pass.

3. Manual validation (when applicable)
   - For CLI, integration, container, or UX flows, the agent must execute manual checks directly.
   - Record the exact commands run and the observed outcome.

4. Documentation updates
   - Update docs when behavior, flags, workflows, or troubleshooting changes.
   - Typical targets include `README.md`, `man/`, spin docs, and agent instructions.

5. Commit quality
   - When a commit is requested, include only relevant files.
   - Use a clear, concise commit message in imperative style that explains intent (the why), not just file changes.

6. Handoff quality
   - Report what changed, how it was verified (automated and manual), and any follow-up risks or caveats.
