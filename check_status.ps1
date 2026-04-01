Set-Location 'D:\vibecoding\OFA'

Write-Output "=== Git Status ==="
git status --porcelain

Write-Output "`n=== Git Diff Summary ==="
git diff --stat

Write-Output "`n=== New Files ==="
git status --porcelain | Where-Object { $_ -match '^\?\?' }

Write-Output "`n=== Modified Files ==="
git status --porcelain | Where-Object { $_ -match '^[MA]' }