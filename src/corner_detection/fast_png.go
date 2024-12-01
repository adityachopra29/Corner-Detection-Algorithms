// fast corner detection algorithm in go

package main

import (
	"fmt"
	"image/draw"
	"image/png"

	// "image/jpeg"
	"image"
	"image/color"
	"math"
	"os"
	"sort"
)

func getImageFromFilePath(filePath string) (image.Image, error) {
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
			// Weighted average for luminance perception
			grayVal := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b) / 256) 
			gray.Set(x, y, color.Gray{grayVal})
		}
	}

	return gray
}

func median(values []int) int {
	sort.Ints(values)
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2
	}
	return values[mid]
}

func saveGrayImage(img *image.Gray, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// func savePNGImage(img *draw.Image, filePath string) error {
// 	toimg, _ := os.Create(filePath)
//     defer toimg.Close()
// 	m := image.NewRGBA(image.Rect(0, 0, 800, 600))


//     png.Encode(toimg, m, &png.Options{png.DefaultQuality})

	
// }
func circle(row, col int) [8][2]int {
    point1 := [2]int{row + 3, col}
    point3 := [2]int{row + 3, col - 1}
    point5 := [2]int{row + 1, col + 3}
    point7 := [2]int{row - 1, col + 3}
    point9 := [2]int{row - 3, col}
    point11 := [2]int{row - 3, col - 1}
    point13 := [2]int{row + 1, col - 3}
    point15 := [2]int{row - 1, col - 3}

    return [8][2]int{point1, point3, point5, point7, point9, point11, point13, point15}
}

func adjacencyCheck(p1 [2]int, p2[2]int) bool {
	return (p1[0]-p2[0])*(p1[0]-p2[0])+(p1[1]-p2[1])*(p1[1]-p2[1]) < 9
}

func isCorner(img *image.Gray, row, col int, ROI [8][2]int, threshold int) bool {
	// Central pixel intensity
	I := img.GrayAt(row, col).Y

	// Count pixels that meet the threshold condition
	count := 0
	for _, point := range ROI {
		neighborRow, neighborCol := point[0], point[1]
		if math.Abs(float64(img.GrayAt(neighborRow, neighborCol).Y-I)) > float64(threshold) {
			count++
			if count >= 3 { // Early exit if corner condition is met
				return true
			}
		}
	}

	// A pixel is a corner if at least 3 conditions are met
	return count >= 3
}

func scoreCheck(image *image.Gray, corner [2]int) int {
	// Retrieve the ROI using the circle function
	ROI := circle(corner[0], corner[1])

	// Center pixel intensity
	centerIntensity := int(image.GrayAt(corner[1], corner[0]).Y)

	// Calculate the score
	score := 0
	for _, roiPoint := range ROI {
		x, y := roiPoint[0], roiPoint[1]
		// Boundary check to avoid out-of-bounds access
		if x >= 0 && y >= 0 && x < image.Rect.Max.X && y < image.Rect.Max.Y {
			roiIntensity := int(image.GrayAt(x, y).Y)
			score += int(math.Abs(float64(centerIntensity - roiIntensity)))
		}
	}

	return score
}


func remove(slice [][2]int, s int) [][2]int {
    return append(slice[:s], slice[s+1:]...)
}

func main() {
    img, err := getImageFromFilePath("corner1.png")
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

	// fmt.Println(gray)

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

	
	// Iterate through the image and apply the FAST corner detection algorithm
	Corners := make([][2]int, 0)

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			// Check if the pixel is a corner
			// draw a circle around it
			// check that it has a certain number of pixels with good intensity difference


			ROI := circle(x, y) //     why we are putting ulta here check
			if isCorner(gray, x, y, ROI, 125) {
				Corners = append(Corners, [2]int{x, y})
			}
		}
	}

	// reducedCorners := make([][2]int, 0)
	// Non-maximum suppression
	for i := 1; i < len(Corners)-1; i++{
		if adjacencyCheck(Corners[i], Corners[i+1]) {
			score1 := scoreCheck(gray, Corners[i-1])
			score2 := scoreCheck(gray, Corners[i])

			if score1 < score2 {
				remove(Corners, i-1)
			}else {
				remove(Corners, i)
			}
			
		}else{ 
			i += 1;
			continue
		}
	}

	err = saveGrayImage(gray, "processed_corner3.png")
	if err != nil {
		fmt.Println("Error saving image: ", err)
	}

	fmt.Println(Corners)


	// Create a new image to draw on
	bounds1 := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	pointColor := color.RGBA{255, 0, 0, 255} // Red color
	for _, point := range Corners {
		x, y := point[0], point[1]
		if x >= bounds1.Min.X && x < bounds1.Max.X && y >= bounds1.Min.Y && y < bounds1.Max.Y {
			rgba.Set(x, y, pointColor)
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


