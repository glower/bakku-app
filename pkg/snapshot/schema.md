### Snapshot storage:

buckets: ["local", "gdrive"]

"local": {
    "C:\Users\Igor\Pictures\Tokyo.jpg": {
        "hash": "foo",
        "size": 123
    },
    "C:\Users\Igor\Pictures\Yokohama.jpg": {
        "hash": "abc",
        "size": 225
    }
}
"gdrive": {
    "C:\Users\Igor\Pictures\Tokyo.jpg": {
        "hash": "bar",
        "size": "234",
    },
}

getActiveStorages():
    return ["local", "gdrive"]

createOrUpdate("C:\Users\Igor\Pictures\Yokohama.jpg", ["local", "gdrive"]):
    sendToStorages = []
    hash := getHashForFile("C:\Users\Igor\Pictures\Yokohama.jpg")
    for storage in ["local", "gdrive"]:
        data, err := getFileData(storage, "C:\Users\Igor\Pictures\Yokohama.jpg")
        if data.hash != hash {
            sendToStorages = append(sendToStorages, storage)
        } 