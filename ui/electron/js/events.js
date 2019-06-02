'use strict';

function setLoader() {
    alert("loading")
    console.log("setLoader()");
}

const createProgressListener = () => {
    // alert("createProgressListener")
    console.log("createProgressListener()");

    var evtSource = new EventSource(`http://localhost:8080/events?stream=files`);
    evtSource.onerror = (err) => {
        console.error("!!! createProgressListener():", err)
    }
    evtSource.onmessage = (evt) => {
        let data = JSON.parse(evt.data)
        // console.log("data:", data)
        let id = data.id
        let el = getFreeListElement(id) //document.getElementById(id);
        if (el != null) {
            el.setAttribute("data-id", id);
            el.innerHTML = `
                <span class="title">${data.file}</span><br>
                <span class="info">updating ... ${data.percent}%</span>
                `;
            //<div class="progress"><div class="determinate" style="width: ${data.percent}%"></div></div>
        }
    }
}

const getFreeListElement = (id) => {
    console.log("getFreeListElement")

    let list = document.getElementsByClassName("collection-item");

    for (var i = 0; i < list.length; i++) {
        if (list[i].innerHTML == "") {
            console.log("return element", i)
            return list[i]
        } else {
            let file = document.querySelector(`#item-${i}`);
            if (file.dataset.id == id) {
                return list[i]   
            }
            // console.log("element id:", file.dataset.id)
        }
    }
    return null
}

document.addEventListener('DOMContentLoaded', createProgressListener);
