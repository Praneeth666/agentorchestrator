---
description: Stages all files, generates a conventional commit message, and pushes to the current branch.
allowed-tools: [Bash]
---

## Objective
Stage all modified files, write an accurate commit message based on the actual changes, and push the code to the remote repository.

## Execution Plan
1. **Identify the Branch:** Check the current active git branch using `!git branch --show-current`.
2. **Check for Changes:** Verify there are uncommitted changes using `!git status --porcelain`. If everything is clean, stop and inform the user.
3. **Stage Files:** Run `!git add -A`.
4. **Generate Commit Message:** Analyze the exact changes using `!git diff --cached`. Write a concise, professional commit message following conventional commit standards (e.g., `feat: add database configs` or `fix: resolve pointer panic`).
5. **Commit:** Run `git commit -m "[Your Generated Message]"` using the Bash tool.
6. **Push:** Run `git push origin [current-branch]` using the Bash tool.

Confirm the commit message and push status clearly to the user upon completion.
