package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nikhilm25/Reddit-multimedia-downloader-in-go/internal/helper"
	"github.com/nikhilm25/Reddit-multimedia-downloader-in-go/internal/reddit"
	"github.com/urfave/cli/v2"
)

// handles all the shit we need to run this bad boy
type App struct {
	downloader *helper.Downloader
	cli        *cli.App
}

func NewApp() *App {
	app := &App{
		downloader: helper.NewDownloader(),
	}
	app.setupCLI()
	return app
}

func (a *App) setupCLI() {
	a.cli = &cli.App{
		Name:    "Reddit-multimedia-downloader-in-go",
		Usage:   "A reddit multimedia downloader",
		Version: "0.66.5",
		Flags:   getFlags(),
		Action:  a.handleAction,
	}
}

func (a *App) handleAction(ctx *cli.Context) error {
	url := ctx.String("url")
	if url == "" {
		return cli.ShowAppHelp(ctx)
	}
	return a.controller(url, ctx.Bool("dash"))
}

func (a *App) controller(rawUrl string, useDash bool) error {
	title := fmt.Sprintf("%d", time.Now().Unix())
	data, err := a.fetchRedditData(rawUrl, useDash)
	if err != nil {
		return err
	}
	return a.processMedia(data, title)
}

func main() {
	// default url incase the url flag isnt passed
	var raw_url string

	if len(os.Args) > 1 {
		raw_url = os.Args[1]
	}

	app := NewApp()

	if err := app.cli.Run(os.Args); err != nil {
		fmt.Println()
		helper.ErrorLog.Println(err)
	}
}

// grabs the reddit stuff, might fail but whatever
func (a *App) fetchRedditData(rawUrl string, useDash bool) (*reddit.RedditData, error) {
	body, err := helper.GetJSONBody(rawUrl)
	if err != nil {
		return nil, err
	}

	redditData, err := reddit.ExtractRedditData(body, useDash)
	if err != nil {
		return nil, err
	}

	return redditData, nil
}

// does all the heavy lifting for downloading n stuff
func (a *App) processMedia(redditData *reddit.RedditData, title string) error {
	var wg sync.WaitGroup

	if redditData.IsDash {
		helper.InfoLog.Println("Downloading DASHPlaylist")
		helper.DownloadDashPlaylist(redditData.MediaUrl, title)
		return nil
	}

	if redditData.IsRedditGallery {
		for _, url := range redditData.GalleryUrls {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				helper.Download(url, "", title)
			}(url)
		}
		wg.Wait()
	} else {
		media, audio := helper.GetMediaUrl(redditData.MediaUrl)
		helper.Download(media, audio, title)
	}

	return nil
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "a reddit post url",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "dash",
			Usage: "download reddit video using Dash playlist with ffmpeg",
		},
	}
}
