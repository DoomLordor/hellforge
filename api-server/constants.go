package apiserver

const (
	plugin = `const UrlMutatorPlugin = (system) => ({
  rootInjects: {
    setBasePath: (basePath) => {
      const jsonSpec = system.getState().toJSON().spec.json;
      const newJsonSpec = Object.assign({}, jsonSpec, { basePath });

      return system.specActions.updateJsonSpec(newJsonSpec);
    }
  }
});`

	swaggerDefault = `{
  "swagger": "2.0",
  "info": {
    "description": "API documentation",
    "version": "1.0.0",
    "title": "API"
  },
  "basePath": "/",
  "produces": [
	"application/json"
  ],
  "consumes": [
    "application/json"
  ],
  "paths": {},
  "definitions": {},
  "parameters": {},
  "responses": {}
}`
)

// Content-Type header value
const (
	ContentTypeJSON                = "application/json"
	ContentTypeXML                 = "application/xml"
	ContentTypePDF                 = "application/pdf"
	ContentTypeZIP                 = "application/zip"
	ContentTypeGZIP                = "application/gzip"
	ContentTypeMicrosoftExcel      = "application/vnd.ms-excel"
	ContentTypeMicrosoftPowerPoint = "application/vnd.ms-powerpoint"
	ContentTypeMicrosoftWord       = "application/msword"
	ContentTypeOpenDocumentText    = "application/vnd.oasis.opendocument.text"
	ContentTypeOpenDocumentSheet   = "application/vnd.oasis.opendocument.spreadsheet"

	// Text types
	ContentTypeHTML       = "text/html"
	ContentTypePlainText  = "text/plain"
	ContentTypeCSS        = "text/css"
	ContentTypeCSV        = "text/csv"
	ContentTypeJavaScript = "text/javascript"

	// Image types
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeGIF  = "image/gif"
	ContentTypeSVG  = "image/svg+xml"
	ContentTypeWEBP = "image/webp"
	ContentTypeICO  = "image/x-icon"
	ContentTypeBMP  = "image/bmp"
	ContentTypeTIFF = "image/tiff"

	// Audio/video types
	ContentTypeMPEG      = "audio/mpeg"
	ContentTypeWAV       = "audio/wav"
	ContentTypeOGG       = "audio/ogg"
	ContentTypeWEBM      = "video/webm"
	ContentTypeMP4       = "video/mp4"
	ContentTypeQuickTime = "video/quicktime"

	// Font types
	ContentTypeTrueType = "font/ttf"
	ContentTypeOpenType = "font/otf"
	ContentTypeWOFF     = "font/woff"
	ContentTypeWOFF2    = "font/woff2"

	// Multipart types
	ContentTypeMultipartFormData = "multipart/form-data"
	ContentTypeMultipartMixed    = "multipart/mixed"

	// Application types
	ContentTypeOctetStream = "application/octet-stream"
	ContentTypeRSS         = "application/rss+xml"
	ContentTypeAtom        = "application/atom+xml"
	ContentTypeJSONLD      = "application/ld+json"
	ContentTypeWASM        = "application/wasm"
	ContentTypeSQL         = "application/sql"
)

type DataType int

const (
	DataTypeBytes DataType = iota
	DataTypeString
	DataTypeStructJSON
	DataTypeStructXML
)
