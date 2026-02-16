# Project Checkpoint

Perform a comprehensive project checkpoint review with multi-perspective advice. The goal is awareness and learning through an **advisory council** of perspectives.

**Important**: This is an advisory process. The council provides suggestions; the user decides what to actually do. "Suggested Actions" and "Predictions" are the council's recommendations, not user commitments. Track them to calibrate the council's advice quality over time.

## Process

### Phase 1: Gather Context

#### 1.1 Find Last Checkpoint

Look for existing checkpoint files in `docs/checkpoints/*.md` (excluding INDEX.md). Find the most recent one (by date in filename) and extract:
- The `last_commit` field from the frontmatter
- **Suggested Actions** from that checkpoint (for calibration review)
- **Predictions** from that checkpoint (for verification)
- **Open Questions** (to check if answered)

If no checkpoint exists, this is the inaugural checkpoint—review from the first commit.

#### 1.2 Gather Automated Metrics

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

#### 1.3 Analyze Commits

Review all commits since last checkpoint:

```bash
git log --oneline --reverse <last_commit>..HEAD
git log --format="%h %ad %s" --date=short <last_commit>..HEAD
```

#### 1.4 Query Bug Tracker

Use streamctl MCP tools to get bug status:

```
# All bugs for project
bug_list(project="<project>")

# Open bugs (potential blockers)
bug_list(project="<project>", status="pending")
bug_list(project="<project>", status="in_progress")

# Critical issues requiring immediate attention
bug_list(project="<project>", severity="critical")
```

Extract:
- **Open bugs by severity**: critical / normal / low counts
- **Bugs closed since last checkpoint**: Progress indicator
- **Critical bugs**: Immediate blockers to highlight in Risks

### Phase 2: Analyze

#### 2.1 Review Previous Suggestions (Calibration)

Review the council's previous suggestions to calibrate advice quality:

1. **Suggested Actions**: For each suggestion, determine status:
   - **High/Medium priority**: Done / Partial / Not Done
   - **Low priority that wasn't done**: Move to Backlog (not a calibration failure)
2. **Predictions**: Compare predictions to reality. What was accurate? What surprised us?
3. **Open Questions**: Which were answered? Carry forward unanswered ones.

**Calibration score**: Only count High/Medium priority suggestions. Formula: `(Done + 0.5*Partial) / Total`. Low-priority items that get skipped aren't failures—they're correctly deprioritized.

**Backlog rule**: If a Low-priority suggestion wasn't done, move it to the Backlog section. Don't re-suggest it. The backlog is a parking lot, not a nag list.

#### 2.2 Assess Current State

Examine:
- What changed (files, architecture, tests)
- Velocity patterns (commits per day, burst vs steady)
- What shipped vs what's still in progress
- Test coverage and gaps
- Documentation currency (do docs match code?)
- Technical debt signals (TODOs, FIXMEs, commented code)
- Workstream status (check streamctl if available)
- **Bug status**: Open bugs, bugs closed this period, critical blockers
- **Surprises**: What didn't go as expected?

### Phase 3: Council Review

Each of the 11 council personas reviews the user's work and decisions.

**Using bug data**: Council members should reference bugs as evidence in their observations:
- **QA/Operator**: "3 test bugs closed (F010, F011, F023). No open critical bugs."
- **Security Engineer**: "No security-severity bugs. Nebula CA migration bug still open."
- **Tech Lead**: "Bug backlog at 0. All blockers resolved."
- **Engineer**: "Bug #294 revealed Patroni TTL misconfiguration—architectural insight."

**What counts as a "Decision"?** Examples:
- **Architecture**: single binary, Nebula-only networking, PostgreSQL over SQLite
- **Prioritization**: defer feature X, Docker-before-OVH, security gates production
- **Process**: TDD mandatory, agent tooling investment, parallel bootstrap clusters
- **Scope**: MVP definition, what's in v1 vs v1.1, bridge deferred

**Each persona provides 7 fields:**

| Field | Description |
|-------|-------------|
| Observation | Current state analysis from their perspective |
| Concern | Specific risk or gap they're watching |
| Action | One specific, actionable recommendation |
| Decisions Reviewed | List 3-4 key decisions from this period |
| Verdict | One of: STRONGLY AGREE, AGREE, MOSTLY AGREE, DISAGREE |
| Why not higher | What would earn more agreement (skip if STRONGLY AGREE) |
| Why not lower | What prevents disagreement (skip if DISAGREE) |

**The 11 Council Personas**:

| Persona | Focus Areas |
|---------|-------------|
| Tech Lead | Team effectiveness, process, velocity blockers, accountability |
| Engineer | Code quality, testing gaps, technical debt, implementation risks |
| Architect | System design, scaling, integration boundaries, technical decisions |
| Security Engineer | Dependencies, credentials, attack surface, vulnerabilities |
| QA / Operator | Test quality, debuggability, error messages, observability |
| Product Manager | Shipped vs planned, user impact, feature gaps, priorities |
| CTO | Technical vision, build vs buy, debt trajectory, platform investment |
| CFO | Cost efficiency, ROI, burn rate, commercial viability |
| CEO | Product-market fit, competitive position, sustainability, pivots |
| VP Engineering | Velocity, morale, process maturity, incident risk |
| CISO | Threat model, compliance, incident response, supply chain |

