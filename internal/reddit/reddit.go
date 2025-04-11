package reddit

import (
	"encoding/json"
	"strings"

	"github.com/nikhilm25/Reddit-multimedia-downloader-in-go/internal/helper"
)

// gets all the good stuff from reddit json, its a mess but it works
func ExtractRedditData(body []byte, useDash bool) (RedditData, error) {
	var (
		dataDump   Reddit
		redditData RedditData
	)

	if err := json.Unmarshal(body, &dataDump); err != nil {
		return redditData, err
	}

	if len(dataDump) == 0 || len(dataDump[0].Data.Children) == 0 {
		return redditData, nil
	}

	data := dataDump[0].Data.Children[0].Data
	return extractMediaData(data, useDash)
}

// figures out what kinda media we got, theres like a million types ffs
func extractMediaData(data struct {
	MediaMetadata       mediaMetadata
	SecureMedia         secureMedia
	CrossPost           []*crosspostParentList
	Preview             *preview
	URLOverriddenByDest string
}, useDash bool) (RedditData, error) {

	var redditData RedditData

	// Handle DASH video
	if useDash && data.SecureMedia.RedditVideo != nil {
		return handleDashVideo(data.SecureMedia.RedditVideo)
	}

	// Handle Reddit Gallery
	if data.MediaMetadata != nil {
		return handleGallery(data.MediaMetadata)
	}

	// Handle Crosspost
	if data.CrossPost != nil && data.CrossPost[0].SecureMedia != nil {
		return handleCrosspost(data.CrossPost[0])
	}

	// Handle External Providers
	if data.SecureMedia.Oembed != nil {
		return handleExternalProvider(data.SecureMedia.Oembed)
	}

	// Handle Preview Media
	if data.Preview != nil {
		return handlePreview(data.Preview)
	}

	// Handle Regular Media
	return handleRegularMedia(data.SecureMedia, data.URLOverriddenByDest)
}

// handles that fancy dash video stuff, its pretty neat
func handleDashVideo(video *redditVideo) (RedditData, error) {
	var redditData RedditData
	redditData.MediaUrl = strings.Split(video.DASH, "?")[0]
	redditData.IsDash = true
	return redditData, nil
}

func handleGallery(mediaMetadata mediaMetadata) (RedditData, error) {
	var redditData RedditData
	for _, v := range mediaMetadata {
		image_url := v.S.URL
		url := strings.ReplaceAll(image_url, "amp;", "")
		redditData.IsRedditGallery = true
		redditData.GalleryUrls = append(redditData.GalleryUrls, url)
	}
	return redditData, nil
}

func handleCrosspost(crossPost *crosspostParentList) (RedditData, error) {
	var redditData RedditData
	if crossPost.SecureMedia.RedditVideo != nil {
		redditData.MediaUrl = crossPost.SecureMedia.RedditVideo.FallbackURL
	}
	return redditData, nil
}

// deals with those annoying gfycat links n other external bs
func handleExternalProvider(oembed *oembed) (RedditData, error) {
	var redditData RedditData
	gfycat := "https://gfycat.com"
	provider_url := oembed.ProviderURL

	switch provider_url {
	case gfycat:
		url := strings.ReplaceAll(oembed.ThumbnailURL, "size_restricted.gif", "mobile.mp4")
		redditData.MediaUrl = url
		return redditData, nil
	default:
		helper.ErrorLog.Printf("%s is not a supported provider, going to fallback options", provider_url)
	}
	return redditData, nil
}

func handlePreview(preview *preview) (RedditData, error) {
	var redditData RedditData

	if preview.Video != nil {
		redditData.MediaUrl = preview.Video.FallbackURL
		return redditData, nil
	}

	if preview.Images[0].Variants.GIF != nil {
		redditData.MediaUrl = preview.Images[0].Variants.GIF.Source.URL
		return redditData, nil
	}

	return redditData, nil
}

func handleRegularMedia(secureMedia secureMedia, urlOverriddenByDest string) (RedditData, error) {
	var redditData RedditData

	if secureMedia.RedditVideo == nil && secureMedia.Oembed == nil {
		redditData.MediaUrl = urlOverriddenByDest
		return redditData, nil
	}

	if secureMedia.RedditVideo != nil {
		redditData.MediaUrl = secureMedia.RedditVideo.FallbackURL
		return redditData, nil
	}

	return redditData, nil
}
