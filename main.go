package main

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Edition      string   `toml:"edition"`
	Seed         string   `toml:"seed"`
	SlimeColor   string   `toml:"slime_color"`
	OutputDir    string   `toml:"output_dir"`
	Format       string   `toml:"format"`
	Regions      [][]int  `toml:"regions"`
	Workers      int      `toml:"workers"`
	OutputName   string   `toml:"output_name"`
	EnableRing   bool     `toml:"enable_ring"`
	RingCount    int      `toml:"ring_count"`
	RingOrigin   []int    `toml:"ring_origin"`
	RingSize     int      `toml:"ring_size"`
}

type MTwister struct {
	state       [624]uint32
	index       int
	initialized bool
}

func NewMTwister() *MTwister {
	return &MTwister{
		index: 624,
	}
}

func (mt *MTwister) InitGenrand(seed uint32) {
	mt.state[0] = seed & 0xffffffff
	for i := 1; i < 624; i++ {
		mt.state[i] = (1812433253 * (mt.state[i-1] ^ (mt.state[i-1] >> 30)) + uint32(i))
		mt.state[i] &= 0xffffffff
	}
	mt.index = 624
	mt.initialized = true
}

func (mt *MTwister) NextState() {
	if !mt.initialized {
		panic("MTwister not initialized")
	}
	for i := 0; i < 624; i++ {
		y := (mt.state[i] & 0x80000000) + (mt.state[(i+1)%624] & 0x7fffffff)
		mt.state[i] = mt.state[(i+397)%624] ^ (y >> 1)
		if y%2 != 0 {
			mt.state[i] ^= 0x9908b0df
		}
	}
	mt.index = 0
}

func (mt *MTwister) GenrandInt32() uint32 {
	if mt.index >= 624 {
		mt.NextState()
	}
	y := mt.state[mt.index]
	mt.index++
	y ^= (y >> 11)
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= (y >> 18)
	return y
}

func IsSlimeChunkBedrock(cX, cZ int) bool {
	chunkXUint := uint32(cX)
	chunkZUint := uint32(cZ)
	seed := (uint64(chunkXUint) * 0x1f1f1f1f) ^ uint64(chunkZUint)
	seed &= 0xffffffff
	mt := NewMTwister()
	mt.InitGenrand(uint32(seed))
	n := uint64(mt.GenrandInt32())
	m := uint64(0xcccccccd)
	product := n * m
	hi := (product >> 32) & 0xffffffff
	hiShift3 := (hi >> 3) & 0xffffffff
	res := (((hiShift3 + (hiShift3 * 4)) & 0xffffffff) * 2) & 0xffffffff
	return n == res
}

func stringToSeed(s string) int64 {
	if s == "" {
		return 0
	}
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

func IsSlimeChunkJava(seedStr string, chunkX, chunkZ int) bool {
	var worldSeed int64
	if seedStr == "" {
		worldSeed = 0
	} else {
		if num, err := strconv.ParseInt(seedStr, 10, 64); err == nil {
			worldSeed = num
		} else {
			worldSeed = stringToSeed(seedStr)
		}
	}
	seed := worldSeed +
		int64(chunkX)*int64(chunkX)*4987142 +
		int64(chunkX)*5947611 +
		int64(chunkZ)*int64(chunkZ)*4392871 +
		int64(chunkZ)*389711
	seed ^= 987234911
	rng := NewJavaRandom(seed)
	return rng.NextInt(10) == 0
}

type JavaRandom struct {
	seed int64
}

func NewJavaRandom(seed int64) *JavaRandom {
	return &JavaRandom{seed: (seed ^ 0x5DEECE66D) & ((1 << 48) - 1)}
}

func (r *JavaRandom) Next(bits int) int {
	r.seed = (r.seed*25214903917 + 11) & ((1 << 48) - 1)
	return int(r.seed >> (48 - bits))
}

func (r *JavaRandom) NextInt(bound int) int {
	if bound <= 0 {
		return 0
	}
	if (bound & -bound) == bound {
		return int((int64(bound) * int64(r.Next(31))) >> 31)
	}
	for {
		bits := r.Next(31)
		val := bits % bound
		if bits-val+(bound-1) >= 0 {
			return val
		}
	}
}

func ParseColor(colorStr string) color.RGBA {
	if colorStr == "" {
		return color.RGBA{255, 255, 255, 255}
	}
	if colorStr[0] == '#' {
		colorStr = colorStr[1:]
	}
	hexLen := len(colorStr)
	var r, g, b, a uint8 = 255, 255, 255, 255
	if hexLen == 8 {
		r = parseHexByte(colorStr[0:2])
		g = parseHexByte(colorStr[2:4])
		b = parseHexByte(colorStr[4:6])
		a = parseHexByte(colorStr[6:8])
	} else if hexLen == 6 {
		r = parseHexByte(colorStr[0:2])
		g = parseHexByte(colorStr[2:4])
		b = parseHexByte(colorStr[4:6])
		a = 255
	} else if hexLen == 4 {
		r = parseHexNibble(colorStr[0]) * 17
		g = parseHexNibble(colorStr[1]) * 17
		b = parseHexNibble(colorStr[2]) * 17
		a = parseHexNibble(colorStr[3]) * 17
	} else if hexLen == 3 {
		r = parseHexNibble(colorStr[0]) * 17
		g = parseHexNibble(colorStr[1]) * 17
		b = parseHexNibble(colorStr[2]) * 17
		a = 255
	}
	return color.RGBA{r, g, b, a}
}

func parseHexByte(s string) uint8 {
	val, _ := strconv.ParseUint(s, 16, 8)
	return uint8(val)
}

func parseHexNibble(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

type pixelTask struct {
	x, z    int
	imgX    int
	isSlime bool
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d)/float64(time.Microsecond))
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))
	} else {
		return fmt.Sprintf("%.2fs", float64(d)/float64(time.Second))
	}
}

