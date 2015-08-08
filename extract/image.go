package extract

import (
	"strings"

	"github.com/koffeinsource/kaffeeshare/data"
	"github.com/koffeinsource/kaffeeshare/request"
)

func image(i *data.Item, sourceURL string, contentType string, log request.Context) {
	if !(strings.Index(contentType, "image/") == 0) {
		return
	}

	log.Infof("Running Image plugin.")

	i.ImageURL = ""
	i.Caption = sourceURL[strings.LastIndex(sourceURL, "/")+1:]
	i.Description = "<img src=\"" + sourceURL + "\">"
}
