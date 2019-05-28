1. To develop under Windows please download mingw-w64 from

https://sourceforge.net/projects/mingw-w64/files/latest/download

and use following settings:

Ver: 8.1.0
Arch: x86_64
Threads: posix
Exeptions: seh

Add C:\Program Files\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin to the PATH environment variable

2. Create an .bakkuapp directory (mkdir .bakkuapp) in the User drirectory, copy the config.yml.example to this directory and rename it to config.yml

3. Copy credentials.json with credentials for GoogleDrive file to .bakkuapp directory

> mkdir .bakkuapp
> notepad.exe .\.bakkuapp\credentials.json
> notepad.exe .\.bakkuapp\config.yml