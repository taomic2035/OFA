@echo off
cd /d D:\vibecoding\OFA

echo Checking git status... > D:\vibecoding\OFA\git_status.txt
git status >> D:\vibecoding\OFA\git_status.txt 2>&1

echo. >> D:\vibecoding\OFA\git_status.txt
echo Git remote: >> D:\vibecoding\OFA\git_status.txt
git remote -v >> D:\vibecoding\OFA\git_status.txt 2>&1

echo. >> D:\vibecoding\OFA\git_status.txt
echo Git log: >> D:\vibecoding\OFA\git_status.txt
git log --oneline -3 >> D:\vibecoding\OFA\git_status.txt 2>&1

echo. >> D:\vibecoding\OFA\git_status.txt
echo Pushing to origin... >> D:\vibecoding\OFA\git_status.txt
git push -u origin main 2>&1 >> D:\vibecoding\OFA\git_status.txt

echo. >> D:\vibecoding\OFA\git_status.txt
echo Done. >> D:\vibecoding\OFA\git_status.txt

echo Output written to D:\vibecoding\OFA\git_status.txt
type D:\vibecoding\OFA\git_status.txt