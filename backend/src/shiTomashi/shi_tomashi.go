// implementation of harris corner detection in go

package shiTomashi

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"sort"
)

func getJPGImageFromFilePath(filePath string) (image.Image, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    image, err := jpeg.Decode(f)
    return image, err
}

func getPNGImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, err := png.Decode(f)
	return image, err
}

func rgbToGray(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Weighted average for luminance perception to convert rgb to gray
			grayVal := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)/256)
			gray.Set(x, y, color.Gray{grayVal})
		}
	}
	return gray
}

func median(arr []int) int {
	sort.Ints(arr)
	mid := len(arr) / 2
	if len(arr)%2 == 0 {
		return (arr[mid-1] + arr[mid]) / 2
	}
	return arr[mid]
}

// func saveGrayImage(img *image.Gray, filePath string) error {
// 	f, err := os.Create(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()
// 	return png.Encode(f, img)
// }

func convolution(img image.Image, kernel [3][3]int, minX int, minY int, maxX int, maxY int) [][]int {

	newImg := make([][]int, maxY)
	for i := range newImg {
		newImg[i] = make([]int, maxX)
	}

	// Taking convolution at each point
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			// Convolution with  kernel
			var sum int = 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					// Handle boundary conditions
					imgX := x + i
					imgY := y + j

					if imgX >= minX && imgX < maxX && imgY >= minY && imgY < maxY {
						r, g, b, _ := img.At(imgX, imgY).RGBA()
						grayVal := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)/256)
						sum += int(grayVal) * kernel[j+1][i+1]
						// temp = append(temp, int(px))
					}
				}
			}

			if sum < 0 {
				sum = 0
			} else if sum > 255 {
				sum = 255
			}
			newImg[y][x] = sum
		}
	}
	return newImg
}

func make2DSlice(rows, cols int) [][]float64 {
	slice := make([][]float64, rows)
	for i := range slice {
		slice[i] = make([]float64, cols)
	}
	return slice
}

func ShiTomashi(inputPath, outputPath string) {
    img, err := getJPGImageFromFilePath(inputPath)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Convert image to grayscale, taking dimensions and fixing region in which corner detection will be performed
	gray := rgbToGray(img)
	bounds := gray.Bounds()
	minX := (bounds.Min.X) / 4
	maxX := 3 * (bounds.Max.X) / 4
	minY := (bounds.Min.Y) / 4
	maxY := 3 * (bounds.Max.Y) / 4
	window := 3
	threshold := 10 //threshold to pass for harris
	minDist := 10   //the minimum distance between any 2 points

	Corners := make([][3]float64, 0)

	// apply median filter for salt and pepper noise

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			temp := make([]int, 0, 9)
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					px := gray.GrayAt(x+i, y+j).Y
					temp = append(temp, int(px))
				}
			}
			medianValue := median(temp)
			gray.SetGray(x, y, color.Gray{uint8(medianValue)})
		}
	}

	// applying harris corner detection algorithm on each point in image
	// finding derivatives of each point and making differential square matrix

	// Laplace Kernel
	// kernel := [3][3]int{
	//     {1, 1, 1},
	//     {1, -8, 1},
	//     {1, 1, 1},
	// }

	// sobel matrices that are used to calculate derivative matrices using convolution
	sobelX := [3][3]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	sobelY := [3][3]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	dx := convolution(gray, sobelX, minX, minY, maxX, maxY)
	dy := convolution(gray, sobelY, minX, minY, maxX, maxY)


	Ixx := make2DSlice(maxY, maxX)
	Iyy := make2DSlice(maxY, maxX)
	Ixy := make2DSlice(maxY, maxX)

	// Compute gradient products
	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			Ixx[y][x] = float64(dx[y][x] * dx[y][x])
			Iyy[y][x] = float64(dy[y][x] * dy[y][x])
			Ixy[y][x] = float64(dx[y][x] * dy[y][x])
		}
	}

	// Sum gradients within a window and calculate minimum eigenvalue
	for y := window; y < maxY-window; y++ {
		for x := window; x < maxX-window; x++ {
			var sumIxx, sumIyy, sumIxy float64
			for i := -window / 2; i <= window/2; i++ {
				for j := -window / 2; j <= window/2; j++ {
					sumIxx += Ixx[y+i][x+j]
					sumIyy += Iyy[y+i][x+j]
					sumIxy += Ixy[y+i][x+j]
				}
			}

			// Calculate eigenvalues of the structure tensor
			trace := sumIxx + sumIyy
			det := (sumIxx * sumIyy) - (sumIxy * sumIxy)
			eigen1 := (trace + math.Sqrt(trace*trace-4*det)) / 2
			eigen2 := (trace - math.Sqrt(trace*trace-4*det)) / 2

			// Use the minimum eigenvalue as the response
			response := math.Min(eigen1, eigen2)

			if response > float64(threshold) {
				Corners = append(Corners, [3]float64{float64(x), float64(y), response})
			}
		}
	}

	// Sort corners by response value
	sort.Slice(Corners, func(i, j int) bool {
		return Corners[i][2] > Corners[j][2]
	})

	// Filter corners by minimum distance
	eFiltered := [][3]float64{Corners[0]}
	for _, corner := range Corners {
		bigger := true
		for _, filteredCorner := range eFiltered {
			distance := math.Sqrt(math.Pow(corner[0]-filteredCorner[0], 2) + math.Pow(corner[1]-filteredCorner[1], 2))
			if distance <= float64(minDist) {
				bigger = false
				break
			}
		}
		if bigger {
			eFiltered = append(eFiltered, corner)
		}
	}

	// Draw corners on the image
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	pointColor := color.RGBA{255, 0, 0, 255} // Red
	for _, point := range eFiltered {
		x, y := point[0], point[1]
		rgba.Set(int(x), int(y), pointColor)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outFile.Close()

	err = png.Encode(outFile, rgba)
	if err != nil {
		fmt.Println("Error encoding image:", err)
		return
	}

	fmt.Println("Modified image saved as:", outputPath)
}
