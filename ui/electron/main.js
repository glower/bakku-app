const { app, BrowserWindow, Tray, Menu, ipcMain, Notification } = require('electron');
const path = require('path');
const EventSource = require("eventsource");
const windowFactory = require('./helpers/window-factory');
const { APP_NAME, MAIN_WINDOW_WIDTH, FEEDBACK_LINK } = require('./helpers/constants');
const WindowsToaster = require('node-notifier').WindowsToaster;

let tray = null
let window = undefined;
var notifier = new WindowsToaster({
   withFallback: false, // Fallback to Growl or Balloons?
   // customPath: void 0 // Relative/Absolute path if you want to use your fork of SnoreToast.exe
});

app.on('ready', () => {
   createTray()
   createWindow()
   createNotificationListener()
})

const createTray = () => {
   tray = new Tray(path.join('icon.png'))
   setTrayConfigs(tray);
   setTrayListeners(tray);
}

function setTrayConfigs(tray) {
   tray.setHighlightMode('never');
   tray.setIgnoreDoubleClickEvents(true);
}

function setTrayListeners(tray) {
   tray.on('right-click', () => manageTrayRightClick(tray));
   tray.on('click', (event, bounds) => {
      const windowBounds = window.getBounds();
      const trayBounds = tray.getBounds();
      const x = Math.round(trayBounds.x + (trayBounds.width / 2) - (windowBounds.width / 2));;
      const y = Math.round(trayBounds.y - windowBounds.height);
      window.setPosition(x, y);
      window.isVisible() ? hideAllWindows() : showAllWindows();
   });
}

const createWindow = () => {
   window = windowFactory.get('main');
   setWindowConfigs(window);
   // setApplicationMenuToEnableCopyPaste();

   window.loadFile(path.join(__dirname, 'index.html'));
   window.webContents.send('loading', {});
   setWindowListeners(window);

   // https://www.npmjs.com/package/node-notifier#within-electron-packaging
   // setTimeout(function(){
   //    createNotificationListener
   // }, 5000);

   setInterval(() => updater.execute(window), 86400000);
}

function setWindowListeners(window) {
   window.on('closed', () => window = null);
   window.on('blur', () => window.hide());
}

function setWindowConfigs(window) {
   window.setVisibleOnAllWorkspaces(true);
}

function hideAllWindows() {
   BrowserWindow.getAllWindows().forEach(window => window.hide());
}

function showAllWindows() {
   BrowserWindow.getAllWindows().forEach(win => {
      win.show();
      if (win.id !== window.id) win.center();
   });
}

function manageTrayRightClick(tray) {
   window.hide();

   const trayMenuTemplate = [
      {
         label: APP_NAME,
         enabled: false
      },
      {
         type: 'separator'
      },
      {
         label: 'Config',
         type: 'normal',
         // click: () => localStorage.save('activateNotifications', !localStorage.get('activateNotifications'))
      },
      {
         type: 'separator'
      },
      {
         label: 'Quit',
         click: function () {
            window.setClosable(true);
            app.quit();
         }
      }
   ];
   const trayMenu = Menu.buildFromTemplate(trayMenuTemplate);

   tray.popUpContextMenu(trayMenu);
}


ipcMain.on('fixHeight', (event, height) => window.setSize(MAIN_WINDOW_WIDTH, height, true));
if (app.dock) app.dock.hide();

const createNotificationListener = () => {
   console.log("createNotificationListener")
   // http://server/events?stream=messages
   var evtSource = new EventSource("http://localhost:8080/events?stream=messages");
   evtSource.onerror = (err) => {
      console.log("ERROR", err)
   }
   evtSource.onmessage = (evt) => {
      let data = JSON.parse(evt.data)
      console.log(data)
      notifier.notify(
         {
            title: data.type, // String. Required
            message: data.message, // String. Required if remove is not defined
            // icon: void 0, // String. Absolute path to Icon
            sound: false, // Bool | String (as defined by http://msdn.microsoft.com/en-us/library/windows/apps/hh761492.aspx)
            wait: false, // Bool. Wait for User Action against Notification or times out
            //   id: void 0, // Number. ID to use for closing notification.
            //   appID: void 0, // String. App.ID and app Name. Defaults to no value, causing SnoreToast text to be visible.
            //   remove: void 0, // Number. Refer to previously created notification to close.
            //   install: void 0 // String (path, application, app id).  Creates a shortcut <path> in the start menu which point to the executable <application>, appID used for the notifications.
         },
         function (error, response) {
            console.log(response);
         }
      );
   }
}




//// client:
// var EventSource = require('..')
// var es = new EventSource('http://localhost:8080/sse')
// es.addEventListener('server-time', function (e) {
//   console.log(e.data)
// })