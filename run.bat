@echo off
set image=forum

echo Building Docker image...
docker build -t %image% .

echo Running Docker container...
docker run -it --rm -p 8080:8080 %image%
pause

