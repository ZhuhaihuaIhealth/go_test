// ─── URL Parsing ────────────────────────────────────────────────

const UrlParser = {
  parse(url) {
    return (
      this.parseGitHub(url) ||
      this.parseGitLab(url) ||
      this.parseRawMd(url)
    );
  },

  parseGitHub(url) {
    // https://github.com/{owner}/{repo}/blob/{ref}/{path}
    const m = url.match(
      /^https?:\/\/github\.com\/([^/]+)\/([^/]+)\/blob\/(.+\.(md|markdown))(?:[?#].*)?$/i
    );
    if (!m) return null;
    const refAndPath = m[3];
    const owner = m[1];
    const repo = m[2];
    return {
      platform: 'GitHub',
      owner,
      repo,
      filePath: refAndPath,
      fileName: refAndPath.split('/').pop(),
      rawUrl: `https://raw.githubusercontent.com/${owner}/${repo}/${refAndPath}`,
    };
  },

  parseGitLab(url) {
    // https://gitlab.com/{owner}/{repo}/-/blob/{ref}/{path}
    const m = url.match(
      /^https?:\/\/gitlab\.com\/([^/]+)\/([^/]+)\/-\/blob\/(.+\.(md|markdown))(?:[?#].*)?$/i
    );
    if (!m) return null;
    const refAndPath = m[3];
    const owner = m[1];
    const repo = m[2];
    return {
      platform: 'GitLab',
      owner,
      repo,
      filePath: refAndPath,
      fileName: refAndPath.split('/').pop(),
      rawUrl: `https://gitlab.com/${owner}/${repo}/-/raw/${refAndPath}`,
    };
  },

  parseRawMd(url) {
    // Any URL ending in .md or .markdown
    if (!/\.(md|markdown)(\?.*)?$/i.test(url)) return null;
    const u = new URL(url);
    const segments = u.pathname.split('/').filter(Boolean);
    return {
      platform: u.hostname,
      owner: segments[0] || '',
      repo: segments[1] || '',
      filePath: segments.join('/'),
      fileName: segments[segments.length - 1] || 'document.md',
      rawUrl: url,
    };
  },
};

// ─── Image Extraction ───────────────────────────────────────────

const ImageExtractor = {
  extract(mdContent) {
    const images = [];
    const seen = new Set();

    const codeBlockRanges = this._getCodeBlockRanges(mdContent);
    const isInCodeBlock = (idx) =>
      codeBlockRanges.some(([s, e]) => idx >= s && idx <= e);

    // ![alt](url) and ![alt](url "title")
    const mdRe = /!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g;
    let m;
    while ((m = mdRe.exec(mdContent)) !== null) {
      if (isInCodeBlock(m.index)) continue;
      if (!seen.has(m[2])) {
        seen.add(m[2]);
        images.push({ url: m[2], fullMatch: m[0], type: 'md' });
      }
    }

    // <img src="url" ...>
    const htmlRe = /<img[^>]+src=["']([^"']+)["'][^>]*>/gi;
    while ((m = htmlRe.exec(mdContent)) !== null) {
      if (isInCodeBlock(m.index)) continue;
      if (!seen.has(m[1])) {
        seen.add(m[1]);
        images.push({ url: m[1], fullMatch: m[0], type: 'html' });
      }
    }

    // Reference-style: [id]: url
    const refRe = /^\[([^\]]+)\]:\s+(\S+)/gm;
    while ((m = refRe.exec(mdContent)) !== null) {
      if (isInCodeBlock(m.index)) continue;
      const url = m[2];
      if (this._looksLikeImage(url) && !seen.has(url)) {
        seen.add(url);
        images.push({ url, fullMatch: m[0], type: 'ref' });
      }
    }

    return images;
  },

  _getCodeBlockRanges(content) {
    const ranges = [];
    const re = /```[\s\S]*?```|`[^`\n]+`/g;
    let m;
    while ((m = re.exec(content)) !== null) {
      ranges.push([m.index, m.index + m[0].length]);
    }
    return ranges;
  },

  _looksLikeImage(url) {
    return /\.(png|jpe?g|gif|svg|webp|bmp|ico|avif)(\?.*)?$/i.test(url);
  },

  resolveUrl(imageUrl, rawBaseUrl) {
    if (/^https?:\/\//i.test(imageUrl)) return imageUrl;
    if (imageUrl.startsWith('//')) return 'https:' + imageUrl;
    const baseDir = rawBaseUrl.substring(0, rawBaseUrl.lastIndexOf('/') + 1);
    try {
      return new URL(imageUrl, baseDir).href;
    } catch {
      return baseDir + imageUrl;
    }
  },
};

// ─── Downloader ─────────────────────────────────────────────────

class MarkdownDownloader {
  constructor({ rawUrl, fileName, mode, onProgress, onLog }) {
    this.rawUrl = rawUrl;
    this.fileName = fileName || 'document.md';
    this.mode = mode; // 'zip' | 'embedded'
    this.onProgress = onProgress;
    this.onLog = onLog;
  }

  async run() {
    this.onLog('正在获取 Markdown 文件...', 'info');
    const mdContent = await this._fetch(this.rawUrl);
    if (!mdContent) throw new Error('无法获取 Markdown 文件内容');

    this.onLog(`已获取 ${this.fileName}（${this._size(mdContent.length)}）`, 'success');
    this.onProgress(10);

    const images = ImageExtractor.extract(mdContent);
    this.onLog(`检测到 ${images.length} 张图片`, 'info');

    if (images.length === 0) {
      this.onLog('没有需要下载的图片，直接保存 MD 文件', 'info');
      this.onProgress(90);
      this._downloadTextFile(mdContent, this.fileName);
      this.onProgress(100);
      this.onLog('下载完成！', 'success');
      return;
    }

    const downloaded = await this._downloadImages(images);

    if (this.mode === 'embedded') {
      await this._saveEmbedded(mdContent, downloaded);
    } else {
      await this._saveZip(mdContent, downloaded);
    }

    this.onProgress(100);
    const successCount = downloaded.filter((d) => d.blob).length;
    this.onLog(
      `完成！成功下载 ${successCount}/${images.length} 张图片`,
      successCount === images.length ? 'success' : 'warn'
    );
  }

  async _downloadImages(images) {
    const results = [];
    const concurrency = 5;
    let completed = 0;
    const total = images.length;

    const prepared = [];
    const usedNames = new Set();
    for (const img of images) {
      const absUrl = ImageExtractor.resolveUrl(img.url, this.rawUrl);
      const localName = this._uniqueFileName(img.url, usedNames);
      usedNames.add(localName);
      prepared.push({ ...img, absUrl, localName });
    }

    const downloadOne = async (item) => {
      try {
        this.onLog(`下载: ${this._truncate(item.url, 50)}`, 'info');
        const resp = await fetch(item.absUrl);
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const blob = await resp.blob();
        completed++;
        this.onProgress(10 + Math.round((completed / total) * 70));
        this.onLog(`✓ ${item.localName}（${this._size(blob.size)}）`, 'success');
        return { ...item, blob, mimeType: blob.type };
      } catch (err) {
        completed++;
        this.onProgress(10 + Math.round((completed / total) * 70));
        this.onLog(`✗ ${this._truncate(item.url, 50)}: ${err.message}`, 'error');
        return { ...item, blob: null, error: err.message };
      }
    };

    for (let i = 0; i < prepared.length; i += concurrency) {
      const batch = prepared.slice(i, i + concurrency);
      const batchResults = await Promise.all(batch.map(downloadOne));
      results.push(...batchResults);
    }

    return results;
  }

  async _saveZip(mdContent, downloaded) {
    this.onLog('正在打包 ZIP...', 'info');
    this.onProgress(85);

    const zip = new JSZip();
    let updatedMd = mdContent;

    for (const img of downloaded) {
      if (img.blob) {
        zip.file(`images/${img.localName}`, img.blob);
        updatedMd = this._replaceImageUrl(
          updatedMd,
          img.url,
          `images/${img.localName}`
        );
      }
    }

    zip.file(this.fileName, updatedMd);

    const zipBlob = await zip.generateAsync({
      type: 'blob',
      compression: 'DEFLATE',
      compressionOptions: { level: 6 },
    });

    this.onProgress(95);
    const baseName = this.fileName.replace(/\.(md|markdown)$/i, '');
    this._downloadBlob(zipBlob, `${baseName}-complete.zip`);
  }

  async _saveEmbedded(mdContent, downloaded) {
    this.onLog('正在内嵌图片为 Base64...', 'info');
    this.onProgress(85);

    let updatedMd = mdContent;

    for (const img of downloaded) {
      if (img.blob) {
        const dataUri = await this._blobToDataUri(img.blob);
        updatedMd = this._replaceImageUrl(updatedMd, img.url, dataUri);
      }
    }

    this.onProgress(95);
    this._downloadTextFile(updatedMd, this.fileName);
  }

  _replaceImageUrl(mdContent, oldUrl, newUrl) {
    const escaped = oldUrl.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    return mdContent.replace(new RegExp(escaped, 'g'), newUrl);
  }

  _uniqueFileName(url, usedNames) {
    let name;
    try {
      const u = new URL(url, 'https://placeholder.com');
      name = decodeURIComponent(u.pathname.split('/').pop());
    } catch {
      name = url.split('/').pop().split('?')[0];
    }

    if (!name || name === '/') name = 'image';
    if (!/\.\w{2,5}$/.test(name)) name += '.png';

    name = name.replace(/[^a-zA-Z0-9._\-\u4e00-\u9fff]/g, '_');

    if (!usedNames.has(name)) return name;

    const dot = name.lastIndexOf('.');
    const base = name.substring(0, dot);
    const ext = name.substring(dot);
    let i = 2;
    while (usedNames.has(`${base}_${i}${ext}`)) i++;
    return `${base}_${i}${ext}`;
  }

  _blobToDataUri(blob) {
    return new Promise((resolve) => {
      const reader = new FileReader();
      reader.onloadend = () => resolve(reader.result);
      reader.readAsDataURL(blob);
    });
  }

  _downloadBlob(blob, filename) {
    const url = URL.createObjectURL(blob);
    chrome.downloads.download({ url, filename, saveAs: true }, () => {
      setTimeout(() => URL.revokeObjectURL(url), 10000);
    });
  }

  _downloadTextFile(text, filename) {
    const blob = new Blob([text], { type: 'text/markdown;charset=utf-8' });
    this._downloadBlob(blob, filename);
  }

  async _fetch(url) {
    const resp = await fetch(url);
    if (!resp.ok) throw new Error(`HTTP ${resp.status} - ${resp.statusText}`);
    return resp.text();
  }

  _size(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1024 / 1024).toFixed(1) + ' MB';
  }

  _truncate(str, max) {
    return str.length > max ? str.substring(0, max) + '...' : str;
  }
}

// ─── Popup Controller ───────────────────────────────────────────

class PopupController {
  constructor() {
    this.fileInfo = null;
    this.isDownloading = false;
    this.els = {
      statusCard: document.getElementById('statusCard'),
      statusIcon: document.getElementById('statusIcon'),
      fileInfoEl: document.getElementById('fileInfo'),
      optionsSection: document.getElementById('optionsSection'),
      downloadBtn: document.getElementById('downloadBtn'),
      progressSection: document.getElementById('progressSection'),
      progressBar: document.getElementById('progressBar'),
      logContainer: document.getElementById('logContainer'),
      manualToggle: document.getElementById('manualToggle'),
      manualInputGroup: document.getElementById('manualInputGroup'),
      manualUrl: document.getElementById('manualUrl'),
      manualConfirm: document.getElementById('manualConfirm'),
    };
  }

  async init() {
    this._bindEvents();
    await this._detectCurrentPage();
  }

  _bindEvents() {
    this.els.downloadBtn.addEventListener('click', () => this._onDownload());

    this.els.manualToggle.addEventListener('click', () => {
      this.els.manualInputGroup.classList.toggle('visible');
    });

    this.els.manualConfirm.addEventListener('click', () => {
      const url = this.els.manualUrl.value.trim();
      if (url) this._setFileInfo(url);
    });

    this.els.manualUrl.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        const url = this.els.manualUrl.value.trim();
        if (url) this._setFileInfo(url);
      }
    });
  }

  async _detectCurrentPage() {
    try {
      const [tab] = await chrome.tabs.query({
        active: true,
        currentWindow: true,
      });
      if (tab?.url) {
        this._setFileInfo(tab.url);
      } else {
        this._showStatus('warn', '无法获取当前页面 URL');
      }
    } catch (err) {
      this._showStatus('err', '检测失败: ' + err.message);
    }
  }

  _setFileInfo(url) {
    const info = UrlParser.parse(url);
    if (!info) {
      this._showStatus('warn', '当前页面不是 Markdown 文件');
      this.els.optionsSection.style.display = 'none';
      this.fileInfo = null;
      return;
    }

    this.fileInfo = info;
    this._showStatus('ok', `检测到 ${info.platform} Markdown 文件`);
    this.els.statusCard.classList.add('detected');

    this.els.fileInfoEl.style.display = 'grid';
    this.els.fileInfoEl.innerHTML = `
      <span class="label">仓库</span>
      <span class="value">${info.owner}/${info.repo}</span>
      <span class="label">文件</span>
      <span class="value">${info.fileName}</span>
      <span class="label">路径</span>
      <span class="value">${info.filePath}</span>
    `;

    this.els.optionsSection.style.display = 'block';
  }

  _showStatus(type, message) {
    const icons = { ok: '✅', warn: '⚠️', err: '❌', info: '⏳' };
    this.els.statusIcon.className = 'status-icon ' + type;
    this.els.statusIcon.innerHTML = `
      <span>${icons[type] || '⏳'}</span>
      <span>${message}</span>
    `;
  }

  async _onDownload() {
    if (this.isDownloading || !this.fileInfo) return;
    this.isDownloading = true;

    const mode = document.querySelector('input[name="mode"]:checked').value;

    this.els.downloadBtn.disabled = true;
    this.els.downloadBtn.classList.add('downloading');
    this.els.downloadBtn.innerHTML = `
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin">
        <path d="M12 2v4m0 12v4m-7.07-14.07l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/>
      </svg>
      正在下载...
    `;

    this.els.progressSection.classList.add('visible');
    this.els.logContainer.innerHTML = '';
    this.els.progressBar.style.width = '0%';
    this.els.progressBar.className = 'progress-bar-fill';

    const downloader = new MarkdownDownloader({
      rawUrl: this.fileInfo.rawUrl,
      fileName: this.fileInfo.fileName,
      mode,
      onProgress: (pct) => {
        this.els.progressBar.style.width = pct + '%';
      },
      onLog: (msg, type) => this._log(msg, type),
    });

    try {
      await downloader.run();
      this.els.progressBar.classList.add('done');
      this.els.downloadBtn.innerHTML = '✅ 下载完成';
    } catch (err) {
      this.els.progressBar.classList.add('error');
      this._log('下载失败: ' + err.message, 'error');
      this.els.downloadBtn.innerHTML = '❌ 下载失败';
    } finally {
      setTimeout(() => {
        this.isDownloading = false;
        this.els.downloadBtn.disabled = false;
        this.els.downloadBtn.classList.remove('downloading');
        this.els.downloadBtn.innerHTML = `
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7,10 12,15 17,10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          抓取并下载
        `;
      }, 3000);
    }
  }

  _log(message, type = 'info') {
    const icons = { info: '›', success: '✓', warn: '⚠', error: '✗' };
    const entry = document.createElement('div');
    entry.className = `log-entry ${type}`;
    entry.innerHTML = `
      <span class="log-icon">${icons[type] || '›'}</span>
      <span>${message}</span>
    `;
    this.els.logContainer.appendChild(entry);
    this.els.logContainer.scrollTop = this.els.logContainer.scrollHeight;
  }
}

// ─── Init ───────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', () => {
  new PopupController().init();
});
