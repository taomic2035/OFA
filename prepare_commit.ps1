Set-Location 'D:\vibecoding\OFA'

Write-Output "=== Checking Git Status ==="
$status = git status --porcelain 2>&1
Write-Output $status

if (-not $status) {
    Write-Output "No changes to commit"
    exit 0
}

Write-Output "`n=== Files to be committed ==="
git status --short 2>&1

Write-Output "`n=== Ready to commit ==="
Write-Output "Commit message:"
Write-Output @"
feat(sdk): add offline capability for C++/Go/Desktop SDK

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

Co-Authored-By: Claude <noreply@anthropic.com>
"@

Write-Output "`nPress any key to continue, Ctrl+C to cancel..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

Write-Output "`nAdding files..."
git add -A 2>&1

Write-Output "`nCommitting..."
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

Write-Output "`n=== Commit Result ==="
git log -1 --oneline 2>&1

Write-Output "`n=== Ready to push ==="
Write-Output "Run 'git push origin main' to push to GitHub"