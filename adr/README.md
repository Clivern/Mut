# Architecture Decision Records (ADRs)

## What are ADRs?

Architecture Decision Records (ADRs) document important architectural decisions made in this project, including the context, the decision itself, and its consequences.

## Why ADRs?

- **Knowledge Preservation**: Capture the "why" behind decisions, not just the "what"
- **Onboarding**: Help new team members understand past decisions
- **Avoid Revisiting**: Prevent rehashing settled debates
- **Context**: Preserve the context that led to decisions
- **Accountability**: Clear ownership and timeline of decisions

## Structure

ADRs are stored in this directory using the following naming convention:

```
NNNN-title-with-dashes.md
```

Where:
- `NNNN` is a sequential number (0001, 0002, etc.)
- Title uses lowercase with dashes
- Files are markdown format

## ADR Lifecycle

Each ADR has a status:

- **Proposed**: Under discussion, not yet decided
- **Accepted**: Decision has been made and approved
- **Deprecated**: No longer applies but kept for historical context
- **Superseded**: Replaced by a newer ADR (link to the new one)
- **Rejected**: Proposed but decided against (still valuable to record why)

## When to Write an ADR

Write an ADR when making decisions about:

- Technology choices (database, frameworks, languages)
- Architecture patterns (microservices, monolith, event-driven)
- API design approaches
- Security and authentication strategies
- Data models and storage strategies
- Development and deployment processes
- Third-party integrations

## Template

Use the template in `0000-template.md` when creating new ADRs.

## Process

1. Copy `0000-template.md` to a new file with the next sequential number
2. Fill in the sections with relevant information
3. Set status to "Proposed"
4. Open for discussion (PR, team meeting, etc.)
5. Update status to "Accepted" or "Rejected" based on outcome
6. Update this index below

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0000](0000-template.md) | ADR Template | - | - |

## References

- [Michael Nygard's ADR template](https://github.com/joelparkerhenderson/architecture-decision-record)
- [ADR GitHub Organization](https://adr.github.io/)

