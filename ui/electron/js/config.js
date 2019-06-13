'use strict';

const { ipcRenderer } = require('electron');

// document.addEventListener('DOMContentLoaded', createSettingsButtonListener);
document.addEventListener("DOMContentLoaded", () => {
    var elem = document.querySelector('.tabs'); 
    var instance = M.Tabs.init(elem, {});


    document.querySelector('#settings-tab').addEventListener('click', () => {
        console.log("TAB: [settings] call settings endpoint!");
        let r = ipcRenderer.send('get-config-action');
        ipcRenderer.on('reply', (event, data) => {
            // let json = JSON.stringify(data);
            // console.log(bufferOriginal.toString('utf8'));
            console.log("data", data.toString('utf8'));
        });
    });
})