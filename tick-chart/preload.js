const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  setOpacity: (value) => ipcRenderer.send('set-opacity', value),
  setAlwaysOnTop: (checked) => ipcRenderer.send('set-always-on-top', checked),
  setCode: (code) => ipcRenderer.send('set-code', code),
  setColorInverted: (inverted) => ipcRenderer.send('set-color-inverted', inverted),
  startDrag: () => ipcRenderer.send('start-drag')
});
