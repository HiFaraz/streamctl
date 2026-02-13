#!/bin/bash
# Hook: TaskCompleted
# Logs completed tasks to the active streamctl workstream

set -e

INPUT=$(cat)
TASK=$(echo "$INPUT" | jq -r '.task_subject // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')

# Skip if no task subject
if [ -z "$TASK" ]; then
  exit 0
fi

# Detect project from cwd (use directory name)
PROJECT=$(basename "$CWD")

# Log to streamctl (silently fail if no active workstream)
streamctl log "$PROJECT" "Completed: $TASK" 2>/dev/null || true

exit 0
