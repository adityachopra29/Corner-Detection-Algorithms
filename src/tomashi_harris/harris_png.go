// implementation of harris corner detection in go

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"image/draw"
	"os"
	"sort"
	"math"
)


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
			grayVal := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b) / 256) 
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

func saveGrayImage(img *image.Gray, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}


func convolution(img image.Image, kernel [3][3]int) [][]int {

    bounds := img.Bounds()

    newImg := [][] int{};

	// Laplace Kernel
    // kernel := [3][3]int{ 
    //     {1, 1, 1}, 
    //     {1, -8, 1}, 
    //     {1, 1, 1}, 
    // }
	// Taking convolution at each point
	for y := 1; y < bounds.Max.Y-1; y++ {
        for x := 1; x < bounds.Max.X-1; x++ {
			// Convolution with  kernel
            var sum int = 0;
			for i:=-1; i<2; i++ {
				for j:=-1; j<2; j++ {
					r, g, b, _ := img.At(x+i, y+j).RGBA()
					grayVal := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b) / 256) 
					sum += int(grayVal) * kernel[j+1][i+1]
				}
			}
			
			if sum < 0 {
				sum = 0
			} else if sum > 255 {
				sum = 255
			}
            // newImgSet(x, y, color.Gray{Y: uint8(math.Abs(float64(sum)))})
			newImg[x][y] = sum

        }
    }
    return newImg
}

func main(){
	fmt.Println("hey")

	img, err := getPNGImageFromFilePath("mountain.png")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Convert image to grayscale, taking dimensions and fixing region in which corner detection will be performed
	gray := rgbToGray(img)
	bounds := gray.Bounds()
	minX := (bounds.Min.X)/4
	maxX := 3 * (bounds.Max.X)/4
	minY := (bounds.Min.Y)/4
	maxY := 3 * (bounds.Max.Y)/4
	window := 3
	threshold := 10; //threshold to pass for harris 
	minDist := 10; //the minimum distance between any 2 points
	// width := maxX - minX;
	// height := maxY - minY

	Corners := make([][3]float64, 0)


	// apply median filter for salt and pepper noise

	for y:= minY; y < maxY; y++ {
		for x:= minX; x < maxX; x++ {
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
	Ixx := [][] int{};
	Iyy := [][] int{};
	Ixy := [][] int{};

	// dx and dy using Sobel kernels
	dx := (convolution(img, sobelX))
	dy := (convolution(img, sobelY))

	//  Ixx and Iyy (squared gradients)
	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			Ixx[x][y] = dx[x][y] * dx[x][y]
			Iyy[x][y] = dy[x][y] * dy[x][y]
			Ixy[x][y] = dx[x][y] * dy[x][y]
		}
	}
	Sxx := make([][] int, 0)
	Syy := make([][] int, 0)
	Sxy := make([][] int, 0)
	det := make([][] float64, 0)
	trace := make([][] float64, 0)

	// Sum of square gradients in window and finding the corners
	for y := window; y < maxY-window; y++ {
		for x := window; x < maxX-window; x++ {
			for i := 0; i<window; i++{
				for j:=0; j<window; j++{
					Sxx[x][y] += Ixx[x+i][y+j]
					Syy[x][y] += Iyy[x+i][y+j]
					Sxy[x][y] += Ixy[x+i][y+j]
				}
			}
			// Determinant and trace
			det[x][y] = float64((Sxx[x][y] * Syy[x][y]) - (Sxy[x][y] * Sxy[x][y]))
			trace[x][y] = float64(Sxx[x][y] + Syy[x][y])

			// r := min(Sxx[x][y], Syy[x][y])
			k := 0.04 //suitable constant
			r := det[x][y] - k*(trace[x][y])


			if r < float64(threshold){
				Corners = append(Corners, [3]float64{float64(x),float64(y), r})
			}
		}
	}

	sort.Slice(Corners, func(i, j int) bool {
		return Corners[i][2] > Corners[j][2]
	})

	// Filter corners by distance
	eFiltered := [][3]float64{Corners[0]} //the final filtered set of corners
	for _, corner := range Corners {
		bigger := true
		for _, filteredCorner := range eFiltered {
			// Euclidean distance comparison
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

	
	// Make a new image and draw all the points on it
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	pointColor := color.RGBA{255, 0, 0, 255} // Red color
	for _, point := range Corners {
		x, y := point[0], point[1]
		if x >= float64(minX) && x < float64(maxX) && y >= float64(minY) && y < float64(maxY) {
			rgba.Set(int(x), int(y), pointColor)
		}
	}

	outputFile := "modified.png"

	// Save the new image
	outFile, err := os.Create(outputFile)
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

	fmt.Println("Modified image saved as:", outputFile)
}
