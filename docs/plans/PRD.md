# Intent-Guided CLI Wrapper for Arbitrary Binaries  
## Progressive Disclosure + Contextual Composition Architecture

---

## 1. Overview

This document defines the architecture for a universal command-line wrapper capable of enhancing any binary with:

- Progressive disclosure (step-by-step guidance instead of static documentation)
- Context-aware next-step suggestions
- Workflow navigation
- Error recovery assistance
- Smart command composition (including piping and chaining)
- Future extensibility via local LLM reasoning and self-improvement mechanisms

The system operates **without modifying the underlying binaries**, acting as an intelligent execution layer that augments user interaction.

---

## 2. Core Objective

Transform any CLI tool from a static command executor into a **guided, adaptive workflow interface** that:

- Reduces reliance on external documentation
- Surfaces relevant next actions after each command
- Enables discovery of advanced usage patterns organically
- Supports power-user behaviors like composition and chaining
- Evolves over time through learning and optional model-based reasoning

---

## 3. High-Level Architecture

The system is composed of loosely coupled components to allow incremental evolution:

```
User → Wrapper CLI → Context Engine → Suggestion System → Policy Layer → Output
                         ↓                ↓
                  Execution Engine    Providers (Rules / Adapters / LLM)
                         ↓
                    Target Binary
```

---

## 4. Core Components

### 4.1 Invocation Layer (Wrapper CLI)

The entry point used by the user.

Responsibilities:
- Accept user commands
- Identify target binary
- Forward arguments transparently
- Enable/disable guidance modes
- Maintain session context

Example:
```
meta git commit -m "message"
```

This is translated to:
```
git commit -m "message"
```

---

### 4.2 Execution Engine

Responsible for running the actual binary.

Responsibilities:
- Execute command without modification
- Capture stdout, stderr, exit codes
- Preserve native behavior and performance
- Stream output to user in real time

Constraints:
- Must not alter command semantics
- Must not interfere with interactive tools

---

### 4.3 Execution Observer

Captures and structures the outcome of execution.

Extracts:
- Exit status (success / failure / partial)
- Error patterns
- Output characteristics (empty, verbose, structured)
- Timing and side effects (if detectable)

This feeds downstream context interpretation.

---

### 4.4 Context Engine

Builds a structured representation of the current state.

Inputs:
- Current command + arguments
- Execution result
- Session history (recent commands)
- Environment state (files, repo, config presence)
- Adapter metadata (if available)

Outputs:
- Inferred intent (e.g., setup, inspect, deploy, debug)
- Workflow stage
- Confidence levels
- Context envelope for suggestion generation

---

### 4.5 Adapter System (Tool Abstraction Layer)

Defines structured knowledge about specific binaries.

Supports multiple levels:

#### Level 0 — Generic
- No prior knowledge
- Uses help parsing and heuristics

#### Level 1 — Inferred
- Parses `--help`, man pages
- Builds command graph dynamically

#### Level 2 — Curated
- Explicit workflow definitions
- Intent mappings
- Error heuristics
- Known best practices

Adapter responsibilities:
- Describe command structure
- Define workflow transitions
- Map outputs to states
- Constrain valid suggestions

---

### 4.6 Suggestion System

Generates candidate next steps after each command.

#### Multi-Provider Model

Suggestions are produced by independent providers:

1. **Rule-Based Provider**
   - Static workflows from adapters
   - Known command sequences

2. **Structure-Based Provider**
   - Derived from command adjacency (help parsing)

3. **Outcome-Based Provider**
   - Reacts to success/failure states

4. **History-Based Provider (future)**
   - Learns from repeated usage patterns

5. **LLM-Based Provider (future)**
   - Contextual reasoning
   - Smart composition and chaining
   - Interpretation of ambiguous outputs

---

### 4.7 LLM-Based Suggestion Provider (Future Milestone)

A local LLM acts as a **bounded reasoning engine**, not a source of truth.

Capabilities:
- Infer user intent from incomplete context
- Suggest next commands when workflows are unclear
- Generate command compositions:
  - Pipes (`|`)
  - Chains (`&&`)
  - Filters and transformations
- Interpret complex or noisy stderr output
- Suggest advanced usage patterns

Example:
```
kubectl get pods | grep error | awk '{print $1}' | xargs kubectl logs
```

