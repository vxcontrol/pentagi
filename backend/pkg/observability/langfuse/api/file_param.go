package api

import (
	"io"
)

// FileParam is a file type suitable for multipart/form-data uploads.
type FileParam struct {
	io.Reader
	filename    string
	contentType string
}

// FileParamOption adapts the behavior of the FileParam. No options are
// implemented yet, but this interface allows for future extensibility.
type FileParamOption interface {
	apply()
}

// NewFileParam returns a *FileParam type suitable for multipart/form-data uploads. All file
// upload endpoints accept a simple io.Reader, which is usually created by opening a file
// via os.Open.
//
// However, some endpoints require additional metadata about the file such as a specific
// Content-Type or custom filename. FileParam makes it easier to create the correct type
// signature for these endpoints.
func NewFileParam(
	reader io.Reader,
	filename string,
	contentType string,
	opts ...FileParamOption,
) *FileParam {
	return &FileParam{
		Reader:      reader,
		filename:    filename,
		contentType: contentType,
	}
}

func (f *FileParam) Name() string        { return f.filename }
func (f *FileParam) ContentType() string { return f.contentType }
