# 史莱姆区块可视化工具

[![许可证: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go 版本](https://img.shields.io/badge/Go-1.16+-blue.svg)](https://golang.org/)
[![平台](https://img.shields.io/badge/platform-Android%20%7C%20Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](https://github.com)

[中文](README.zh-CN.md) | [English](README.md)

一个高性能的 Minecraft 史莱姆区块可视化工具。生成显示 Java 版和基岩版史莱姆区块分布的 PNG 地图。

## 功能特点

- **双版本支持** - 同时支持 Java 版和基岩版
- **高性能** - 可配置工作进程数的并行处理
- **灵活的区域选择** - 自定义区域列表或自动环模式
- **可定制输出** - 可配置的颜色、命名模板和输出格式
- **跨平台** - 支持 Android、Windows、Linux 和 macOS

## 快速开始

### 环境要求

- Go 1.16 或更高版本

### 构建

```bash
chmod +x build.sh
./build.sh
```

这将生成可执行二进制文件 `app`（Windows 上为 `app.exe`）。

### 配置

在可执行文件同目录下创建 `config.toml` 文件：

```toml
# 版本类型："java" 或 "bedrock"
edition = "java"

# 种子（仅 Java 版需要）
seed = "123456789"

# 史莱姆区块颜色（十六进制格式）
slime_color = "#00ff00"

# 输出目录
output_dir = "./output"

# 图片格式（目前仅支持 png）
format = "png"

# 进程数（0 = 自动 = CPU 核心数）
workers = 0

# 输出文件名模板
# 可用变量：{x1}， {z1}， {x2}， {z2}
output_name = "{x1}_{z1}_{x2}_{z2}"

# 环模式（启用后将忽略 regions 数组）
enable_ring = true

# 环数（每个象限 n×n 个区块）
ring_count = 3

# 原点坐标
ring_origin = [0， 0]

# 环大小（每个区块的边长）
ring_size = 512

# 自定义区域（仅在 enable_ring = false 时使用）
# 格式：[x1, z1, x2, z2]
regions = [
  [0， 0， 511， 511]，
  [512， 0， 1023， 511]
]
```

### 使用方法

运行编译后的二进制文件：

```bash
./app
```

工具会从同目录读取 `config.toml`，并在指定输出目录生成 PNG 图片。

## 算法参考

基岩版史莱姆区块算法基于以下人员的逆向工程工作：

- **@protolambda** - [Slime Finder PE](https://github.com/depressed-pho/slime-finder-pe)
- **@jocopa3** - [基岩版史莱姆区块算法](https://gist.github.com/protolambda/00b85bf34a75fd8176342b1ad28bfccc)

特别感谢他们对 Minecraft 社区的贡献。

## 性能表现

- **并行处理** - 使用 Go 的 goroutine 进行并发区块评估
- **可配置工作进程** - 根据 CPU 调整工作进程数
- **内存高效** - 处理区域时无需将整个地图加载到内存

8 核 CPU 上的性能基准测试：

| 区域大小 | 区块数量 | 耗时 | 内存占用 |
|---------|---------|------|----------|
| 512×512 | 262，144 | ~2.5秒 | ~50MB |
| 1024×1024 | 1，048，576 | ~9.8秒 | ~180MB |
| 2048×2048 | 4，194，304 | ~41秒 | ~700MB |

## 文件命名

输出文件名支持以下变量：

- `{x1}` - 起始 X 坐标
- `{z1}` - 起始 Z 坐标
- `{x2}` - 结束 X 坐标
- `{z2}` - 结束 Z 坐标

示例：
- `{x1}_{z1}_{x2}_{z2}.png` → `0_0_511_511.png`
- `region_{x1}_{z1}.png` → `region_0_0.png`

## 环模式

环模式自动围绕中心点生成环状区域：

- **1 环**：每个象限 1×1 个区块（共 4 个区域）
- **2 环**：每个象限 2×2 个区块（共 16 个区域）
- **3 环**：每个象限 3×3 个区块（共 36 个区域）

这对于生成以特定坐标（如出生点）为中心的地图非常有用。

## 项目结构

```
slime-chunk-visualizer/
├── build.sh          # 构建脚本
├── main.go           # 主程序入口
├── config.toml       # 配置文件
├── chunks/           # 默认输出目录
├── README.md         # 英文文档
├── README.zh-CN.md   # 中文文档
└── LICENSE           # MIT 许可证
```

## 故障排除

### "Failed to load config.toml"

工具会在可执行文件同目录查找 `config.toml`。请确保文件存在且 TOML 语法正确。

### 内存不足错误

减小区域大小或降低 `ring_count`。对于非常大的区域，考虑将其拆分为更小的块。

### Java 版性能缓慢

Java 版使用更复杂的随机数生成器。如果不需要 Java 特定功能，可考虑使用基岩版模式。

## 致谢

- Mojang 开发了 Minecraft
- @protolambda 和 @jocopa3 逆向工程了基岩版史莱姆区块算法
- Go 社区提供了出色的并发支持

## 贡献

欢迎贡献！请随时提交拉取请求。

1. Fork 本仓库
2. 创建您的功能分支（`git checkout -b feature/AmazingFeature`）
3. 提交您的更改（`git commit -m '添加一些 AmazingFeature'`）
4. 推送到分支（`git push origin feature/AmazingFeature`）
5. 打开一个拉取请求

## 免责声明

本项目不隶属于 Mojang 或 Microsoft。Minecraft 是 Mojang AB 的商标。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。