#### Constraints:
- Must not hallucinate commands
- All suggestions must be validated against:
  - Known commands
  - Parsed help structure
  - Adapter constraints
- Must operate within policy boundaries
- Must degrade gracefully if unavailable

---

### 4.8 Policy & Ranking Layer

Central decision-maker.

Responsibilities:
- Merge suggestions from all providers
- Rank by relevance and confidence
- Filter unsafe or invalid suggestions
- Limit output to a small set

Typical output:
- 1 primary suggestion
- 1–2 secondary suggestions
- Optional recovery path

Policy inputs:
- Context engine output
- Provider confidence scores
- Safety constraints
- User preferences

---

### 4.9 Output Renderer

Displays results to the user.

Structure:
1. Native command output
2. Minimal separator
3. Suggested next steps

Example:
```
✔ Plan successful

Next:
→ terraform apply

Also useful:
- terraform show
- terraform plan -out=tfplan

Recovery:
- terraform validate
```

Design goals:
- Concise
- Non-intrusive
- Readable in terminal
- No noise

---

## 5. Execution Flow

1. User invokes wrapper command
2. Wrapper resolves target binary
3. Context engine initializes session state
4. Execution engine runs command
5. Observer captures outcome
6. Context engine updates intent + stage
7. Suggestion providers generate candidates
8. Policy layer ranks and filters
9. Output renderer displays suggestions

---

## 6. Command Composition (Power User Feature)

A key differentiator is support for **smart composition discovery**.

### What this means

The system can suggest:
- Pipes between commands
- Chained operations
- Data transformations
- Multi-step workflows collapsed into one-liners

### Why this matters

Power users rely heavily on:
- `|`
- `&&`
- `xargs`
- `awk`, `sed`, `grep`

These are:
- Hard to discover
- Poorly documented in context
- Rarely suggested by traditional tools

### Role of LLM

The LLM excels at:
- Recognizing transformation intent
- Generating syntactically valid pipelines
- Combining tools across domains

### Validation Layer

All composed suggestions must be:
- syntactically valid
- composed of known commands
- aligned with detected output structure (if possible)

---

## 7. Modes of Operation

### Transparent Mode
- No suggestions
- Pure passthrough

### Guided Mode (default)
- Context-aware suggestions after execution

### Explain Mode
- Shows reasoning behind suggestions

### Quiet Mode
- Suppresses all guidance

### Learning Mode (future)
- Collects usage patterns locally

---

## 8. Data & State Model

### Session State
- Recent commands
- Inferred intent
- Workflow stage

### Static Knowledge
- Adapters
- Command graphs
- Error heuristics

### Observational Data (future)
- Command sequences
- Success/failure patterns
- Accepted vs ignored suggestions

### Learned Policy (future)
- Ranking weights
- Transition probabilities
- Suppression rules

---

## 9. Self-Improvement Architecture (Future)

The system should support a controlled feedback loop:

1. Observe usage patterns
2. Identify common transitions
3. Propose improvements
4. Validate changes
5. Apply via policy updates

Important:
- Learning must not directly modify execution behavior
- All updates must pass through policy validation
- Changes must be reversible

---

## 10. Safety Constraints

The wrapper must:

- Never alter underlying command behavior
- Never fabricate commands
- Never suggest destructive actions without explicit signals
- Always preserve exit codes and output integrity
- Validate all suggestions before display
- Maintain deterministic fallback behavior

---

## 11. Extensibility Strategy

The system is designed to evolve in stages:

### V1
- Rule-based suggestions
- Adapter system
- Post-execution hints

### V2
- Help parsing for unknown binaries

### V3
- Session-aware context modeling

### V4
- Local LLM for contextual reasoning and composition

### V5
- Self-improving policy layer

---

## 12. Key Design Principle

All intelligence (rules, adapters, LLM, learning) must remain:

- **decoupled from execution**
- **bounded by policy**
- **validated before exposure**

This ensures the system remains:
- predictable
- safe
- extensible
- compatible with any binary

---

## 13. Summary

This architecture introduces a universal wrapper that transforms any CLI into a guided, context-aware interface by:

- Observing execution
- Inferring intent
- Suggesting next steps
- Enabling discovery of advanced usage patterns
- Supporting future reasoning via local LLMs
- Allowing gradual evolution toward self-improving systems

It bridges the gap between raw command execution and intelligent workflow navigation without requiring changes to existing tools.

---