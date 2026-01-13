#!/usr/bin/env node

const os = require('os');
const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const VERSION = require('./package.json').version;
const REPO = 'nathanbarrett/dev-swarm-go';

function getPlatform() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
  };

  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64',
  };

  const p = platformMap[platform];
  const a = archMap[arch];

  if (!p || !a) {
    console.error(`Unsupported platform: ${platform}-${arch}`);
    process.exit(1);
  }

  return { platform: p, arch: a };
}

function getBinaryName() {
  const { platform, arch } = getPlatform();
  return `dev-swarm_${VERSION}_${platform}_${arch}.tar.gz`;
}

function getDownloadUrl() {
  const binaryName = getBinaryName();
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;
}

async function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    https.get(url, (response) => {
      // Handle redirects
      if (response.statusCode === 302 || response.statusCode === 301) {
        download(response.headers.location, dest).then(resolve).catch(reject);
        return;
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Download failed: ${response.statusCode}`));
        return;
      }

      response.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

async function install() {
  const binDir = path.join(__dirname, 'bin');
  const binaryPath = path.join(binDir, 'dev-swarm-binary');
  const tarPath = path.join(__dirname, 'dev-swarm.tar.gz');

  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  // Download binary
  console.log(`Downloading dev-swarm v${VERSION}...`);
  const url = getDownloadUrl();

  try {
    await download(url, tarPath);
  } catch (err) {
    console.error(`Failed to download: ${err.message}`);
    console.error(`URL: ${url}`);
    console.error('');
    console.error('Make sure the release exists at:');
    console.error(`https://github.com/${REPO}/releases/tag/v${VERSION}`);
    process.exit(1);
  }

  // Extract binary
  console.log('Extracting...');
  try {
    execSync(`tar -xzf "${tarPath}" -C "${binDir}"`, { stdio: 'inherit' });

    // Rename extracted binary
    const extractedName = 'dev-swarm';
    const extractedPath = path.join(binDir, extractedName);

    if (fs.existsSync(extractedPath)) {
      fs.renameSync(extractedPath, binaryPath);
    }

    // Make executable
    fs.chmodSync(binaryPath, '755');

    // Clean up
    fs.unlinkSync(tarPath);

    console.log('dev-swarm installed successfully!');
  } catch (err) {
    console.error(`Failed to extract: ${err.message}`);
    process.exit(1);
  }
}

install();
