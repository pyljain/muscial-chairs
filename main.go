package main

import (
	"context"
	"log"
	"mc/internal/actions"
	"os"

	"github.com/urfave/cli/v3"
)

// main is the entry point for the application.
// This function parses flags and launches the server.
func main() {
	cmd := &cli.Command{
		Name:  "musicalchair",
		Usage: "Keep a defined number of workers ready to receive requests and manage task assignment",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "connection-string",
						Aliases: []string{"c"},
						Value:   "",
						Usage:   "Mongo connection string",
					},
					&cli.IntFlag{
						Name:  "worker-count",
						Value: 3,
						Usage: "Number of workers to run",
					},
					&cli.IntFlag{
						Name:  "threshold",
						Value: 2,
						Usage: "Number of additional workers (surge) that may be running",
					},
				},
				Action: actions.Run,
			},
			// {
			// 	Name:  "run",
			// 	Usage: "Run the server",
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name:    "connection-string",
			// 			Aliases: []string{"c"},
			// 			Value:   "",
			// 			Usage:   "Mongo connection string",
			// 		},
			// 		&cli.IntFlag{
			// 			Name:  "worker-count",
			// 			Value: 3,
			// 			Usage: "Number of workers to run",
			// 		},
			// 	},
			// },
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