func generateFileName(nameTemplate string, x1, z1, x2, z2 int, format string) string {
	if nameTemplate == "" {
		nameTemplate = "{x1}_{z1}_{x2}_{z2}"
	}
	result := nameTemplate
	result = strings.ReplaceAll(result, "{x1}", strconv.Itoa(x1))
	result = strings.ReplaceAll(result, "{z1}", strconv.Itoa(z1))
	result = strings.ReplaceAll(result, "{x2}", strconv.Itoa(x2))
	result = strings.ReplaceAll(result, "{z2}", strconv.Itoa(z2))
	if !strings.HasSuffix(result, "."+format) {
		result = result + "." + format
	}
	return result
}

func generateRingRegions(ringCount, size int, origin []int) [][]int {
	if ringCount < 1 {
		ringCount = 1
	}
	if size < 1 {
		size = 512
	}
	ox, oz := 0, 0
	if len(origin) >= 2 {
		ox, oz = origin[0], origin[1]
	}
	regions := make([][]int, 0, 4*ringCount*ringCount)
	for q := 0; q < 4; q++ {
		sx := 1
		if q == 1 || q == 2 {
			sx = -1
		}
		sz := 1
		if q == 2 || q == 3 {
			sz = -1
		}
		for i := 0; i < ringCount; i++ {
			for j := 0; j < ringCount; j++ {
				var x1, x2, z1, z2 int
				if sx == 1 {
					x1 = ox + i*size
					x2 = ox + i*size + size - 1
				} else {
					x1 = ox - (i+1)*size + 1
					x2 = ox - i*size
				}
				if sz == 1 {
					z1 = oz + j*size
					z2 = oz + j*size + size - 1
				} else {
					z1 = oz - (j+1)*size + 1
					z2 = oz - j*size
				}
				regions = append(regions, []int{x1, z1, x2, z2})
			}
		}
	}
	return regions
}

