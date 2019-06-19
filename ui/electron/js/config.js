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
                list.innerHTML += `
                <li class = "file-field input-field">
                    <div class = "btn">
                        <span>Add</span>
                        <input type="file"id="filepicker" webkitdirectory />
                    </div>                  
                    <div class="file-path-wrapper">
                        <input class="file-path validate" type="text" placeholder="Add directory"  />
                    </div>
                </li>`
            }
        });
        document.getElementById("filepicker").addEventListener("change", function(event) {
            let files = event.target.files;
            ipcRenderer.send('set-config-action', files);
            for (let i=0; i < files.length; i++) {
                console.log(files[i]);
            };
        }, false);
    });
})