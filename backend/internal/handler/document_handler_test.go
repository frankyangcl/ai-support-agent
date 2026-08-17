package handler_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/auth0/go-jwt-middleware/v3/core"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func documentUploadRouter(maxMB int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewDocumentHandler(nil, maxMB)
	r.POST("/api/documents/upload", func(c *gin.Context) {
		claims := &validator.ValidatedClaims{}
		claims.RegisteredClaims.Subject = "auth0|owner-a"
		c.Request = c.Request.WithContext(core.SetClaims(c.Request.Context(), claims))
		c.Next()
	}, h.UploadDocument)
	return r
}

func multipartUpload(t *testing.T, filename, contentType string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = w.Close()
	return body, w.FormDataContentType()
}

func performUpload(r http.Handler, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/documents/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUploadMissingFile(t *testing.T) {
	w := performUpload(documentUploadRouter(10), bytes.NewBufferString(""), "multipart/form-data; boundary=x")
	if w.Code != 400 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestUploadRejectsInvalidExtension(t *testing.T) {
	body, ct := multipartUpload(t, "notes.txt", "text/plain", []byte("hello"))
	w := performUpload(documentUploadRouter(10), body, ct)
	if w.Code != 400 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestUploadRejectsPathTraversal(t *testing.T) {
	body, ct := multipartUpload(t, `..\secret.pdf.exe`, "application/pdf", []byte("%PDF-test"))
	w := performUpload(documentUploadRouter(10), body, ct)
	if w.Code != 400 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestUploadRejectsSpoofedPDF(t *testing.T) {
	body, ct := multipartUpload(t, "fake.pdf", "application/pdf", []byte("not a pdf"))
	w := performUpload(documentUploadRouter(10), body, ct)
	if w.Code != 400 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestUploadRejectsOversize(t *testing.T) {
	body, ct := multipartUpload(t, "large.pdf", "application/pdf", append([]byte("%PDF-"), make([]byte, 2<<20)...))
	w := performUpload(documentUploadRouter(1), body, ct)
	if w.Code != 413 {
		t.Fatalf("got %d", w.Code)
	}
}
