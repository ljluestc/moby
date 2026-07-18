package client

import (
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ImageImportSource holds source information for ImageImport
type ImageImportSource struct {
	Source     io.Reader // Source is the data to send to the server to create this image from. You must set SourceName to "-" to leverage this.
	SourceName string    // SourceName is the name of the image to pull. Set to "-" to leverage the Source attribute.
}

// ImageImportOptions holds information to import images from the client host.
type ImageImportOptions struct {
	Tag      string           // Tag is the name to tag this image with. This attribute is deprecated.
	Message  string           // Message is the message to tag the image with
	Changes  []string         // Changes are the raw changes to apply to this image
	Platform ocispec.Platform // Platform is the target platform of the image
	// SourceHeaders are HTTP headers to be forwarded verbatim to the source
	// URL when the image is imported from a remote URL (i.e. fromSrc is not
	// "-"). This allows importing from HTTP-authenticated endpoints such as
	// OpenStack Glance, where every request must carry an X-Auth-Token header.
	// The map is only consulted when SourceName in ImageImportSource is a URL;
	// it is silently ignored for stdin imports.
	SourceHeaders map[string][]string
}
