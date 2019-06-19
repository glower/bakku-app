'use strict';

const { ipcRenderer } = require('electron');

// document.addEventListener('DOMContentLoaded', createSettingsButtonListener);
document.addEventListener("DOMContentLoaded", () => {
    var elem = document.querySelector('.tabs'); 
    var instance = M.Tabs.init(elem, {});


    document.querySelector('#settings-tab').addEventListener('click', () => {
        ipcRenderer.send('get-config-action');
        ipcRenderer.on('get-config-action-reply', (event, data) => {
            let json = JSON.parse(data.toString('utf8'));
            console.log(json);
            if (json.dirs_to_watch != null) {
                let list = document.getElementById("directories-list");
                list.innerHTML = ""; // reset the old list
                json.dirs_to_watch.forEach(dir => {
                    console.log("Dir:", dir.path)
                    let checked = ""
                    if (dir.active == true) {
                        checked = `checked="checked"`
                    }
                    list.innerHTML += `
                    <li class="collection-item">
                        <label>
                            <input type="checkbox" class="filled-in" ${checked} />
                            <span>${dir.path}</span>
                        </label>
                    </li>`
                });
            }
        });
    });
})