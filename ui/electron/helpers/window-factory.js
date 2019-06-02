'use strict';

const { BrowserWindow } = require('electron');
const { APP_NAME, MAIN_WINDOW_WIDTH, MAIN_WINDOW_HEIGHT, UPDATER_WINDOW_HEIGHT, UPDATER_WINDOW_WIDTH } = require('./constants');

function getMain() {
    return new BrowserWindow({
        skipTaskbar: true,
        width: MAIN_WINDOW_WIDTH,
        height: MAIN_WINDOW_HEIGHT,
        resizable: true,
        useContentSize: false,
        movable: false,
        minimizable: false,
        maximizable: false,
        closable: false,
        alwaysOnTop: true,
        fullscreenable: false,
        title: APP_NAME,
        show: false,
        frame: false,
    });
}


exports.get = function (type, options) {
    const windows = {
        'main': getMain,
    };

    return windows[type](options);
};