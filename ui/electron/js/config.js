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
            if (json.directories != null) {
                let list = document.getElementById("directories-list");
                list.innerHTML = ""; // reset the old list
                json.directories.forEach(dir => {
                    console.log("Dir:", dir)
                    list.innerHTML += `
                    <li class="collection-item">
                        <label>
                            <input type="checkbox" class="filled-in" checked="checked" />
                            <span>${dir}</span>
                        </label>
                    </li>`
                });
            }
        });
    });
})