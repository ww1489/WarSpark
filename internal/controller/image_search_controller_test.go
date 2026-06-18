package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
)

func TestImageSearchControllerCreateJobMapsMultipartUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeImageSearchService{
		createResult: imagesearchdomain.Job{
			ID:           "job_123",
			SearchStatus: "created",
			UploadedImage: &imagesearchdomain.UploadedImage{
				ID:               "img_123",
				ImageURL:         "/uploads/image-search/job_123/img_123.png",
				Width:            512,
				Height:           512,
				Normalized:       true,
				RawRetentionDays: 7,
			},
			CreatedAt: time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
		},
	}
	controller := NewImageSearchController(service)

	router := gin.New()
	router.POST("/api/v1/image-search/jobs", controller.CreateJob)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "base.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(testControllerPNG(t)); err != nil {
		t.Fatalf("write png: %v", err)
	}
	_ = writer.WriteField("war_target_id", "target_123")
	_ = writer.WriteField("enemy_name", "No. 7")
	_ = writer.WriteField("expected_th", "16")
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/image-search/jobs", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.RemoteAddr = "203.0.113.7:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.createInput.FileName != "base.png" {
		t.Fatalf("expected file name base.png, got %q", service.createInput.FileName)
	}
	if service.createInput.ClientIP != "203.0.113.7" {
		t.Fatalf("expected client IP from RemoteAddr, got %q", service.createInput.ClientIP)
	}
	if service.createInput.TargetContext.WarTargetID != "target_123" {
		t.Fatalf("expected war target context, got %#v", service.createInput.TargetContext)
	}
	if service.createInput.TargetContext.ExpectedTH == nil || *service.createInput.TargetContext.ExpectedTH != 16 {
		t.Fatalf("expected TH 16 context, got %#v", service.createInput.TargetContext.ExpectedTH)
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			ID            string `json:"job_id"`
			SearchStatus  string `json:"search_status"`
			UploadedImage struct {
				ID         string `json:"image_id"`
				Normalized bool   `json:"normalized"`
			} `json:"uploaded_image"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 || response.Data.ID != "job_123" || response.Data.SearchStatus != "created" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Data.UploadedImage.ID != "img_123" || !response.Data.UploadedImage.Normalized {
		t.Fatalf("unexpected uploaded image: %#v", response.Data.UploadedImage)
	}
}

func TestImageSearchControllerCreateJobRequiresImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	controller := NewImageSearchController(&fakeImageSearchService{})
	router := gin.New()
	router.POST("/api/v1/image-search/jobs", controller.CreateJob)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/image-search/jobs", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestImageSearchControllerCreateJobRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeImageSearchService{}
	limiter := &fakeUploadLimiter{
		err: imagesearchdomain.NewUploadError(imagesearchdomain.ErrorRateLimited, "rate limited"),
	}
	controller := NewImageSearchController(service, ImageSearchControllerOptions{UploadLimiter: limiter})

	router := gin.New()
	router.POST("/api/v1/image-search/jobs", controller.CreateJob)

	body, contentType := multipartImageRequest(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/image-search/jobs", body)
	request.Header.Set("Content-Type", contentType)
	request.RemoteAddr = "203.0.113.7:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if limiter.clientID != "203.0.113.7" {
		t.Fatalf("expected limiter to receive client IP, got %q", limiter.clientID)
	}
	if service.createCalled {
		t.Fatal("expected service CreateJob not to be called")
	}

	var response struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Message != imagesearchdomain.ErrorRateLimited {
		t.Fatalf("expected rate_limited message, got %#v", response)
	}
}

type fakeImageSearchService struct {
	createInput  imagesearchdomain.UploadInput
	createResult imagesearchdomain.Job
	createCalled bool
}

func (f *fakeImageSearchService) CreateJob(_ context.Context, input imagesearchdomain.UploadInput) (imagesearchdomain.Job, error) {
	f.createCalled = true
	f.createInput = input
	return f.createResult, nil
}

func (f *fakeImageSearchService) GetJob(context.Context, string) (imagesearchdomain.Job, error) {
	return imagesearchdomain.Job{}, nil
}

func (f *fakeImageSearchService) GetResults(context.Context, string) (imagesearchdomain.Results, error) {
	return imagesearchdomain.Results{}, nil
}

func (f *fakeImageSearchService) RetryJob(context.Context, string) (imagesearchdomain.Job, error) {
	return imagesearchdomain.Job{}, nil
}

type fakeUploadLimiter struct {
	clientID string
	err      error
}

func (f *fakeUploadLimiter) AllowUpload(_ context.Context, clientID string) error {
	f.clientID = clientID
	return f.err
}

func multipartImageRequest(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "base.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(testControllerPNG(t)); err != nil {
		t.Fatalf("write png: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func testControllerPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}
