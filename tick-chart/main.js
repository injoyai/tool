const { app, BrowserWindow, ipcMain, Menu, Tray, nativeImage } = require('electron');
const path = require('path');
const fs = require('fs');

app.commandLine.appendSwitch('no-sandbox');
app.commandLine.appendSwitch('disable-gpu');

let mainWindow = null;
let tray = null;

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

function createIcon() {
  const iconPath = path.join(__dirname, 'icon.png');
  if (fs.existsSync(iconPath)) {
    return nativeImage.createFromPath(iconPath);
  }

  const size = 32;
  const canvas = Buffer.alloc(size * size * 4);
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const idx = (y * size + x) * 4;
      const cx = size / 2;
      const cy = size / 2;
      const dist = Math.sqrt((x - cx) ** 2 + (y - cy) ** 2);
      if (dist < size / 2 - 2) {
        canvas[idx] = 107;
        canvas[idx + 1] = 114;
        canvas[idx + 2] = 128;
        canvas[idx + 3] = 255;
      } else {
        canvas[idx] = 0;
        canvas[idx + 1] = 0;
        canvas[idx + 2] = 0;
        canvas[idx + 3] = 0;
      }
    }
  }
  return nativeImage.createFromBuffer(canvas, { width: size, height: size });
}

function createWindow() {
  const config = loadConfig();

  const opacity = (config.opacity >= 0.1 && config.opacity <= 1.0) ? config.opacity : 1.0;

  mainWindow = new BrowserWindow({
    width: 600,
    height: 200,
    title: '',
    resizable: true,
    show: false,
    frame: true,
    skipTaskbar: true,
    icon: null,
    opacity: opacity,
    x: 100,
    y: 100,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js'),
      backgroundThrottling: false
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

  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

function createTray() {
  const icon = createIcon();
  tray = new Tray(icon);

  const contextMenu = Menu.buildFromTemplate([
    { label: '显示', click: () => { if (mainWindow) mainWindow.show(); } },
    { label: '退出', click: () => { app.quit(); } }
  ]);

  tray.setToolTip('分时图');
  tray.setContextMenu(contextMenu);

  tray.on('click', () => {
    if (mainWindow) {
      if (mainWindow.isVisible()) {
        mainWindow.hide();
      } else {
        mainWindow.show();
      }
    }
  });
}

app.whenReady().then(() => {
  Menu.setApplicationMenu(null);
  createTray();
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

app.on('before-quit', () => {
  if (tray) {
    tray.destroy();
    tray = null;
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
