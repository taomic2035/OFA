Set-Location 'D:\vibecoding\OFA'

Write-Output "=== Step 1: Git Status ==="
git status --short 2>&1

Write-Output "`n=== Step 2: Adding files ==="
git add -A 2>&1

Write-Output "`n=== Step 3: Committing ==="
git commit -m "feat(sdk): add offline capability for C++/Go/Desktop SDK

- C++ SDK: offline/p2p/constraint modules (6 files)
  - LocalScheduler with retry mechanism
  - P2P client with TCP/UDP discovery
  - Constraint checker with default rules

- Go Agent SDK: offline/p2p/constraint modules (3 files)
  - OfflineManager with L1-L4 levels
  - P2PClient with peer discovery
  - ConstraintChecker with pattern matching

- Desktop SDK: offline/p2p/constraint modules (3 files)
  - Integrated with system tray agent
  - Auto-sync when online
  - Network monitoring

- Updated CMakeLists.txt and SDK_ROADMAP.md

Co-Authored-By: Claude <noreply@anthropic.com>" 2>&1

Write-Output "`n=== Step 4: Pushing to GitHub ==="
git push origin main 2>&1

Write-Output "`n=== Step 5: Verifying ==="
Write-Output "Remote main:"
git rev-parse origin/main 2>&1
Write-Output "Local main:"
git rev-parse HEAD 2>&1

Write-Output "`n=== Done ==="