### Phase 4: Write the Checkpoint

Create a new checkpoint file at `docs/checkpoints/YYYY-MM-DD-<brief-description>.md`:

```markdown
---
date: YYYY-MM-DD
last_commit: <current HEAD commit hash (full)>
previous_checkpoint: <previous checkpoint filename or "none">
calibration_score: <X%>
overall_rating: <+/-X>
---

# Project Checkpoint: YYYY-MM-DD

## Summary
<2-3 sentence summary of this period>

## Calibration Review

### Previous Suggested Actions (High/Medium Only)
| Suggestion | Priority | Status | Notes |
|------------|----------|--------|-------|
| <suggestion> | High/Med | ✅ Done / ⏳ Partial / ❌ Not Done | <brief note> |

**Calibration Score**: X% (Y of Z High/Med suggestions followed)

### Moved to Backlog
| Item | Reason |
|------|--------|
| <low-priority item> | Low priority, skipped → backlog |

### Previous Predictions vs Reality
| Prediction | Confidence | Actual | Lesson |
|------------|------------|--------|--------|
| <what we predicted> | X% | <what happened> | <what we learned> |

### Questions Log
| Question | First Asked | Status | Resolution |
|----------|-------------|--------|------------|
| <question> | YYYY-MM-DD | ✅ Answered / ⏳ Open | <answer if resolved> |

## Period Reviewed
- **From**: <previous commit hash (short)> on <date>
- **To**: <current HEAD (short)> on <date>
- **Commits**: <count>
- **Days**: <count>

## Key Accomplishments
- <what shipped, bulleted list>

## Bug Status

### Closed This Period
| Bug ID | Description | Workstream | Severity |
|--------|-------------|------------|----------|
| #XXX | <bug text> | <workstream> | normal |

### Still Open
| Bug ID | Description | Workstream | Severity | Status |
|--------|-------------|------------|----------|--------|
| #XXX | <bug text> | <workstream> | critical | in_progress |

If no bugs: "No bugs tracked this period."

## Surprises
<What didn't go as expected? Why? What did we learn?>

## Metrics

| Metric | Previous | Current | Δ | Trend |
|--------|----------|---------|---|-------|
| Lines of Go | X | Y | +/-Z | ↑/↓/→ |
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
| Open bugs (critical) | X | Y | +/-Z | ↓ good |
| Open bugs (total) | X | Y | +/-Z | |
| Bugs closed this period | - | Z | - | |

**Trend symbols**: ↑ = increasing (good if tests, bad if debt), ↓ = decreasing, → = stable

## Council Review

Each persona provides observation, concern, action, and verdict on user's decisions.

### Tech Lead
**Observation**: <current state analysis>
**Concern**: <specific risk>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <STRONGLY AGREE / AGREE / MOSTLY AGREE / DISAGREE>
**Why not higher**: <what would earn more agreement>
**Why not lower**: <what prevents disagreement>

### Engineer
**Observation**: <code quality, patterns>
**Concern**: <technical debt or testing gap>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### Architect
**Observation**: <system design coherence>
**Concern**: <scaling or integration issue>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### Security Engineer
**Observation**: <security posture>
**Concern**: <attack surface or vulnerability>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### QA / Operator
**Observation**: <test quality, operational readiness>
**Concern**: <debuggability or observability gap>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### Product Manager
**Observation**: <feature completeness>
**Concern**: <priority or scope issue>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### CTO
**Observation**: <technical vision alignment>
**Concern**: <build vs buy or debt issue>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### CFO
**Observation**: <cost efficiency>
**Concern**: <ROI or burn rate issue>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### CEO
**Observation**: <strategic trajectory>
**Concern**: <market position or sustainability>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### VP Engineering
**Observation**: <velocity and process maturity>
**Concern**: <operational or incident risk>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### CISO
**Observation**: <threat model and compliance>
**Concern**: <incident response or supply chain>
**Action**: <one specific recommendation>
**Decisions Reviewed**:
1. <decision 1>
2. <decision 2>
3. <decision 3>
**Verdict**: <verdict>
**Why not higher**: <justification>
**Why not lower**: <justification>

### Verdict Summary

| Verdict | Count | Personas |
|---------|-------|----------|
| STRONGLY AGREE | X | <list> |
| AGREE | X | <list> |
| MOSTLY AGREE | X | <list> |
| DISAGREE | X | <list> |

**Consensus**: <one sentence on overall council sentiment>

## Performance Review (Committee)

The advisory council provides a unified performance assessment.

### Rating Scale
| Rating | Meaning |
|--------|---------|
| **+2** | Exceptional — went significantly beyond expectations, others should learn from this |
| **+1** | Strong — exceeded expectations, delivered quality work |
| **0** | Meets expectations — solid, competent work |
| **-1** | Below expectations — gaps exist, needs improvement |
| **-2** | Significant concerns — patterns that must change |

### Assessment

| Dimension | Rating | Evidence | Why Not Higher | Why Not Lower |
|-----------|--------|----------|----------------|---------------|
| **Technical Execution** | +/-X | <specific examples> | <what would make it higher> | <what prevents lower> |
| **Prioritization** | +/-X | <examples> | <gaps> | <strengths> |
| **Risk Management** | +/-X | <examples> | <gaps> | <strengths> |
| **Communication** | +/-X | <examples> | <gaps> | <strengths> |
| **Process Discipline** | +/-X | <examples> | <gaps> | <strengths> |

**Overall Rating**: +/-X
**Why not higher**: <what would push to next level>
**Why not lower**: <what's working well>

### What You Did Well
1. <specific accomplishment with evidence>
2. <specific accomplishment with evidence>

### Growth Areas
1. <specific area with actionable recommendation>
2. <specific area with actionable recommendation>

## Retrospective

**Written by the council on behalf of the user—what should you consider doing differently?**

<2-3 key lessons or process improvements to consider>

## Suggested Actions

**High priority = blocking or high-impact. Medium = important but not urgent. Low priority items go to Backlog.**

For technical issues, consider filing bugs via `bug_report()` instead of just listing suggestions. Bugs persist across sessions and can be tracked on workstreams.

### High Priority
| # | Suggestion | Rationale | File Bug? |
|---|------------|-----------|-----------|
| 1 | <specific suggestion> | <why critical/blocking> | ✅ / ❌ |

### Medium Priority
| # | Suggestion | Rationale | File Bug? |
|---|------------|-----------|-----------|
| 1 | <specific suggestion> | <why it matters> | ✅ / ❌ |

## Backlog

**Parking lot for low-priority items. Not tracked for calibration. Pull when priorities change.**

| Item | Added | Context |
|------|-------|---------|
| <item> | YYYY-MM-DD | <why it's here> |

## Predictions

**Specific, verifiable predictions to test next checkpoint.**

| Prediction | Confidence | Rationale |
|------------|------------|-----------|
| <prediction> | X% | <why you believe this> |

## Open Questions

**Questions needing resolution. Carry forward unanswered ones.**

| Question | First Asked | Priority | Context |
|----------|-------------|----------|---------|
| <question> | YYYY-MM-DD | High/Med/Low | <why it matters> |

## Risks to Watch

### From Bug Tracker

Query `bug_list(status="pending", severity="critical")` and `bug_list(status="in_progress", severity="critical")` to populate:

| Bug ID | Description | Severity | Workstream | Status |
|--------|-------------|----------|------------|--------|
| #XXX | <bug text> | critical | <workstream> | pending/in_progress |

If no critical bugs, note: "No critical bugs in tracker."

### Council-Identified

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|------------|
| <risk> | High/Med/Low | High/Med/Low | <what we're doing> |
```

