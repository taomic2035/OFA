@echo off
cd /d D:\vibecoding\OFA\src\center
D:\Go\go\bin\go.exe build -o center.exe ./cmd/center
if exist center.exe (
    echo Build successful
    start /b center.exe
    timeout /t 5 /nobreak > nul

    echo Testing health endpoint...
    powershell -Command "(Invoke-WebRequest -Uri 'http://localhost:8080/health' -UseBasicParsing).Content"

    echo Testing TTS voices endpoint...
    powershell -Command "(Invoke-WebRequest -Uri 'http://localhost:8080/api/v1/tts/voices' -UseBasicParsing).Content"

    echo Testing TTS synthesize endpoint...
    powershell -Command "$body = '{\"text\": \"你好\", \"voice_id\": \"zh_female_tianmei\", \"format\": \"mp3\"}'; (Invoke-WebRequest -Uri 'http://localhost:8080/api/v1/tts/synthesize' -Method POST -ContentType 'application/json' -Body $body -UseBasicParsing).Content"
) else (
    echo Build failed
)