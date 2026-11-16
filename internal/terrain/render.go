package terrain

import (
    "fmt"
    "image"
    "image/png"
    "os"
    "runtime"
    "sync"
)


func ExportPNG(world World, types []TerrainType, path string) error {
    expectedLen := world.Width * world.Height
    if len(types) != expectedLen {
        return fmt.Errorf("несоответствие размеров: типы=%d, ожидаемо=%d", len(types), expectedLen)
    }

    img := image.NewRGBA(image.Rect(0, 0, world.Width, world.Height))


    numWorkers := runtime.NumCPU()
    if numWorkers < 1 {
        numWorkers = 1
    }
    
    rowsPerWorker := world.Height / numWorkers
    if rowsPerWorker < 1 {
        rowsPerWorker = 1
    }
    
    var wg sync.WaitGroup
    
    for workerID := 0; workerID < numWorkers; workerID++ {
        wg.Add(1)
        go func(startRow, endRow int) {
            defer wg.Done()
            for y := startRow; y < endRow; y++ {
                for x := 0; x < world.Width; x++ {
                    idx := y*world.Width + x
                    props, err := TerrainProps(types[idx])
                    if err != nil {
                        continue 
                    }
                    img.Set(x, y, props.Color)
                }
            }
        }(workerID*rowsPerWorker, min((workerID+1)*rowsPerWorker, world.Height))
    }
    
    wg.Wait()

    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("не удалось создать файл %s: %w", path, err)
    }
    defer file.Close()

    if err := png.Encode(file, img); err != nil {
        return fmt.Errorf("ошибка кодирования PNG: %w", err)
    }

    return file.Close()
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}