Set-Location 'D:\vibecoding\OFA'
Write-Output "=== Git Log ==="
git log --oneline -10
Write-Output ""
Write-Output "=== Git Status ==="
git status
Write-Output ""
Write-Output "=== Git Remote ==="
git remote -v