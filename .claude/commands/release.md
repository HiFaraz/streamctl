# Release

Prepare a release by building, testing, and ensuring all documentation is up to date.

## Process

### 1. Build and Test

Run the build and test suite:

```bash
make build
make test
```

If tests fail, stop and report the failure. Do not proceed with documentation updates until tests pass.

### 2. Review Unreleased Changes

Check what's in the unreleased section of the changelog:

```bash
head -100 CHANGELOG.md
```

Also review recent commits since last release:

```bash
git log --oneline $(git describe --tags --abbrev=0 2>/dev/null || echo "HEAD~20")..HEAD
```

### 3. Update Project CLAUDE.md

Ensure `/home/faraz/streamctl/CLAUDE.md` accurately reflects:

- Current MCP tools (check `internal/mcp/server.go` for tool registrations)
- Current schema (check `internal/store/sqlite.go` for table definitions)
- Parameter documentation for all tools
- Any new features or workflows

Compare tools in code vs docs:

```bash
grep -o 'mcp.NewTool("[^"]*"' internal/mcp/server.go | sort
```

### 4. Update CHANGELOG.md

Ensure `CHANGELOG.md` has:

- All new features documented under `## Unreleased`
- Clear descriptions with usage examples where helpful
- Breaking changes clearly marked (if any)

The changelog should be ready to convert `## Unreleased` to a version number when tagging.

### 5. Update Global CLAUDE.md

Ensure `/home/faraz/.claude/CLAUDE.md` has:

- Updated workstream management section if streamctl tools changed
- Current usage examples for new features
- Any new workflows agents should follow

Focus on the "Workstream Management (streamctl)" section.

### 6. Verify Documentation Consistency

Check that the three documentation sources are consistent:

| Source | Purpose | Location |
|--------|---------|----------|
| CLAUDE.md (project) | Technical reference | `/home/faraz/streamctl/CLAUDE.md` |
| CHANGELOG.md | Release history | `/home/faraz/streamctl/CHANGELOG.md` |
| CLAUDE.md (global) | Usage guidance | `/home/faraz/.claude/CLAUDE.md` |

Common inconsistencies to check:
- Tool names match between all docs
- Parameter names/types match
- New features mentioned in changelog are also in CLAUDE.md files

### 7. Report

Summarize:

1. **Build status**: Pass/Fail
2. **Test status**: X tests passed
3. **Documentation updates made**: List any changes
4. **Ready for release**: Yes/No

If ready, the user can then:
- Bump version in `cmd/streamctl/main.go`
- Move `## Unreleased` to `## X.Y.Z` in CHANGELOG.md
- Commit, tag, and push

## Notes

- This command does NOT create git commits or tags
- This command does NOT bump version numbers
- This is a pre-release checklist, not an automated release
- The user decides when to actually tag and release
