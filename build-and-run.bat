ECHO OFF
ECHO Building bakku-app for Windows ...
ECHO Make sure to install docker
docker build -t bakkuapp-win -f Dockerfile.windows .
docker run --name bakkuapp-win bakkuapp-win:latest
docker cp bakkuapp-win:/bin/bakku-app.exe bin/
docker rm bakkuapp-win
echo
echo Starting the app
echo
.\bin\bakku-app.exe