func GenerateSlimeChunkImage(region []int, slimeColor color.RGBA, outputDir, format, edition, seed string, workers int, outputName string) (time.Duration, error) {
	start := time.Now()
	if len(region) != 4 {
		return 0, fmt.Errorf("invalid region format, expected [x1, z1, x2, z2]")
	}
	x1, z1, x2, z2 := region[0], region[1], region[2], region[3]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if z1 > z2 {
		z1, z2 = z2, z1
	}
	width := x2 - x1 + 1
	height := z2 - z1 + 1
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	isSlimeFunc := func(cx, cz int) bool {
		if strings.ToLower(edition) == "java" {
			return IsSlimeChunkJava(seed, cx, cz)
		}
		return IsSlimeChunkBedrock(cx, cz)
	}
	totalPixels := width * height
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if totalPixels < 10000 {
		workers = 1
	}
	if workers > runtime.NumCPU() {
		workers = runtime.NumCPU()
	}
	numJobs := height
	if workers > numJobs {
		workers = numJobs
	}
	jobsPerWorker := (numJobs + workers - 1) / workers
	var wg sync.WaitGroup
	pixels := make([][]pixelTask, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID, startZ, endZ int) {
			defer wg.Done()
			localPixels := []pixelTask{}
			for z := startZ; z < endZ && z <= z2; z++ {
				for x := x1; x <= x2; x++ {
					if isSlimeFunc(x, z) {
						localPixels = append(localPixels, pixelTask{
							x:       x,
							z:       z,
							imgX:    x - x1,
							isSlime: true,
						})
					}
				}
			}
			pixels[workerID] = localPixels
		}(w, z1+w*jobsPerWorker, z1+(w+1)*jobsPerWorker)
	}
	wg.Wait()
	for _, batch := range pixels {
		for _, p := range batch {
			imgY := p.z - z1
			img.Set(p.imgX, imgY, slimeColor)
		}
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create output directory: %v", err)
	}
	filename := generateFileName(outputName, x1, z1, x2, z2, format)
	filepath := filepath.Join(outputDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		return 0, fmt.Errorf("failed to encode image: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("Generated: %s (%dx%d pixels, %s)\n", filepath, width, height, formatDuration(elapsed))
	return elapsed, nil
}

func main() {
	totalStart := time.Now()
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}
	scriptDir := filepath.Dir(execPath)
	configPath := filepath.Join(scriptDir, "config.toml")
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		fmt.Printf("Warning: Failed to load config.toml: %v\n", err)
		fmt.Println("Using default configuration...")
		config = Config{
			Edition:    "bedrock",
			Seed:       "",
			SlimeColor: "#ffffff",
			OutputDir:  "chunks",
			Format:     "png",
			Regions:    [][]int{{0, 0, 511, 511}},
			Workers:    0,
			OutputName: "{x1}_{z1}_{x2}_{z2}",
			EnableRing: false,
			RingCount:  2,
			RingOrigin: []int{0, 0},
			RingSize:   512,
		}
	}
	if config.Edition == "" {
		config.Edition = "bedrock"
	}
	if config.SlimeColor == "" {
		config.SlimeColor = "#ffffff"
	}
	if config.OutputDir == "" {
		config.OutputDir = "chunks"
	}
	if config.Format == "" {
		config.Format = "png"
	}
	if config.OutputName == "" {
		config.OutputName = "{x1}_{z1}_{x2}_{z2}"
	}
	if config.RingCount < 1 {
		config.RingCount = 2
	}
	if config.RingSize < 1 {
		config.RingSize = 512
	}
	if len(config.RingOrigin) < 2 {
		config.RingOrigin = []int{0, 0}
	}
	if !filepath.IsAbs(config.OutputDir) {
		config.OutputDir = filepath.Join(scriptDir, config.OutputDir)
	}
	slimeColor := ParseColor(config.SlimeColor)
	var regions [][]int
	if config.EnableRing {
		regions = generateRingRegions(config.RingCount, config.RingSize, config.RingOrigin)
	} else {
		if len(config.Regions) == 0 {
			regions = [][]int{{0, 0, 511, 511}}
		} else {
			regions = config.Regions
		}
	}
	workers := config.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	editionDisplay := strings.ToUpper(config.Edition[:1]) + config.Edition[1:]
	seedDisplay := config.Seed
	if seedDisplay == "" {
		seedDisplay = "(not used)"
	}
	fmt.Printf("Slime Chunk Generator\n")
	fmt.Printf("====================\n")
	fmt.Printf("Edition: %s\n", editionDisplay)
	if strings.ToLower(config.Edition) == "java" {
		fmt.Printf("Seed: %s\n", seedDisplay)
	}
	if config.EnableRing {
		fmt.Printf("Mode: Ring (%dx%d per quadrant, %d total)\n", config.RingCount, config.RingCount, 4*config.RingCount*config.RingCount)
		fmt.Printf("Origin: [%d, %d]\n", config.RingOrigin[0], config.RingOrigin[1])
		fmt.Printf("Ring size: %d\n", config.RingSize)
	} else {
		fmt.Printf("Mode: Custom regions\n")
	}
	fmt.Printf("Workers: %d (CPU cores: %d)\n", workers, runtime.NumCPU())
	fmt.Printf("Output directory: %s\n", config.OutputDir)
	fmt.Printf("Output name: %s\n", config.OutputName)
	fmt.Printf("Slime color: %s (RGBA: %d,%d,%d,%d)\n", config.SlimeColor, slimeColor.R, slimeColor.G, slimeColor.B, slimeColor.A)
	fmt.Printf("Format: %s\n", config.Format)
	fmt.Printf("Regions to generate: %d\n\n", len(regions))
	var totalImageTime time.Duration
	for i, region := range regions {
		fmt.Printf("Processing region %d/%d: %v\n", i+1, len(regions), region)
		elapsed, err := GenerateSlimeChunkImage(region, slimeColor, config.OutputDir, config.Format, config.Edition, config.Seed, workers, config.OutputName)
		if err != nil {
			fmt.Printf("Error generating region %v: %v\n", region, err)
		} else {
			totalImageTime += elapsed
		}
	}
	totalElapsed := time.Since(totalStart)
	fmt.Printf("\nTotal time: %s (image generation: %s)\n", formatDuration(totalElapsed), formatDuration(totalImageTime))
}