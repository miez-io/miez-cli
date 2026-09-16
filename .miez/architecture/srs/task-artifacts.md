# SRS: task artifacts and worker invocation

ISO/IEC/IEEE 29148 software requirements specification using EARS.

**Status:** implemented target contract, drafted 2026-09-16.

## 1. Introduction

This capability adds reusable task artifacts to a miez team. A task captures
what a worker should do for a particular kind of request. The selected worker
contributes the persona and judgment, assigned skills contribute topic-specific
know-how and guardrails, and the user's message contributes the immediate
context and desired outcome.

The four artifact roles are deliberately separate:

- **Worker:** who the worker is: identity, motivation, goals, beliefs,
  judgment, and boundaries. Every worker resolves to a selectable GitHub
  Copilot custom agent in the target contract.
- **Skill:** how to solve a topic: reusable, worker-independent procedures,
  knowledge, and guardrails assigned to workers.
- **Task:** what to do: a partial reusable todo or work objective that can be
  applied with different workers and skill sets.
- **Workflow:** how to chain work: ordered coordination of workers and task
  assignments for a more automated process. miez renders the coordination;
  GitHub Copilot remains responsible for execution.

## 2. Stakeholders and scope

- **Stakeholders:** team authors, repository operators, GitHub Copilot users,
  and maintainers of the miez compiler.
- **Actors:** a team author, a Copilot user, the miez builder/compiler, and
  GitHub Copilot.
- **In scope:** task source artifacts, generated task catalog entries,
  Copilot task prompts, selected agent context, and workflow references to
  worker/task assignments and agent delegation.
- **Out of scope:** executing prompts, invoking an LLM, selecting a Copilot
  agent through a provider API, and supporting a second render target.

## 3. Non-goals

- The capability does not turn miez into a task runner or workflow engine.
- The target contract contains no prompt-only worker form; every worker is a
  Copilot custom agent.
- A task does not own a persona, duplicate a skill body, or silently select a
  worker.
- A task prompt does not depend on invoking one Copilot slash prompt from
  inside another prompt. Prompt files and skills are not treated as nested
  dispatch mechanisms.
- The capability does not prescribe a parameter language for reusable task
  inputs; the user's accompanying message is the initial input mechanism.
- A workflow declaration does not imply that miez can observe or guarantee
  Copilot's execution of each phase.

## 4. Requirements

### 4.1 Functional

#### `task/domain-roles-are-distinct`

The system shall treat workers, skills, tasks, and workflows as distinct
artifact roles with separate ownership and invocation semantics.

Acceptance:

- WHEN a team catalog is loaded THEN it exposes workers, skills, tasks, and
  workflows as separate collections.
- WHEN a task is rendered THEN its content does not replace the selected
  worker persona or inline the canonical skill bodies.

#### `worker/all-workers-render-as-agents`

The target worker contract shall render every worker as a GitHub Copilot custom
agent with its persona, effective skills, tools, and model configuration.

Acceptance:

- WHEN a worker is rendered THEN its managed output is a selectable Copilot
  custom agent.
- WHEN a worker source does not declare `kind: agent` THEN the target build
  rejects it before rendering.

#### `task/is-a-reusable-work-objective`

The system shall represent a task as a reusable work objective that is
independent of a worker's identity and assigned skill set.

Acceptance:

- WHEN the same task is invoked with two different workers THEN the task
  objective remains unchanged while the active worker context may differ.
- WHEN a task has no worker binding THEN it remains valid for use with any
  selectable agent worker.

#### `task/is-discoverable`

When a valid team is built, the system shall make every task artifact
discoverable by a stable task identifier and source reference.

Acceptance:

- WHEN a team contains valid task artifacts THEN the generated catalog exposes
  each task exactly once in deterministic order.
- WHEN a task identifier is duplicated or invalid THEN the build fails and
  identifies the task artifact.

#### `task/catalog-includes-task-relationships`

When a workflow references a worker/task assignment or agent delegation, the
system shall preserve those relationships in the generated catalog.

Acceptance:

- WHEN a workflow names a task assignment THEN the generated workflow entry
  retains the worker and task identifiers in phase order.
- WHEN a workflow names delegated worker agents THEN the generated workflow
  entry retains those worker identifiers in declared order.

#### `task/renders-as-a-copilot-prompt`

When a valid task is rendered for GitHub Copilot, the system shall produce a
manually invokable prompt that contains the task instructions and directs the
active worker to apply its persona and assigned skills.

Acceptance:

- WHEN a task is rendered THEN a Copilot user can discover and invoke one
  prompt for that task.
- WHEN a task prompt is rendered THEN it contains the task body without
  copying the selected worker's persona or the canonical skill bodies.

#### `task/selected-agent-context-is-preserved`

When a user invokes a task while a worker agent is selected, the system
shall preserve that worker as the active persona for the task.

Acceptance:

- WHEN a user selects agent A and invokes task T THEN Copilot receives T in
  agent A's conversation context.
