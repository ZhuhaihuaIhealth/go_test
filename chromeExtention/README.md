# MD Grabber - Chrome Extension

抓取远程 Markdown 文件及其所有图片，一键打包下载。解决从 GitHub 等平台下载 MD 文件时图片缺失的问题。

## 功能特性

- **自动检测** GitHub / GitLab 上的 Markdown 文件
- **图片完整下载** 自动解析 MD 中的所有图片引用（`![]()`、`<img>`、引用式），下载并打包
- **两种下载模式**
  - **ZIP 打包**: MD 文件 + images 文件夹，图片路径自动替换为本地路径
  - **单文件模式**: 图片转为 Base64 内嵌到 MD 文件中
- **并发下载** 图片 5 个一组并发下载，速度快
- **智能解析** 自动跳过代码块中的图片引用，避免误抓
- **手动输入** 支持手动粘贴任意 MD 文件的 raw URL

## 安装方法

1. 打开 Chrome 浏览器，进入 `chrome://extensions/`
2. 开启右上角的 **开发者模式**
3. 点击 **加载已解压的扩展程序**
4. 选择本项目的 `chromeExtention` 文件夹
5. 扩展安装完成，工具栏会出现 MD Grabber 图标

## 使用方法

1. 在浏览器中打开一个 GitHub 仓库中的 `.md` 文件（如 README.md）
2. 点击工具栏中的 **MD Grabber** 图标
3. 插件会自动检测当前页面的 Markdown 文件信息
4. 选择下载模式（ZIP 打包 / 单文件内嵌）
5. 点击 **抓取并下载** 按钮
6. 等待下载完成，文件会保存到你的下载目录

## 支持的平台

| 平台 | URL 格式 | 状态 |
|------|---------|------|
| GitHub | `github.com/{user}/{repo}/blob/{branch}/{path}.md` | ✅ 完整支持 |
| GitLab | `gitlab.com/{user}/{repo}/-/blob/{branch}/{path}.md` | ✅ 完整支持 |
| 任意 raw URL | 以 `.md` 或 `.markdown` 结尾的 URL | ✅ 支持 |

## 支持的图片引用格式

- `![alt text](image-url)` — 标准 Markdown 图片
- `![alt text](image-url "title")` — 带标题的图片
- `<img src="image-url" />` — HTML 图片标签
- `[ref-id]: image-url` — 引用式图片定义
- 相对路径（`./images/pic.png`、`../assets/img.jpg`）自动解析为绝对路径

## 技术栈

- Chrome Extension Manifest V3
- JSZip 3.10.1（ZIP 打包）
- 纯原生 JS，无需构建工具
