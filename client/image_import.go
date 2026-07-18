package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/distribution/reference"
)

// importSrcHeadersKey is the HTTP request header used to forward custom source
// URL headers from the client to the daemon during image import.
// The value is a base64-encoded JSON-marshalled map[string][]string.
const importSrcHeadersKey = "X-Import-Src-Headers"

// ImageImportResult holds the response body returned by the daemon for image import.
type ImageImportResult interface {
	io.ReadCloser
}

// ImageImport creates a new image based on the source options. It returns the
// JSON content in the [ImageImportResult].
//
// The underlying [io.ReadCloser] is automatically closed if the context is canceled,
func (cli *Client) ImageImport(ctx context.Context, source ImageImportSource, ref string, options ImageImportOptions) (ImageImportResult, error) {
	if ref != "" {
		// Check if the given image name can be resolved
		if _, err := reference.ParseNormalizedNamed(ref); err != nil {
			return nil, err
		}
	}

	query := url.Values{}
	if source.SourceName != "" {
		query.Set("fromSrc", source.SourceName)
	}
	if ref != "" {
		query.Set("repo", ref)
	}
	if options.Tag != "" {
		query.Set("tag", options.Tag)
	}
	if options.Message != "" {
		query.Set("message", options.Message)
	}
	if p := formatPlatform(options.Platform); p != "unknown" {
		// TODO(thaJeztah): would we ever support multiple platforms here? (would require multiple rootfs tars as well?)
		query.Set("platform", p)
	}
	for _, change := range options.Changes {
		query.Add("changes", change)
	}

	var headers http.Header
	if len(options.SourceHeaders) > 0 && source.SourceName != "-" {
		encoded, err := encodeImportSrcHeaders(options.SourceHeaders)
		if err != nil {
			return nil, err
		}
		headers = http.Header{importSrcHeadersKey: []string{encoded}}
	}

	resp, err := cli.postRaw(ctx, "/images/create", query, source.Source, headers)
	if err != nil {
		return nil, err
	}
	return &imageImportResult{
		ReadCloser: newCancelReadCloser(ctx, resp.Body),
	}, nil
}

// ImageImportResult holds the response body returned by the daemon for image import.
type imageImportResult struct {
	io.ReadCloser
}

var (
	_ io.ReadCloser     = (*imageImportResult)(nil)
	_ ImageImportResult = (*imageImportResult)(nil)
)

// encodeImportSrcHeaders serialises a map[string][]string as base64-encoded
// JSON for transmission in the X-Import-Src-Headers request header.
func encodeImportSrcHeaders(headers map[string][]string) (string, error) {
	b, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// decodeImportSrcHeaders is the inverse of encodeImportSrcHeaders.
func decodeImportSrcHeaders(encoded string) (map[string][]string, error) {
	b, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var headers map[string][]string
	if err := json.Unmarshal(b, &headers); err != nil {
		return nil, err
	}
	return headers, nil
}
