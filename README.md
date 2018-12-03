# Bakku-app: windows/linux backub service

This is my fun project to get some practice in golang. The idea is to store changed files automatically to some cloud or whatever is emplemented as a storage provider. I use native windows/linux API to get notifications on file changes (using CGO).

This version is unstable and under development, don't have any tests here.

TODO (more like idea list):
- [x] add recursive directory notifications for linux
- [x] write storage plugin for local file system
- [ ] write storage plugin for S3
- [x] write storage plugin google drive
- [x] write local configuration (viper)
- [ ] write REST endpoints for updating configuration
- [ ] write GUI in electron for displaying progress
- [x] scan and backup all files from the directory on the first run
- [x] store local state of the backuped files for each storage provider (leveldb files are stored)
- [x] on start compare states between local and remote storage
- [x] need some sort of flat DB to store info about files (leveldb is used)
- [ ] implement filters like: store jpg to google drive and store raw to S3
- [x] add abstraction for leveldb
- [x] add abstraction for viper configs (config and config/storage)
- [ ] add tests after first stable version
    - [x] add tests for basic file change notification
    - [ ] add tests for snapshots
    - [ ] add tests for configs
    - [ ] add tests for storage
- [ ] add logic what to do with deleted files
- [ ] add logrus or something else for logging
- [ ] add a concept for error handling (notifications?)
