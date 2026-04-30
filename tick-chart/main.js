const { app, BrowserWindow, ipcMain, Menu } = require('electron');
const path = require('path');
const fs = require('fs');

app.commandLine.appendSwitch('no-sandbox');
app.commandLine.appendSwitch('disable-gpu');

let mainWindow = null;

const configPath = path.join(app.getPath('userData'), 'config.json');

function loadConfig() {
  try {
    if (fs.existsSync(configPath)) {
      const data = fs.readFileSync(configPath, 'utf-8');
      return JSON.parse(data);
    }
  } catch (e) {
    console.error('Failed to load config:', e);
  }
  return { opacity: 1.0, alwaysOnTop: false, code: '', colorInverted: false };
}

function saveConfig(config) {
  try {
    fs.writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8');
  } catch (e) {
    console.error('Failed to save config:', e);
  }
}

function createWindow() {
  const config = loadConfig();

  const opacity = (config.opacity >= 0.1 && config.opacity <= 1.0) ? config.opacity : 1.0;

  mainWindow = new BrowserWindow({
    width: 600,
    height: 200,
    title: '',
    resizable: true,
    show: true,
    frame: true,
    icon: null,
    opacity: opacity,
    x: 100,
    y: 100,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  if (config.alwaysOnTop) {
    mainWindow.setAlwaysOnTop(true);
  }

  mainWindow.loadFile('index.html', {
    query: {
      opacity: opacity.toString(),
      alwaysOnTop: config.alwaysOnTop.toString(),
      code: config.code || '',
      colorInverted: config.colorInverted.toString()
    }
  });

  mainWindow.webContents.on('did-finish-load', () => {
    console.log('Page loaded successfully');
  });

  mainWindow.webContents.on('did-fail-load', (event, errorCode, errorDescription) => {
    console.error('Failed to load:', errorCode, errorDescription);
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

app.whenReady().then(() => {
  Menu.setApplicationMenu(null);
  createWindow();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

ipcMain.on('set-opacity', (event, value) => {
  const opacity = parseFloat(value);
  if (mainWindow) {
    mainWindow.setOpacity(opacity);
  }
  const config = loadConfig();
  config.opacity = opacity;
  saveConfig(config);
});

ipcMain.on('set-always-on-top', (event, checked) => {
  const alwaysOnTop = !!checked;
  if (mainWindow) {
    mainWindow.setAlwaysOnTop(alwaysOnTop);
  }
  const config = loadConfig();
  config.alwaysOnTop = alwaysOnTop;
  saveConfig(config);
});

ipcMain.on('set-code', (event, code) => {
  const config = loadConfig();
  config.code = code;
  saveConfig(config);
});

ipcMain.on('set-color-inverted', (event, inverted) => {
  const colorInverted = !!inverted;
  const config = loadConfig();
  config.colorInverted = colorInverted;
  saveConfig(config);
});

ipcMain.on('start-drag', (event) => {
  if (mainWindow) {
    mainWindow.moveTop();
  }
});
