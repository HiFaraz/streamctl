# Project Checkpoint

Perform a comprehensive project checkpoint review with accountability tracking, predictions, and multi-perspective advice. The goal is not just awareness but **learning**: predict, execute, verify, improve.

## Process

### 1. Find Last Checkpoint

Look for existing checkpoint files in `docs/checkpoints/*.md` (excluding INDEX.md). Find the most recent one (by date in filename) and extract:
- The `last_commit` field from the frontmatter
- **Committed Actions** from that checkpoint (for accountability review)
- **Predictions** from that checkpoint (for verification)
- **Open Questions** (to check if answered)

If no checkpoint exists, this is the inaugural checkpoint—review from the first commit.

### 2. Review Previous Commitments (Accountability Loop)

**This is critical.** Before analyzing new progress, review what was committed last checkpoint:

1. **Committed Actions**: For each action, determine status (Done/Partial/Not Started/Dropped)
2. **Predictions**: Compare predictions to reality. What was accurate? What surprised us?
3. **Open Questions**: Which were answered? Carry forward unanswered ones.

Calculate an accountability score: `(Done + Partial*0.5) / Total Committed Actions`

### 3. Gather Automated Metrics

Run these commands to gather objective metrics:

```bash
# Lines of Go code
find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1

# Test function count
grep -r "func Test" --include="*_test.go" | wc -l

# Test density (tests per 1K lines)
# = (test_count * 1000) / lines_of_go

# Static analysis issues
staticcheck ./... 2>&1 | wc -l

# TODOs and FIXMEs
grep -rn "TODO\|FIXME" --include="*.go" | wc -l

# Cyclomatic complexity (if gocyclo available)
gocyclo -avg . 2>/dev/null || echo "gocyclo not installed"

# Binary size (if binary exists)
ls -lh ./streamctl 2>/dev/null | awk '{print $5}' || echo "binary not built"

# Build time
time make 2>&1 | grep real || echo "build timing unavailable"

# Dependency vulnerabilities (if govulncheck available)
govulncheck ./... 2>/dev/null | grep -c "Vulnerability" || echo "0 or govulncheck not installed"
```

### 4. Analyze Progress Since Last Checkpoint

Review all commits:

```bash
git log --oneline --reverse <last_commit>..HEAD
git log --format="%h %ad %s" --date=short <last_commit>..HEAD
```

Examine:
- What changed (files, architecture, tests)
- Velocity patterns (commits per day, burst vs steady)
- What shipped vs what's still in progress
- **Surprises**: What didn't go as expected?

### 5. Assess Current State

Examine:
- Test coverage and gaps
- Documentation currency (do docs match code?)
- Technical debt signals (TODOs, FIXMEs, commented code)
- Workstream status (check streamctl if available)
- Open risks or blockers noted in code/docs

### 6. Provide Multi-Perspective Advice

Structure analysis as advice from different roles. Each perspective should include **one specific, actionable recommendation**.

**Tech Lead Perspective**
- Team/agent effectiveness and coordination
- Process improvements
- What's blocking velocity?
- Accountability: are we doing what we say we'll do?

**Engineer Perspective**
- Code quality and patterns
- Testing gaps that worry you
- Technical debt to address
- Implementation risks

**Architect Perspective**
- System design coherence
- Scaling concerns
- Integration points and boundaries
- Technical decisions that need revisiting

**Security Engineer Perspective**
- New dependencies added? Vetted?
- Credential/secret handling changes?
- Attack surface changes?
- Known vulnerabilities in dependencies?

**QA / Operator Perspective**
- Test quality (not just quantity): flaky tests? edge cases?
- Can this be debugged at 3am? Runbooks updated?
- Error messages actionable?
- Observability adequate?

**Product Manager Perspective**
- What's actually shipped vs planned?
- User-facing impact of recent work
- Feature completeness gaps
- Priority alignment

**Executive Perspective**
- Big picture trajectory
- Resource allocation (time/focus)
- Risk exposure
- Strategic questions to answer

### 7. Write the Checkpoint

Create a new checkpoint file at `docs/checkpoints/YYYY-MM-DD-<brief-description>.md`:

