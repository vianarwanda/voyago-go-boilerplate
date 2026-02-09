# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records (ADRs) for the Voyago Core API project.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## ADR Format

Each ADR follows this structure:

1. **Title** - A short descriptive title
2. **Status** - Proposed, Accepted, Deprecated, or Superseded
3. **Context** - The issue motivating this decision
4. **Decision** - The change that we're proposing or have agreed to
5. **Consequences** - What becomes easier or harder as a result

## ADR Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](./001-monolithic-architecture.md) | Adoption of Modular Monolith Architecture | Accepted | 2026-02-07 |

## Creating a New ADR

1. Create a new file: `{number}-{short-title}.md`
2. Use the next sequential number
3. Follow the template structure
4. Update this README's index table
5. Submit for review

## ADR Lifecycle

- **Proposed** - Under discussion
- **Accepted** - Approved and implemented
- **Deprecated** - No longer relevant
- **Superseded** - Replaced by a newer ADR

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
