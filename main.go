package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

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

type mapKey struct {
	Key1 string
	Key2 int
}

func TestFunc(scrs []ScreenInfo) map[string][]int {
	r := map[string][]int{}
	dirs := [7]string{"left", "right", "up", "down", "next", "prev", "fit"}
	for _, k := range dirs {
		s := []int{}
		for i := 0; i < len(scrs); i++ {
			s = append(s, 0)
		}
		r[k] = s
	}
	for ia, sa := range scrs {
		for ib, sb := range scrs {
			if sa != sb {
				if IsectArea(ScreenInfo{
					W: sa.W,
					H: sa.H,
					X: sa.X + sa.W,
					Y: sa.Y,
				}, sb) != 0 {
					r["right"][ia] = ib
					r["left"][ib] = ia
				}
				if IsectArea(ScreenInfo{W: sa.W,
					H: sa.H,
					X: sa.X,
					Y: sa.Y + sa.H,
				}, sb) != 0 {
					r["down"][ia] = ib
					r["up"][ib] = ia
				}
			}
		}
		r["next"][ia] = (ia + 1) % len(scrs)
		r["prev"][ia] = (ia - 1) % len(scrs)
		r["fit"][ia] = ia
	}
	return r
}

func GetWinIdList() []string {
	listId := []string{}
	out, err := exec.Command("xprop", "-root", "_NET_ACTIVE_WINDOW").Output()
	if err != nil {
		panic(err)
	}
	r := regexp.MustCompile("window id # (0x[0-9a-f]+)")
	listId = append(listId, r.FindStringSubmatch(string(out))[1])
	return listId
}

func GetWindowInfo(listId []string, dir string) {
	for _, id := range listId {
		geoStr := [...]string{
			"Width:", "Height:",
			"Absolute upper-left X:", "Absolute upper-left Y:",
			"Relative upper-left X:", "Relative upper-left Y:", "",
		}
		stateStr := map[string]struct{}{
			"Maximized Vert": struct{}{}, "Maximized Horz": struct{}{}, "Fullscreen": struct{}{},
		}
		geo := [6]int{}
		out, err := exec.Command("xwininfo", "-id", id, "-all").Output()
		if err != nil {
			panic(err)
		}
		state := []string{}

		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			idx := len(geoStr)
			// state := []string{}
			for i, s := range geoStr {
				if strings.HasPrefix(line, s) {
					idx = i
					break
				}
			}
			if geoStr[idx] != "" {
				splitedList := strings.Fields(line)
				v, err := strconv.Atoi(splitedList[len(splitedList)-1])
				if err != nil {
					panic(err)
				}
				geo[idx] = v
			} else if _, ok := stateStr[line]; ok {
				state = append(state, strings.Replace(strings.ToLower(line), " ", "_", -1))
			} else if line == "Desktop" {
				os.Exit(2)
			}
		}
		areas := []int{}
		g := ScreenInfo{
			W: geo[0],
			H: geo[1],
			X: geo[2],
			Y: geo[3],
		}
		srcs := GetScreenInformation()
		for _, s := range srcs {
			areas = append(areas, IsectArea(g, s))
		}
		sidx, err := Index(areas, ArrMax(append(areas, 1)))
		if err != nil {
			os.Exit(3)
		}

		r := TestFunc(srcs)
		if sidx > len(srcs) {
			panic(fmt.Errorf("the index is out of range, ${len(src)}"))
		}
		nscr := srcs[r[dir][sidx]]

		npos := []int{geo[2] - geo[4], geo[3] - geo[5]}
		nsiz := geo[0:2]

		if dir == "fit" {
			// TODO make this
			// fmt.Println("fit")
		} else {
			idx := -1
			dirStr := [7]string{"left", "right", "up", "down", "next", "prev", "fit"}
			for i, v := range dirStr {
				if v == dir {
					idx = i / 2
				}
			}
			if idx < 0 {
				panic(fmt.Errorf("direction is invalid"))
			}
			if idx == 0 {
				npos[0] += nscr.X - srcs[sidx].X
			} else if idx == 1 {
				npos[1] += nscr.Y - srcs[sidx].Y
			} else if idx == 2 {
				npos[0] += nscr.X - srcs[sidx].X
				npos[1] += nscr.Y - srcs[sidx].Y
			}
		}

		cmdOpsiton := [][]string{}
		for _, v := range state {
			cmdOpsiton = append(cmdOpsiton, []string{"-b", "toggle," + v})
		}
		wmctrl(id, cmdOpsiton)

		cmdOpsiton2 := [][]string{}
		ntuple := append(npos, nsiz...)
		cmdOpsiton2 = append(cmdOpsiton2, []string{"-e", fmt.Sprintf("0,%d,%d,%d,%d", ntuple[0], ntuple[1], ntuple[2], ntuple[3])})
		wmctrl(id, cmdOpsiton2)

		wmctrl(id, cmdOpsiton)
	}
}

func wmctrl(id string, ops [][]string) {
	baseCmd := []string{"wmctrl", "-i", "-r"}
	for _, op := range ops {
		cmd := append(baseCmd, id)
		cmd = append(cmd, op...)
		exec.Command(cmd[0], cmd[1:]...).Run()
	}

}

func Index(arr []int, n int) (int, error) {
	for i, v := range arr {
		if v == n {
			return i, nil
		}
	}
	err := fmt.Errorf(`the query ${n} not in arr`)
	return -1, err
}

func ArrMax(arr []int) int {
	res := arr[0]
	for i := 1; i < len(arr); i++ {
		if arr[i] > res {
			res = arr[i]
		}
	}
	return res
}
func main() {
	app := &cli.App{
		Name:      "movescreen",
		Usage:     "Move the screen to adjacent window.",
		UsageText: "movescreen [-r|--ratio] <left|right|up|down|next|prev|fit>",

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "ratio",
				Aliases: []string{"r"},
				Usage:   "keep the ratio of the window.",
			},
		},
		Action: func(c *cli.Context) error {
			// fmt.Println(c.Args())
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
				// fmt.Println(c.NArg())
				panic("hogehoge")
			}
			_, ok := dirStr[c.Args().First()]
			if !ok {
				panic("dir str is invalid")
			}
			// GetScreenInformation()

			GetWindowInfo(GetWinIdList(), c.Args().First())
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
