Set-Location 'D:\vibecoding\OFA'

# 检查状态
$status = git status --porcelain
Write-Output "Status output:"
Write-Output $status

if ($status) {
    Write-Output "`nAdding all files..."
    git add -A

    Write-Output "`nCommitting..."
    git commit -m "feat(sdk): 离线增强完成 - Python/Android/iOS/Node.js/Rust SDK

- Python SDK: offline.py, p2p.py, constraint.py
- Android SDK: OfflineManager, LocalScheduler, P2PClient, ConstraintChecker
- iOS SDK: Offline/, P2P/, Constraint/ modules
- Node.js SDK: offline.ts, p2p.ts, constraint.ts
- Rust SDK: offline.rs, p2p.rs, constraint.rs
- Updated SDK_ROADMAP.md and ROADMAP.md

Co-Authored-By: Claude <noreply@anthropic.com>"

    Write-Output "`nPushing..."
    git push origin main 2>&1
} else {
    Write-Output "No changes to commit"
}

Write-Output "`nDone!"