### Phase 5: Report and Index

#### 5.1 Verbal Summary

After writing the checkpoint, provide a concise verbal summary:

1. **Calibration**: "X of Y high/med suggestions followed (Z%)"
2. **Predictions**: "X of Y predictions accurate. Key miss: ..."
3. **Story**: One sentence on what happened this period
4. **Council verdict**: "X of 11 personas AGREE or higher. Consensus: ..."
5. **Overall rating**: "+X because ... Not higher because ..."
6. **Must do**: High priority actions (if any)
7. **Watch**: Urgent risks (if any)

Keep the verbal summary short—the detailed analysis is in the checkpoint file.

#### 5.2 Update the Index

Add an entry to `docs/checkpoints/INDEX.md`:

```markdown
| Date | Checkpoint | Commits | Days | Calibration | Rating | Consensus |
|------|------------|---------|------|-------------|--------|-----------|
| YYYY-MM-DD | [name](filename.md) | X | Y | Z% | +/-X | <brief> |
```

**Columns**:
- **Commits**: Number of commits in this period
- **Days**: Calendar days in this period
- **Calibration**: % of previous High/Med suggestions followed
- **Rating**: Committee's overall performance rating
- **Consensus**: Brief council sentiment (e.g., "10/11 agree", "split on X")

## Philosophy

The checkpoint process is an **advisory council**, not a commitment tracker.

1. **Awareness**: What happened?
2. **Advice**: What does the council suggest?
3. **Calibration**: How good was previous advice?

The council provides multi-perspective suggestions. The user decides what to actually do. Tracking "calibration score" measures whether the council's advice aligned with what actually happened—not whether the user obeyed.

A checkpoint without calibration can't improve its advice. A checkpoint without predictions is just hindsight. The goal is: **suggest → observe → calibrate → improve advice quality**.
