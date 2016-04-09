package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"

	"github.com/koffeinsource/kaffeeshare/data"
)

// Required to be able to pass different kind of headers in the following functions
type emailHeader interface {
	Get(key string) string
}

// extracts the body of an email
func extractBody(con *data.Context, header emailHeader, bodyReader io.Reader) (*body, error) {
	contentType := header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	if len(mediaType) >= len(contentTypeText) && mediaType[:len(contentTypeText)] == contentTypeText {
		con.Log.Debugf("extractBody: found text")
		return extractTextBody(con, header, bodyReader)
	}

	if len(mediaType) >= len(contentTypeMulti) && mediaType[:len(contentTypeMulti)] == contentTypeMulti {
		con.Log.Debugf("extractBody: multipart")
		return extractMimeBody(con, params["boundary"], bodyReader)
	}

	if len(mediaType) >= len(contentTypeSMIMESig) && mediaType[:len(contentTypeSMIMESig)] == contentTypeSMIMESig {
		con.Log.Debugf("extractAttachment: SMIME sig; ignoring; mediaType: %v", mediaType)
		return nil, nil
	}

	if len(mediaType) >= len(contentTypeImage) && mediaType[:len(contentTypeImage)] == contentTypeImage {
		con.Log.Debugf("extractBody: image, ignoring")
		return nil, nil
	}

	return nil, fmt.Errorf("Unsupported content type: %s; media type: %v", contentType, mediaType)
}

// read through the varios multiple parts
func extractMimeBody(con *data.Context, boundary string, bodyReader io.Reader) (*body, error) {
	var withError *body // stores an email parse with error

	mimeReader := multipart.NewReader(bodyReader, boundary)

	for {
		part, err := mimeReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		defer part.Close()

		result, err := extractBody(con, part.Header, part)

		// this means we tried to decode it, but are not sure
		// lets save this result and try the other parts before return this result
		if result != nil && err != nil {
			withError = result
			con.Log.Infof("extractMimeBody: email guess with error %v", err)
			continue
		}

		if result != nil && result.ContentType[:4] == contentTypeText {
			return result, nil
		}
	}

	if withError != nil {
		return withError, nil
	}

	return nil, fmt.Errorf("Could not parse any of the multiple parts:")
}

// Decode body text and store it in a string
func extractTextBody(con *data.Context, header emailHeader, bodyReader io.Reader) (*body, error) {
	var returnee body

	s, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	con.Log.Debugf("header: %v", header)

	encoding := header.Get("Content-Transfer-Encoding")
	con.Log.Debugf("extractTextBody encoding: %v", encoding)

	if encoding == "base64" {
		b, err := base64.StdEncoding.DecodeString(string(s))
		if err != nil {
			return nil, err
		}
		returnee.Body = string(b)
		returnee.ContentType = header.Get("Content-Type")
		return &returnee, nil
	}

	if encoding == "7bit" {
		// that is just US ASCII (7bit)
		// https://stackoverflow.com/questions/25710599/content-transfer-encoding-7bit-or-8-bit
		returnee.Body = string(s)
		returnee.ContentType = header.Get("Content-Type")
		return &returnee, nil
	}

	if encoding == "quoted-printable" {
		// https://stackoverflow.com/questions/24883742/how-to-decode-mail-body-in-go
		r := quotedprintable.NewReader(bytes.NewReader(s))
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		returnee.Body = string(b)
		returnee.ContentType = header.Get("Content-Type")
		return &returnee, nil
	}

	// ok, let's guess this is just plain text and put it into a string
	returnee.Body = string(s)
	returnee.ContentType = header.Get("Content-Type")

	return &returnee, fmt.Errorf("Unsupported Content-Transfer-Encoding: %v", encoding)
}