```markdown
---
date: YYYY-MM-DD
last_commit: <current HEAD commit hash (full)>
previous_checkpoint: <previous checkpoint filename or "none">
accountability_score: <X% - calculated from previous commitments>
---

# Project Checkpoint: YYYY-MM-DD

## Summary
<2-3 sentence summary of this period>

## Accountability Review

### Previous Committed Actions
| Action | Status | Notes |
|--------|--------|-------|
| <action from previous checkpoint> | ✅ Done / ⏳ Partial / ❌ Not Started / 🗑️ Dropped | <brief note> |

**Accountability Score**: X% (Y of Z actions completed)

### Previous Predictions vs Reality
| Prediction | Confidence | Actual | Lesson |
|------------|------------|--------|--------|
| <what we predicted> | X% | <what happened> | <what we learned> |

### Questions Log
| Question | First Asked | Status | Resolution |
|----------|-------------|--------|------------|
| <question> | YYYY-MM-DD | ✅ Answered / ⏳ Open | <answer if resolved> |

## Period Reviewed
- **From**: <previous commit or "inception">
- **To**: <current HEAD>
- **Commits**: <count>
- **Days**: <count>

## Key Accomplishments
<Bulleted list of what shipped>

## Surprises
<What didn't go as expected? Why?>

## Metrics

| Metric | Previous | Current | Δ | Trend |
|--------|----------|---------|---|-------|
| Lines of Go | X | Y | +/-Z | ▁▂▄ or ↑↓→ |
| Test functions | X | Y | +/-Z | |
| Test density (per 1K LoC) | X | Y | +/-Z | |
| staticcheck issues | X | Y | +/-Z | |
| TODOs/FIXMEs | X | Y | +/-Z | |
| Cyclomatic complexity (avg) | X | Y | +/-Z | |
| Binary size | X | Y | +/-Z | |
| Build time | X | Y | +/-Z | |
| Dependency vulns | X | Y | +/-Z | |
| Workstreams done | X | Y | +/-Z | |
| Workstreams pending | X | Y | +/-Z | |

## Advice

### Tech Lead
<advice with one specific action>

### Engineer
<advice with one specific action>

### Architect
<advice with one specific action>

### Security Engineer
<advice with one specific action>

### QA / Operator
<advice with one specific action>

### Product Manager
<advice with one specific action>

### Executive
<advice with one specific action>

## Retrospective
<What would you do differently this period? Key lessons learned?>

## Committed Actions
**These will be reviewed next checkpoint.** Link to workstreams where possible.

| # | Action | Owner/Workstream | Target |
|---|--------|------------------|--------|
| 1 | <specific action> | <streamctl workstream or "manual"> | <date or "next checkpoint"> |
| 2 | ... | | |

## Predictions
**These will be verified next checkpoint.**

| Prediction | Confidence | Rationale |
|------------|------------|-----------|
| <specific, verifiable prediction> | X% | <why you believe this> |

## Open Questions
<Questions that need resolution. Carry forward unanswered ones from previous checkpoint.>

| Question | Priority | Context |
|----------|----------|---------|
| <question> | High/Med/Low | <why it matters> |

## Risks to Watch

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|------------|
| <risk> | High/Med/Low | High/Med/Low | <what we're doing about it> |
```

### 8. Report to User

After writing the checkpoint, provide a concise verbal summary:

1. **Accountability headline**: "We completed X of Y committed actions (Z%)"
2. **Prediction accuracy**: "Our predictions were X% accurate. Key miss: ..."
3. **Story of this period**: One sentence on what happened
4. **Top 3 insights** (one must be a surprise)
5. **Top 3 committed actions** for next period
6. **Urgent risks** (if any)

Keep the verbal summary short—the detailed analysis is in the checkpoint file.

### 9. Update the Index

Add an entry to `docs/checkpoints/INDEX.md`:

```markdown
| Date | Checkpoint | Period | Accountability | Key Insight |
|------|------------|--------|----------------|-------------|
| YYYY-MM-DD | [name](filename.md) | X→Y | Z% | <insight> |
```

## Philosophy

The checkpoint process serves three purposes:

1. **Awareness**: What happened?
2. **Accountability**: Did we do what we said?
3. **Learning**: What will we do better?

A checkpoint without accountability is just a status report. A checkpoint without predictions is just hindsight. The goal is to close the loop: **predict → execute → verify → improve**.
