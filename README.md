# Bakku-app: windows/linux backub service

This is my fun project to get some practice in golang. The idea is to store changed files automatically to some cloud or whatever is emplemented as a storage provider. I use native windows/linux API to get notifications on file changes (using CGO). 

TODO:
[-] write storage plugin for local file system
[-] write storage plugin for S3
[-] write local configuration
[-] write REST endpoints for updating configuration
[-] scan and backup all files from the directory on the first run
[-] store localy state of backuped files for each storage provider
[-] on start compare states between local and remote storage
[-] need some sort of flat DB to store info about files 