#!/bin/bash
# Hook: TaskCompleted
# Logs completed tasks to the active streamctl workstream
#
# How it works:
# 1. Claude Code sends JSON to stdin with task details
# 2. We extract task_subject (what was completed) and cwd (current directory)
# 3. Project name = basename of cwd (e.g., /home/user/streamctl -> streamctl)
# 4. streamctl log finds the most recent in_progress workstream for that project
# 5. Logs "Completed: {task_subject}" to that workstream
# 6. Returns systemMessage so user sees confirmation

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

# Log to streamctl and capture output
RESULT=$(streamctl log "$PROJECT" "Completed: $TASK" 2>&1) || true

# Return systemMessage if logging succeeded
if echo "$RESULT" | grep -q "Logged to"; then
  WSNAME=$(echo "$RESULT" | sed 's/Logged to //')
  echo "{\"systemMessage\": \"📝 $WSNAME: $TASK\"}"
fi

exit 0
