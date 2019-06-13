'use strict';

const loading = `
    <div class="preloader-wrapper xsmall active">
        <div class="spinner-layer spinner-green-only">
            <div class="circle-clipper left">
                <div class="circle"></div>
            </div><div class="gap-patch">
                <div class="circle"></div>
            </div><div class="circle-clipper right">
                <div class="circle"></div>
            </div>
        </div>
    </div>
`

const createBackupStatusListener = () => {
    console.log("createBackupStatusListener()");

    var evtSource = new EventSource(`http://localhost:8080/events?stream=status`);
    evtSource.onerror = (err) => {
        // console.error("!!! createBackupStatusListener():", err)
    }
    evtSource.onmessage = (evt) => {
        let data = JSON.parse(evt.data)
        let prog = data.done * 100 / data.total

        let progress = `
        <div class="progress">
            <div class="determinate" style="width: ${prog}%"></div>
        </div>
        <div id="backup-status-text">Uploaded ${data.done} of ${data.total} (${data.status})</div>`
        let el = document.getElementById("backup-status")
        el.innerHTML = progress
    }
}

const createProgressListener = () => {
    console.log("createProgressListener()");

    var evtSource = new EventSource(`http://localhost:8080/events?stream=files`);
    evtSource.onerror = (err) => {
        // console.error("!!! createProgressListener():", err)
    }
    evtSource.onmessage = (evt) => {
        let data = JSON.parse(evt.data)
        let id = data.id
        let el = getFreeListElement(id)
        if (el != null) {
            let progress = "";
            if (data.percent == 0) {
                progress = `<i class="tiny material-icons">file_upload</i> Uploading ...
                    <a href="#!" class="secondary-content">${loading}</a>`
            } else if (data.percent > 0 && data.percent < 100) {
                progress = `<i class="tiny material-icons">file_upload</i> Uploading ... ${data.percent}%
                    <a href="#!" class="secondary-content">${loading}</a>`
            } else {
                progress = `<i class="tiny material-icons">done</i> Synced
                <a href="#!" class="secondary-content"><i class="material-icons">cloud_done</i></a>`
                el.setAttribute("data-done", true);
            }

            el.setAttribute("data-id", id);
            el.innerHTML = `<span class="title">${data.file}</span><br>
                <span class="info">${progress}</span>`;
        }
    }
}

const getFreeListElement = (id) => {
    let list = document.getElementsByClassName("collection-item");

    for (var i = 0; i < list.length; i++) {
        if (list[i].innerHTML == "") {
            console.log("return element", i)
            return list[i]
        } else {
            let file = document.querySelector(`#item-${i}`);
            if (file != null && file.dataset != null && file.dataset.id == id) {
                return list[i]
            }
        }
    }

    // find element what is done
    for (var i = 0; i < list.length; i++) {
        let file = document.querySelector(`#item-${i}`);
        console.log("dataset:", file.dataset)
        if (file.dataset.done != null && file.dataset.done == "true") {
            console.log(`element ${i} is done, return it!`)
            list[i].setAttribute("data-done", false);
            list[i].setAttribute("data-id", "");
            list[i].innerHTML = ""
            return list[i]
        }
    }

    return null
}

document.addEventListener('DOMContentLoaded', createBackupStatusListener);
document.addEventListener('DOMContentLoaded', createProgressListener);
