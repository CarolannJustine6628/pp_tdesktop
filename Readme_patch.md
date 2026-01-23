# Telegram Desktop 补丁说明文档

本文档说明本仓库中包含的两个主要补丁文件的功能、应用方法和注意事项。

## 补丁文件列表

1. **0001-Implement-message-encryption-and-local-message-index.patch**
   - 功能：实现消息加密和本地消息索引
   - 文件大小：约 1411 行

2. **remove-ads.patch**
   - 功能：移除广告和禁用部分商业化功能
   - 文件大小：约 735 行

---

## 补丁 1：消息加密和本地消息索引

### 功能概述

该补丁为 Telegram Desktop 添加了以下功能：

#### 1. 消息加密功能
- **AES-256-CBC 加密**：使用 OpenSSL 实现的消息加密
- **密钥管理**：从 `sec.txt` 文件读取密钥（SHA256 哈希生成）
- **自动加密/解密**：
  - 发送消息时自动加密
  - 接收消息时自动解密
  - 编辑消息时自动处理加密文本

#### 2. 本地消息索引
- **本地存储**：将消息文本存储到本地加密数据库
- **搜索功能**：提供 Go 工具进行本地消息搜索
- **Web 界面**：通过 HTTP 服务器（默认端口 8800）提供搜索界面

#### 3. 图片质量优化
- 禁用图片自动压缩
- 使用原始图片质量（JPEG quality 100）

### 修改的文件

```
新增文件：
- Telegram/SourceFiles/core/message_encryption.cpp
- Telegram/SourceFiles/core/message_encryption.h
- Telegram/SourceFiles/storage/local_message_index.cpp
- Telegram/SourceFiles/storage/local_message_index.h
- go_search_tool/go.mod
- go_search_tool/main.go
- go_search_tool/parser.go

修改文件：
- Telegram/CMakeLists.txt
- Telegram/SourceFiles/apiwrap.cpp
- Telegram/SourceFiles/core/application.cpp
- Telegram/SourceFiles/history/history.cpp
- Telegram/SourceFiles/history/history.h
- Telegram/SourceFiles/history/history_item.cpp
- Telegram/SourceFiles/history/history_item_edition.cpp
- Telegram/SourceFiles/history/history_item_helpers.cpp
- Telegram/SourceFiles/history/history_item_helpers.h
- Telegram/SourceFiles/storage/localimageloader.cpp
- .github/workflows/win.yml
```

### 使用方法

#### 启用消息加密

1. 创建 `sec.txt` 文件（放在应用程序目录或当前工作目录）
2. 在文件中写入任意内容作为密钥种子
3. 启动 Telegram Desktop，加密功能会自动启用

#### 使用本地消息搜索

1. 编译 Go 搜索工具：
   ```bash
   cd go_search_tool
   go build -o go_search_tool
   ```

2. 运行搜索服务器：
   ```bash
   ./go_search_tool -path tdata/local_message_index -port 8800
   ```

3. 在浏览器中访问 `http://localhost:8800` 进行搜索

### 技术细节

- **加密算法**：AES-256-CBC
- **密钥生成**：SHA256(sec.txt 内容)
- **IV 生成**：基于密钥哈希派生
- **数据格式**：Base64 编码的加密文本
- **索引存储**：AES-256-GCM 加密的本地文件

---

## 补丁 2：移除广告和禁用商业化功能

### 功能概述

该补丁移除了 Telegram Desktop 中的广告和部分商业化功能，包括：

#### 1. 禁用赞助消息（广告）
- 移除频道和机器人对话中的赞助消息
- 禁用视频中的赞助广告

#### 2. 禁用反应（Reactions）功能
- 禁用所有消息的反应功能
- 移除"查看谁反应了"的上下文菜单项
- 禁用反应数据的设置和更新

#### 3. 禁用网页预览
- 禁用自己发送消息的网页预览
- 禁用其他用户消息的网页预览

#### 4. 禁用其他功能
- **固定消息**：禁用消息固定功能
- **录音按钮**：禁用语音消息录制
- **通话功能**：禁用语音通话和群组通话
- **静音/取消静音**：禁用静音按钮
- **Stories**：禁用故事功能
- **复制限制**：允许复制受限帖子/消息
- **顶部栏建议**：移除 Premium 广告和生日提醒

#### 5. UI 调整
- 禁用回复消息的背景颜色
- 禁用回复消息的背景表情符号
- 移除代码块的点击复制功能

### 修改的文件

```
- .gitignore
- Telegram/SourceFiles/data/components/sponsored_messages.cpp
- Telegram/SourceFiles/history/history_item.cpp
- Telegram/SourceFiles/history/history_item_edition.cpp
- Telegram/SourceFiles/history/view/controls/history_view_webpage_processor.cpp
- Telegram/SourceFiles/apiwrap.cpp
- Telegram/SourceFiles/history/history_inner_widget.cpp
- Telegram/SourceFiles/history/view/history_view_context_menu.cpp
- Telegram/SourceFiles/data/data_peer.cpp
- Telegram/SourceFiles/history/history_widget.cpp
- Telegram/SourceFiles/history/view/controls/history_view_compose_controls.cpp
- Telegram/SourceFiles/history/view/history_view_top_bar_widget.cpp
- Telegram/SourceFiles/dialogs/dialogs_row.cpp
- Telegram/SourceFiles/dialogs/dialogs_widget.cpp
- Telegram/SourceFiles/info/profile/info_profile_widget.cpp
- Telegram/SourceFiles/core/ui_integration.cpp
- Telegram/SourceFiles/ui/chat/chat_style.h
- Telegram/SourceFiles/history/view/history_view_reply.cpp
- Telegram/SourceFiles/history/view/history_view_list_widget.cpp
- Telegram/SourceFiles/media/view/media_view_overlay_widget.cpp
```

