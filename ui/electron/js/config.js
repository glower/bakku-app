'use strict';

const { ipcRenderer } = require('electron');

const createSettingsButtonListener = () => {
    document.querySelector('#settings-btn').addEventListener('click', () => {
        // document.getElementById("content").innerHTML = content;
        ipcRenderer.send('settings-action', "ping");
    });
}

document.addEventListener('DOMContentLoaded', createSettingsButtonListener);
document.addEventListener("DOMContentLoaded", () => {
    var elem = document.querySelector('.tabs'); 
    var instance = M.Tabs.init(elem, {});
})