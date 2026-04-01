Set-Location 'D:\vibecoding\OFA'

# Git operations
git status --short
git add -A
git commit -m "feat(sdk): complete offline enhancement for Lite/IoT SDK

- Lite SDK: offline/p2p/constraint modules (3 files)
  - Lightweight OfflineManager with power save mode
  - BLE/WiFi P2P client optimized for wearables
  - Battery/privacy/security constraint checker

- IoT SDK: offline/p2p/constraint modules (3 files)
  - MQTT message cache with auto-replay
  - Edge computing rules support
  - P2P device discovery and mesh network
  - Power/time/safety/privacy constraints

- Updated SDK_ROADMAP.md

All 11 SDKs now have offline capability!

Co-Authored-By: Claude <noreply@anthropic.com>"
git push origin main

Write-Output "Done!"