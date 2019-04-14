package bakkuapp

import (
	// autoimport for all implemented backup storages
	_ "github.com/glower/bakku-app/pkg/backup/fake"
	_ "github.com/glower/bakku-app/pkg/backup/gdrive"
	_ "github.com/glower/bakku-app/pkg/backup/local"
)