- WHEN the same task is invoked after selecting agent B THEN the task remains
  T and the active persona is B.

#### `task/user-request-is-task-context`

When a user supplies a message alongside a task invocation, the system shall
make that message available as task-specific context without treating it as a
replacement for the worker persona or assigned skill guardrails.

Acceptance:

- WHEN a user adds repository, scope, or outcome details to a task invocation
  THEN the selected worker can use those details while retaining its persona
  and applicable skills.
- WHEN a task body and user message provide different levels of detail THEN
  the user message supplies the current context and the task body supplies the
  reusable objective.

#### `workflow/agent-delegation-is-explicit`

When a workflow intends to coordinate multiple workers automatically, the
system shall require explicit agent delegation or handoff relationships among
the participating workers.

Acceptance:

- WHEN a workflow declares delegated workers THEN every referenced worker is
  an existing custom agent.
- WHEN a workflow has no supported delegation or handoff relationship THEN it
  remains a coordination instruction and does not claim to invoke workers
  automatically.
- WHEN a workflow references a missing worker THEN validation fails before
  rendering.

#### `task/prompt-composition-is-not-nested`

The system shall render task prompts so that their execution does not depend on
invoking another prompt file or skill as a nested slash command.

Acceptance:

- WHEN a task prompt is rendered THEN it contains the task objective and
  relies on the currently selected agent or provider-supported agent metadata.
- WHEN a workflow requires another worker THEN it uses an explicit
  agent-delegation or handoff relationship instead of embedding a prompt
  invocation instruction.

#### `task/workflow-task-assignments-are-valid`

When a workflow declares task assignments, the system shall validate every
assignment against an existing worker and task and preserve the declared phase
order.

Acceptance:

- WHEN every assignment references existing artifacts THEN the workflow is
  valid and its assignments remain in declared order.
- WHEN an assignment references a missing worker or task THEN build fails and
  identifies the workflow phase and missing identifier.
- WHEN a phase contains multiple assignments THEN each worker/task pairing is
  independently addressable.

#### `task/prompt-identifiers-are-unique`

When task prompts and worker agents are rendered, the system shall reject
collisions in their user-invocable identifiers.

Acceptance:

- WHEN two rendered prompts would have the same invocation identifier THEN
  rendering fails before changing managed files.
- WHEN all identifiers are unique THEN every rendered task prompt and worker
  agent maps to exactly one source artifact.

### 4.2 Quality

#### `task/rendering-is-deterministic`

When the same team source and workspace state are rendered repeatedly, the
system shall produce byte-equivalent task prompts, agent metadata, and catalog
data.

Acceptance:

- WHEN the inputs are unchanged THEN repeated renders produce identical file
  content, paths, and ordering.

#### `task/rendering-is-traceable`

When a task prompt is rendered, the system shall make its source task
identifiable from the generated catalog and managed output plan.

Acceptance:

- WHEN an operator inspects the catalog or render plan THEN each task output
  can be traced to exactly one task source.

### 4.3 Constraints

#### `task/copilot-is-the-only-render-target`

The system shall render task artifacts only for the GitHub Copilot target until
another target has a separately defined contract.

Acceptance:

- WHEN the target is GitHub Copilot THEN task prompts are rendered.
- WHEN an unsupported target is selected THEN rendering fails explicitly.

#### `task/miez-does-not-execute-tasks`

The system shall not execute a task, invoke a worker, or claim completion of a
workflow phase as part of task rendering.

Acceptance:

- WHEN rendering succeeds THEN miez has only written validated artifacts and
  has not called an LLM or Copilot execution API.

### 4.4 Interface

#### `task/agent-invocation-is-select-then-prompt`

When a user wants to combine an agent worker with a reusable task, the system
shall support selecting the agent first and invoking the task prompt in that
agent context.

Acceptance:

- WHEN a user selects an agent and invokes a task prompt THEN the resulting
  interaction uses the selected agent persona and task objective.

#### `task/agent-invocation-is-the-composition-boundary`

When a user wants to combine a worker with a reusable task, the system shall
expose the worker as a custom agent and the task as a prompt that runs in the
selected agent context.

Acceptance:

- WHEN a user selects a worker agent and invokes a task prompt THEN the task
  uses that agent's persona, effective skills, and tools without invoking a
  second prompt file.

## 5. Open questions

- Whether a task prompt should optionally declare the provider's `agent`
  metadata for a worker-bound entry point, while keeping the canonical task
  reusable and worker-neutral.
- Whether the Copilot target should render tasks as `.prompt.md` files for the
  current Local agent, user-invocable skills for Agent Host, or both. The
  target must not assume prompt-file support where the selected host has
  deprecated it.
- Whether a future task input syntax should support named values, files, or
  interactive questions beyond the user's accompanying message.
- Whether miez should model provider-specific subagent permissions and
  handoffs directly in worker/workflow metadata or keep those relationships in
  the workflow Markdown body.