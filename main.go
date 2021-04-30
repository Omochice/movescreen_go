package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

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
			name := "Nefertii"
			if c.NArg() < 1 {
				fmt.Println(c.NArg())
				panic("hogehoge")
			}
			_, ok := dirStr[c.Args().First()]
			if !ok {
				panic("dir str is invalid")
			}
			if c.String("lang") == "spanigh" {
				fmt.Println("hola", name)
			} else {
				fmt.Println("hello", name)
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
