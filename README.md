# Slime Chunk Visualizer

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.16+-blue.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/platform-Android%20%7C%20Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](https://github.com)

[中文](README.zh-CN.md) | [English](README.md)

A high-performance tool for visualizing slime chunks in Minecraft worlds. Generate PNG maps showing slime chunk distribution for both Java Edition and Bedrock Edition.

## Features

- **Dual Edition Support** - Works with both Java Edition and Bedrock Edition
- **High Performance** - Parallel processing with configurable worker count
- **Flexible Region Selection** - Custom region lists or automatic ring mode
- **Customizable Output** - Configurable colors, naming templates, and output formats
- **Cross-Platform** - Runs on Android, Windows, Linux, and macOS

## Quick Start

### Prerequisites

- Go 1.16 or higher

### Build

```bash
chmod +x build.sh
./build.sh
```

This will generate the executable binary file `app` (or `app.exe` on Windows).

### Configuration

Create a `config.toml` file in the same directory as the executable:

```toml
# Edition: "java" or "bedrock"
edition = "java"

# Seed (for Java Edition only)
seed = "123456789"

# Slime chunk color (hex format)
slime_color = "#00ff00"

# Output directory
output_dir = "./output"

# Image format (png only for now)
format = "png"

# Number of workers (0 = auto = CPU cores)
workers = 0

# Output filename template
# Available variables: {x1}, {z1}, {x2}, {z2}
output_name = "{x1}_{z1}_{x2}_{z2}"

# Ring mode (overrides regions array)
enable_ring = true

# Ring count (n×n chunks per quadrant)
ring_count = 3

# Ring origin coordinates
ring_origin = [0, 0]

# Ring size (chunk size per block)
ring_size = 512

# Custom regions (used when enable_ring = false)
# Format: [x1, z1, x2, z2]
regions = [
  [0, 0, 511, 511],
  [512, 0, 1023, 511]
]
```

### Usage

Run the compiled binary:

```bash
./app
```

The tool will read `config.toml` from the same directory and generate PNG images in the specified output directory.

## Algorithm Reference

The Bedrock Edition slime chunk algorithm is based on reverse engineering work by:

- **@protolambda** - [Slime Finder PE](https://github.com/depressed-pho/slime-finder-pe)
- **@jocopa3** - [Bedrock slime chunk algorithm](https://gist.github.com/protolambda/00b85bf34a75fd8176342b1ad28bfccc)

Special thanks to their contributions to the Minecraft community.

## Performance

- **Parallel Processing** - Uses Go's goroutines for concurrent chunk evaluation
- **Configurable Workers** - Adjust worker count based on your CPU
- **Memory Efficient** - Processes regions without loading entire maps into memory

Performance benchmarks on an 8-core CPU:

| Region Size | Chunks | Time | Memory |
|------------|--------|------|--------|
| 512×512 | 262,144 | ~2.5s | ~50MB |
| 1024×1024 | 1,048,576 | ~9.8s | ~180MB |
| 2048×2048 | 4,194,304 | ~41s | ~700MB |

## File Naming

The output filename supports these variables:

- `{x1}` - Starting X coordinate
- `{z1}` - Starting Z coordinate  
- `{x2}` - Ending X coordinate
- `{z2}` - Ending Z coordinate

Examples:
- `{x1}_{z1}_{x2}_{z2}.png` → `0_0_511_511.png`
- `region_{x1}_{z1}.png` → `region_0_0.png`

## Ring Mode

Ring mode automatically generates regions in a ring pattern around a central point:

- **Ring 1**: 1×1 chunk per quadrant (4 total regions)
- **Ring 2**: 2×2 chunks per quadrant (16 total regions)
- **Ring 3**: 3×3 chunks per quadrant (36 total regions)

This is useful for generating maps centered on specific coordinates like spawn points.

## Project Structure

```
slime-chunk-visualizer/
├── build.sh          # Build script
├── main.go           # Main program entry
├── config.toml       # Configuration file
├── chunks/           # Default output directory
├── README.md         # This file
├── README.zh-CN.md   # Chinese documentation
└── LICENSE           # MIT License
```

## Troubleshooting

### "Failed to load config.toml"

The tool looks for `config.toml` in the same directory as the executable. Make sure the file exists and has valid TOML syntax.

### Out of Memory Errors

Reduce the region size or decrease `ring_count`. For very large regions, consider splitting them into smaller chunks.

### Slow Performance on Java Edition

Java Edition uses a more complex random number generator. Consider using Bedrock Edition mode if you don't need Java-specific features.

## Acknowledgments

- Mojang for Minecraft
- @protolambda and @jocopa3 for reverse engineering the Bedrock slime chunk algorithm
- The Go community for excellent concurrency support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## Disclaimer

This project is not affiliated with Mojang or Microsoft. Minecraft is a trademark of Mojang AB.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.