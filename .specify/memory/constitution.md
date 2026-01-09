<!--
Sync Impact Report - Constitution Update
=========================================
Version: None → v1.0.0 (Initial ratification)
Change Type: MAJOR (Initial constitution establishment)
Date: 2026-01-09

Principles Established:
- I. Domain-Driven Architecture (NEW)
- II. CQRS Pattern (NEW)
- III. Infrastructure Adapters (NEW)
- IV. Testing & Quality Gates (NEW)
- V. Observability & Performance (NEW)

Sections Added:
- Technical Standards (NEW)
- Development Workflow (NEW)
- Governance (NEW)

Templates Status:
✅ .specify/templates/plan-template.md - Constitution Check section will be updated
✅ .specify/templates/spec-template.md - Requirements alignment validated (no changes needed)
✅ .specify/templates/tasks-template.md - Task categorization will be updated
✅ .specify/templates/agent-file-template.md - Template structure validated (no changes needed)
✅ .specify/templates/checklist-template.md - Validated (no changes needed)

Follow-up Actions:
- Update plan-template.md with concrete Constitution Check gates
- Update tasks-template.md with Foundational phase constitution alignment

Notes:
- Initial constitution based on project history (commits 2026-01-07 through 2026-01-09)
- Principles derived from ARCHITECTURE_DECISIONS.md and DEVELOPMENT_WORKFLOW.md
- Technology stack reflects cleared v2.0 rebuild requirements
- Testing philosophy: STRONGLY RECOMMENDED but not blocking for rapid iteration
-->

# Refract Constitution

## Core Principles

### I. Domain-Driven Architecture

All business logic MUST reside in the domain layer (`internal/domain/`) with zero dependencies on frameworks, HTTP, databases, or external libraries.

**Requirements**:
- Domain layer contains only pure business logic: entities, value objects, domain errors, and interface definitions
- Domain code MUST be testable without mocks (pure functions, value object validation)
- Infrastructure concerns (HTTP, database, cache) MUST NOT leak into domain layer
- Domain interfaces define contracts; infrastructure provides implementations

**Rationale**: Isolating business logic ensures testability, framework independence, and single source of truth for business rules. This enables the same logic to be used across multiple interfaces (HTTP, CLI, gRPC) without duplication.

### II. CQRS Pattern

Write operations (Commands) and read operations (Queries) MUST be separated into distinct handlers in the application layer.

**Requirements**:
- **Commands**: Located in `internal/application/commands/`, handle state changes, enforce business rules, validate inputs, trigger side effects
- **Queries**: Located in `internal/application/queries/`, fetch and return data, no state changes, optimized for read performance
- Each command/query MUST have a dedicated handler
- Handlers orchestrate domain logic but contain no business rules themselves

**Rationale**: Separating read and write concerns simplifies code reasoning, enables independent optimization of each path, and makes testing more focused (commands test business rules, queries test data retrieval).

### III. Infrastructure Adapters

All external dependencies (databases, HTTP servers, caches, third-party APIs) MUST be implemented as adapters in the infrastructure layer (`internal/infrastructure/`).

**Requirements**:
- Adapters implement domain interfaces (Repository, Generator, Validator patterns)
- HTTP layer maps requests to domain commands/queries via DTOs
- Database layer uses type-safe query builders (e.g., SQLc) - no raw SQL in handlers
- Configuration injected via dependency injection in `cmd/*/main.go` or `internal/app/`

**Rationale**: Adapter pattern enables swapping implementations (PostgreSQL → MySQL, HTTP → gRPC) without touching business logic. Clear boundaries reduce coupling and improve maintainability.

### IV. Testing & Quality Gates

All service boundaries (HTTP endpoints, database contracts, inter-service communication) MUST have integration or contract tests before production deployment.

**Requirements**:
- **Contract tests**: Verify API contracts, request/response schemas, error handling (`tests/contract/`)
- **Integration tests**: Test end-to-end flows including database, cache, external services (`tests/integration/`)
- **Unit tests**: Recommended for complex domain logic (`tests/unit/`)
- Tests MUST pass before merging feature branches
- Breaking changes require migration plan and compatibility layer

**Testing Philosophy**: STRONGLY RECOMMENDED but not blocking for rapid iteration. TDD encouraged for complex business logic.

**Rationale**: Testing service boundaries catches breaking changes early, documents expected behavior, and enables confident refactoring. Focus on integration over exhaustive unit coverage.

### V. Observability & Performance

All services MUST implement structured logging, health checks, and performance monitoring. Critical paths MUST use multi-tier caching.

**Requirements**:
- **Structured Logging**: JSON format with context (request IDs, user IDs, timestamps)
- **Health Checks**: `/health` endpoints report service status, dependency health (database, cache)
- **Caching Strategy**: Multi-tier (in-memory L1, Redis/Valkey L2) for read-heavy operations
- **Performance Targets**: Define per-service (e.g., API <100ms p95, Redirector <10ms p95)
- **Monitoring**: Expose metrics (request counts, latencies, error rates) for observability tooling

