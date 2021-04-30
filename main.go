package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/urfave/cli/v2"
)

type ScreenInfo struct {
	W int
	H int
	X int
	Y int
}

func GetScreenInformation() []ScreenInfo {
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		panic(err)
	}
	r := regexp.MustCompile(` connected( primary)? (([0-9]+)x([0-9]+)\+([0-9]+)\+([0-9]+))`)
	results := []ScreenInfo{}
	for _, l := range r.FindAllStringSubmatch(string(out), -1) {
		w, _ := strconv.Atoi(l[3][:])
		h, _ := strconv.Atoi(l[4][:])
		x, _ := strconv.Atoi(l[5][:])
		y, _ := strconv.Atoi(l[6][:])
		results = append(results, ScreenInfo{
			W: w,
			H: h,
			X: x,
			Y: y,
		})
	}
	return results
}

func Min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func IsectArea(a ScreenInfo, b ScreenInfo) int {
	tlx := Max(a.X, b.X)
	tly := Max(a.Y, b.Y)
	brx := Min(a.X+a.W, b.X+b.W)
	bry := Min(a.Y+a.H, b.Y+b.H)
	return Max(0, brx-tlx) * Max(0, bry-tly)
}

func main() {
	app := &cli.App{
		Name:      "Title",
		Usage:     "hogehgoe",
		UsageText: "konnnitiha",

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "ratio",
				Aliases: []string{"r"},
				Usage:   "this is ratio option",
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println(c.Args())
			dirStr := map[string]struct{}{
				"left":  struct{}{},
				"right": struct{}{},
				"up":    struct{}{},
				"down":  struct{}{},
				"next":  struct{}{},
				"prev":  struct{}{},
				"fit":   struct{}{},
			}
			// name := "Nefertii"
			if c.NArg() < 1 {
				fmt.Println(c.NArg())
				panic("hogehoge")
			}
			_, ok := dirStr[c.Args().First()]
			if !ok {
				panic("dir str is invalid")
			}
			GetScreenInformation()
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
