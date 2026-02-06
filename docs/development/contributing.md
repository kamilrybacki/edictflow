# Contributing

Thank you for your interest in contributing to Edictflow! This guide will help you get started.

## Code of Conduct

Be respectful, inclusive, and constructive. We're building something together.

## How to Contribute

### Reporting Bugs

1. Check [existing issues](https://github.com/kamilrybacki/edictflow/issues)
2. Create a new issue with:
   - Clear title
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, versions)

### Suggesting Features

1. Check existing issues and discussions
2. Open a feature request issue with:
   - Clear description of the feature
   - Use case and motivation
   - Proposed implementation (optional)

### Contributing Code

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

See [Local Setup](local-setup.md) for detailed instructions.

Quick start:

```bash
git clone https://github.com/YOUR_USERNAME/edictflow.git
cd edictflow
task dev
```

## Pull Request Process

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bug-fix
```

Branch naming:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation
- `refactor/` - Code refactoring
- `test/` - Test additions

### 2. Make Changes

Follow the coding standards below.

### 3. Write Tests

- Add tests for new features
- Update tests for changed behavior
- Ensure all tests pass: `task test`

### 4. Update Documentation

- Update relevant documentation
- Add code comments where helpful
- Update CHANGELOG if applicable

### 5. Commit

Follow conventional commits:

```bash
git commit -m "feat: add enforcement mode validation"
git commit -m "fix: handle nil pointer in rule update"
git commit -m "docs: update API reference"
```

Types:

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Formatting, no code change
- `refactor` - Code change, no feature/fix
- `test` - Adding tests
- `chore` - Build, tools, etc.

### 6. Push and Create PR

```bash
git push -u origin feature/my-feature
```

Then create a PR on GitHub with:

- Clear title
- Description of changes
- Link to related issues
- Screenshots (if UI changes)

### 7. Review Process

- Maintainers will review your PR
- Address feedback promptly
- PRs require passing CI and approval

## Coding Standards

### Go

#### Formatting

```bash
# Format code
task fmt

# Run linter
task lint
```

#### Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful variable names
- Keep functions focused and small
- Add comments for exported functions

#### Example

```go
// CreateRule creates a new rule for the specified team.
// It validates the rule content and triggers before saving.
func (s *Service) CreateRule(ctx context.Context, rule *Rule) error {
    if err := validateRule(rule); err != nil {
        return fmt.Errorf("invalid rule: %w", err)
    }

    if err := s.repo.Create(ctx, rule); err != nil {
        return fmt.Errorf("failed to create rule: %w", err)
    }

    s.notifyRuleCreated(ctx, rule)
    return nil
}
```

### TypeScript

#### Formatting

```bash
cd web
npm run lint
npm run format
```

#### Style Guide

- Use TypeScript strict mode
- Prefer functional components
- Use proper typing (no `any`)
- Follow React hooks best practices

#### Example

```typescript
interface RuleCardProps {
  rule: Rule;
  onEdit: (id: string) => void;
}

export function RuleCard({ rule, onEdit }: RuleCardProps) {
  const handleClick = useCallback(() => {
    onEdit(rule.id);
  }, [rule.id, onEdit]);

  return (
    <Card onClick={handleClick}>
      <h3>{rule.name}</h3>
      <Badge>{rule.enforcement_mode}</Badge>
    </Card>
  );
}
```

### SQL Migrations

- Use sequential numbering
- Always provide up and down migrations
- Keep migrations atomic
- Test migrations on production-like data

```sql
-- 000015_add_rule_priority.up.sql
ALTER TABLE rules ADD COLUMN priority INTEGER NOT NULL DEFAULT 100;

-- 000015_add_rule_priority.down.sql
ALTER TABLE rules DROP COLUMN priority;
```

## Testing Requirements

### Coverage

- New features: Add unit tests
- Bug fixes: Add regression test
- Critical paths: Add integration tests

### Running Tests

```bash
# All tests
task test

# Unit only
task test:unit

# Integration
task test:integration

# E2E
task e2e
```

## Documentation

### When to Update

- New features
- Changed behavior
- New API endpoints
- Configuration changes

### Documentation Files

- `/docs/` - MkDocs documentation
- Code comments - For implementation details
- README.md - Project overview

### Building Docs Locally

```bash
pip install mkdocs-material
mkdocs serve
# Open http://localhost:8000
```

## Release Process

Maintainers handle releases, but good to know:

1. Update CHANGELOG.md
2. Update version numbers
3. Create release tag
4. GitHub Actions builds and publishes

## Getting Help

- [GitHub Issues](https://github.com/kamilrybacki/edictflow/issues) - Bug reports, features
- [GitHub Discussions](https://github.com/kamilrybacki/edictflow/discussions) - Questions, ideas

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors are recognized in:

- CONTRIBUTORS.md
- Release notes
- Documentation credits

Thank you for contributing to Edictflow!