**Rationale**: Structured logs enable debugging in production; health checks support automated recovery; caching handles scale; metrics inform optimization efforts.

## Technical Standards

### Technology Stack

**Current Stack** (as of v2.0 rebuild):
- **Languages**: Go (API services), Rust (high-performance services like redirector)
- **Database**: PostgreSQL with migration framework (e.g., goose, sqlx-migrate)
- **Cache**: Redis or Valkey for multi-tier caching
- **Containerization**: Docker + Docker Compose for local dev, production-ready Dockerfiles
- **Query Layer**: Type-safe query builders (SQLc for Go, sqlx for Rust)

**Constraints**:
- New languages/frameworks require justification in Complexity Tracking (plan.md)
- Database changes require migration scripts + rollback plan
- All services MUST be containerized

### Dependency Management

- **Go**: Use Go modules (`go.mod`), pin major versions, audit dependencies quarterly
- **Rust**: Use Cargo with lock files (`Cargo.lock` committed), specify exact versions for production
- **Updates**: Security patches within 2 weeks, feature updates reviewed for breaking changes

### Database Migrations

- **Format**: SQL-based migrations numbered sequentially (e.g., `00001_create_urls_table.sql`)
- **Testing**: Migrations tested on staging before production
- **Rollback**: All migrations MUST include DOWN/rollback script
- **Schema Changes**: Breaking changes require blue-green deployment or compatibility layer

### Deployment

- **Local Development**: `just up` starts all services via Docker Compose
- **Environment Variables**: Defined in `.env.example`, never commit `.env`
- **Secrets**: Use secret management (env vars for local, vault/secrets manager for prod)
- **Ports**: Document all exposed ports in `docker-compose.yml` and README

## Development Workflow

### Branch Strategy

- **Main Branch**: `main` is always deployable, protected from direct pushes
- **Feature Branches**: `###-feature-name` format (e.g., `13-feature-build-redirector-service`)
- **Hotfixes**: `hotfix-description` for urgent production fixes

### Feature Development Process

1. **Specification**: Create feature spec in `specs/###-feature/spec.md` using spec template
2. **Planning**: Generate plan with `/speckit.plan` command → `specs/###-feature/plan.md`
3. **Constitution Check**: Validate compliance in plan.md before Phase 0 research
4. **Tasks**: Generate tasks with `/speckit.tasks` command → `specs/###-feature/tasks.md`
5. **Implementation**: Follow task order, commit per task or logical group
6. **Review**: PR against `main` with spec reference, constitution compliance check
7. **Merge**: Squash or merge commit with conventional commit message

### Code Review Requirements

- **Reviewers**: At least one reviewer for non-hotfix PRs
- **Checklist**:
  - [ ] Constitution compliance verified (domain layer isolation, CQRS, adapters)
  - [ ] Tests added/updated for service boundaries
  - [ ] Migrations tested (if applicable)
  - [ ] Structured logging added for new operations
  - [ ] Documentation updated (README, quickstart.md, API contracts)
  - [ ] No secrets in code or commits
- **Complexity Justification**: If plan.md has Complexity Tracking violations, verify justifications in PR description

### Just Commands (Consistency)

All services MUST provide `justfile` with standardized commands:
- `just up` / `just down` - Start/stop services
- `just logs` / `just <service>-logs` - View logs
- `just test` - Run test suite
- `just fmt` - Format code
- `just rebuild` - Rebuild and restart
- `just db-shell` - Database access

## Governance

### Amendment Process

1. **Proposal**: Document proposed change in GitHub issue or discussion
2. **Review**: Discuss rationale, impact on existing code, migration plan
3. **Approval**: Requires maintainer approval
4. **Update**: Amend constitution with version bump, update templates, propagate changes
5. **Migration**: If existing code violates new principle, create migration tasks

### Versioning Policy

Constitution follows **Semantic Versioning**:
- **MAJOR** (X.0.0): Backward-incompatible governance changes, principle removals or redefinitions
- **MINOR** (1.X.0): New principles/sections added, materially expanded guidance
- **PATCH** (1.0.X): Clarifications, wording fixes, typo corrections, non-semantic refinements

### Compliance Review

- **Pre-Implementation**: Constitution Check in plan.md (before Phase 0 research, re-check after Phase 1 design)
- **Code Review**: PRs must verify alignment with Core Principles
- **Retrospectives**: Quarterly review of constitution effectiveness, propose amendments if needed

### Complexity Justification

When a feature violates a principle (e.g., adds a 4th service, introduces new framework):
- Document in Complexity Tracking table (plan.md)
- Justify why needed and why simpler alternative insufficient
- Reviewer verifies justification is sound
- If pattern repeats, consider amending constitution

### Development Guidance

For runtime development guidance aligned with this constitution, refer to:
- `.specify/templates/plan-template.md` - Feature planning structure
- `.specify/templates/spec-template.md` - Specification format
- `.specify/templates/tasks-template.md` - Task breakdown approach
- Agent-specific guidance (if generated): `.specify/agent-guidance.md`

**Version**: v1.0.0 | **Ratified**: 2026-01-09 | **Last Amended**: 2026-01-09