---

## 应用补丁

### 前置要求

- Git 仓库（基于 Telegram Desktop 官方仓库）
- 当前代码库版本：v6.4.2（建议）

### 应用步骤

#### 方法 1：使用 git apply（推荐）

```bash
# 1. 确保代码库干净
git status

# 2. 先应用加密补丁
git apply 0001-Implement-message-encryption-and-local-message-index.patch

# 3. 再应用去广告补丁（可能有冲突，需要手动解决）
git apply remove-ads.patch

# 4. 如果有冲突，手动解决后：
git add <冲突文件>
git apply --continue
```

#### 方法 2：使用 patch 命令

```bash
# 应用加密补丁
patch -p1 < 0001-Implement-message-encryption-and-local-message-index.patch

# 应用去广告补丁
patch -p1 < remove-ads.patch
```

### 冲突解决

#### 已知冲突点

**文件：`Telegram/SourceFiles/apiwrap.cpp`**

- **位置**：`sendMessage()` 函数
- **原因**：两个补丁都修改了此函数的不同部分
- **解决方法**：
  1. 保留加密补丁的加密逻辑（行 4005-4019）
  2. 保留 remove-ads 的 `ignoreWebPage = true` 修改（原行 4009）
  3. 确保两处修改都正确应用

**示例合并后的代码：**

```cpp
// 加密补丁的修改（保留）
QString messageText = sending.text;
if (Core::MessageEncryption::Instance().isEnabled()) {
    messageText = Core::MessageEncryption::Instance().encryptMessage(sending.text);
}
MTPstring msgText(MTP_string(messageText));

// ... 其他代码 ...

// remove-ads 补丁的修改（保留）
const auto ignoreWebPage = true;  // 原：message.webPage.removed || (exactWebPage && !isLast);
```

#### 其他文件

- `history_item.cpp`：两个补丁修改不同函数，无冲突
- `history_item_edition.cpp`：只有加密补丁修改，无冲突

---

## 编译说明

### 依赖要求

#### 加密补丁需要：
- OpenSSL 开发库（libssl-dev / openssl-devel）
- Go 1.21+（用于搜索工具）

#### 标准编译依赖：
- CMake 3.16+
- Qt 5.15+ 或 Qt 6.x
- C++17 编译器

### 编译步骤

```bash
# 1. 配置构建
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release

# 2. 编译
cmake --build . --parallel

# 3. （可选）编译 Go 搜索工具
cd ../go_search_tool
go build -o go_search_tool
```

---

## 使用注意事项

### 消息加密功能

1. **密钥安全**：
   - `sec.txt` 文件包含密钥种子，请妥善保管
   - 丢失密钥将无法解密已加密的消息
   - 建议备份 `sec.txt` 文件

2. **兼容性**：
   - 加密消息只能被拥有相同 `sec.txt` 的客户端解密
   - 不同设备需要共享相同的 `sec.txt` 文件

3. **性能影响**：
   - 加密/解密操作会增加少量 CPU 开销
   - 对正常使用影响很小

### 去广告功能

1. **功能限制**：
   - 禁用反应功能后，无法使用任何反应
   - 禁用通话功能后，无法进行语音/视频通话
   - 禁用 Stories 后，无法查看和发布故事

2. **数据影响**：
   - 这些修改只影响客户端显示
   - 服务器端数据不受影响
   - 使用官方客户端仍可看到完整功能

---

## 故障排除

### 加密功能不工作

1. 检查 `sec.txt` 文件是否存在
2. 检查文件路径是否正确（应用程序目录或当前目录）
3. 查看日志确认加密模块是否初始化成功

### 搜索工具无法启动

1. 检查 Go 版本：`go version`（需要 1.21+）
2. 检查数据目录路径是否正确
3. 确认端口 8800 未被占用

### 补丁应用失败

1. 确认代码库版本匹配（建议 v6.4.2）
2. 检查是否有未提交的更改
3. 查看冲突信息，手动解决冲突

---

## 版本信息

- **补丁创建日期**：
  - 加密补丁：2026-01-19
  - 去广告补丁：2022-06-22 至 2025-06-04（多个提交）

- **目标版本**：Telegram Desktop v6.4.2

- **测试状态**：请在实际使用前进行充分测试

---

## 许可证

这些补丁基于 Telegram Desktop 的源代码修改。请遵守 Telegram Desktop 的原始许可证要求。

---

## 贡献

如有问题或建议，请：
1. 检查补丁文件中的原始提交信息
2. 查看 Telegram Desktop 官方文档
3. 提交 Issue 或 Pull Request

---

## 更新日志

### 加密补丁
- 2026-01-19：初始版本，实现消息加密和本地索引

### 去广告补丁
- 2022-06-22：初始版本，禁用赞助消息
- 2023-05-22：禁用网页预览
- 2023-05-23：禁用其他用户消息预览
- 2023-05-24：禁用固定消息和录音
- 2023-05-25：禁用通话和静音功能
- 2023-09-28：禁用 Stories
- 2023-11-19：UI 调整（回复背景等）
- 2023-12-19：允许复制受限内容
- 2025-06-04：移除顶部栏建议

---

**注意**：这些补丁修改了 Telegram Desktop 的核心功能，可能影响应用的稳定性和功能完整性。请谨慎使用，并确保理解所有修改的影响。
