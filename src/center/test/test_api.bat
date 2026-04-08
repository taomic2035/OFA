@echo off
echo === OFA Center REST API Test ===
echo.

cd /d D:\vibecoding\OFA\src\center

echo [1] Starting Center...
start /B center.exe
timeout /t 3 /nobreak > nul

echo [2] Testing /health...
curl -s http://localhost:8080/health
echo.

echo [3] Testing /api/v1/tts/voices...
curl -s http://localhost:8080/api/v1/tts/voices
echo.

echo [4] Testing /api/v1/tts/synthesize...
curl -s -X POST http://localhost:8080/api/v1/tts/synthesize -H "Content-Type: application/json" -d "{\"text\":\"你好\",\"voice_id\":\"zh_female_tianmei\",\"format\":\"mp3\"}"
echo.

echo [5] Testing /api/v1/tts/identity/test-user/voice (PUT)...
curl -s -X PUT http://localhost:8080/api/v1/tts/identity/test-user/voice -H "Content-Type: application/json" -d "{\"voice_id\":\"zh_female_meilinvyou_uranus_bigtts\"}"
echo.

echo [6] Testing /api/v1/tts/identity/test-user/voice (GET)...
curl -s http://localhost:8080/api/v1/tts/identity/test-user/voice
echo.

echo [7] Stopping Center...
taskkill /F /IM center.exe > nul 2>&1

echo Done.
pause