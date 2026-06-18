package controller

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	imagesearchdomain "github.com/ww1489/WarSpark/internal/domain/imagesearch"
	"github.com/ww1489/WarSpark/internal/utils"
)

type ImageSearchManager interface {
	CreateJob(ctx context.Context, input imagesearchdomain.UploadInput) (imagesearchdomain.Job, error)
	GetJob(ctx context.Context, id string) (imagesearchdomain.Job, error)
	GetResults(ctx context.Context, id string) (imagesearchdomain.Results, error)
	RetryJob(ctx context.Context, id string) (imagesearchdomain.Job, error)
}

type ImageSearchUploadLimiter interface {
	AllowUpload(ctx context.Context, clientID string) error
}

type ImageSearchControllerOptions struct {
	UploadLimiter ImageSearchUploadLimiter
}

type ImageSearchController struct {
	service       ImageSearchManager
	uploadLimiter ImageSearchUploadLimiter
}

func NewImageSearchController(service ImageSearchManager, options ...ImageSearchControllerOptions) *ImageSearchController {
	var option ImageSearchControllerOptions
	if len(options) > 0 {
		option = options[0]
	}
	return &ImageSearchController{
		service:       service,
		uploadLimiter: option.UploadLimiter,
	}
}

// CreateJob creates an anonymous image search job from an uploaded screenshot.
//
// @Summary Create image search job
// @Tags image-search
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Base screenshot"
// @Param war_target_id formData string false "War target ID"
// @Param expected_th formData int false "Expected TH"
// @Success 200 {object} utils.Response
// @Failure 422 {object} utils.Response
// @Router /api/v1/image-search/jobs [post]
func (c *ImageSearchController) CreateJob(ctx *gin.Context) {
	clientID := clientIP(ctx.Request)
	if c.uploadLimiter != nil {
		if err := c.uploadLimiter.AllowUpload(ctx.Request.Context(), clientID); err != nil {
			failImageSearch(ctx, err)
			return
		}
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "image is required"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		utils.Fail(ctx, utils.NewError(utils.ErrInvalidField, "image unreadable"))
		return
	}
	defer file.Close()

	targetContext, err := imageSearchTargetContext(ctx)
	if err != nil {
		utils.Fail(ctx, err)
		return
	}

	job, err := c.service.CreateJob(ctx.Request.Context(), imagesearchdomain.UploadInput{
		FileName:      fileHeader.Filename,
		Size:          fileHeader.Size,
		Reader:        file,
		ClientIP:      clientID,
		TargetContext: targetContext,
	})
	if err != nil {
		failImageSearch(ctx, err)
		return
	}

	utils.OK(ctx, job)
}

// GetJob returns current image search job status.
//
// @Summary Get image search job
// @Tags image-search
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/image-search/jobs/{job_id} [get]
func (c *ImageSearchController) GetJob(ctx *gin.Context) {
	jobID := ctx.Param("job_id")
	if jobID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "job id is required"))
		return
	}

	job, err := c.service.GetJob(ctx.Request.Context(), jobID)
	if err != nil {
		failImageSearch(ctx, err)
		return
	}
	utils.OK(ctx, job)
}

// GetResults returns aggregated image search results.
//
// @Summary Get image search results
// @Tags image-search
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/image-search/jobs/{job_id}/results [get]
func (c *ImageSearchController) GetResults(ctx *gin.Context) {
	jobID := ctx.Param("job_id")
	if jobID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "job id is required"))
		return
	}

	results, err := c.service.GetResults(ctx.Request.Context(), jobID)
	if err != nil {
		failImageSearch(ctx, err)
		return
	}
	utils.OK(ctx, results)
}

// RetryJob requeues a failed or low-confidence image search job.
//
// @Summary Retry image search job
// @Tags image-search
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/v1/image-search/jobs/{job_id}/retry [post]
func (c *ImageSearchController) RetryJob(ctx *gin.Context) {
	jobID := ctx.Param("job_id")
	if jobID == "" {
		utils.Fail(ctx, utils.NewError(utils.ErrMissingField, "job id is required"))
		return
	}

	job, err := c.service.RetryJob(ctx.Request.Context(), jobID)
	if err != nil {
		failImageSearch(ctx, err)
		return
	}
	utils.OK(ctx, job)
}

func imageSearchTargetContext(ctx *gin.Context) (imagesearchdomain.TargetContext, error) {
	targetContext := imagesearchdomain.TargetContext{
		WarTargetID:   ctx.PostForm("war_target_id"),
		EnemyPosition: ctx.PostForm("enemy_position"),
		EnemyName:     ctx.PostForm("enemy_name"),
	}
	if value := ctx.PostForm("expected_th"); value != "" {
		expectedTH, err := strconv.Atoi(value)
		if err != nil || expectedTH <= 0 {
			return imagesearchdomain.TargetContext{}, utils.NewError(utils.ErrInvalidField, "invalid expected_th")
		}
		targetContext.ExpectedTH = &expectedTH
	}
	return targetContext, nil
}

func failImageSearch(ctx *gin.Context, err error) {
	var uploadErr imagesearchdomain.UploadError
	if errors.As(err, &uploadErr) {
		if uploadErr.Code == imagesearchdomain.ErrorRateLimited {
			utils.JSON(ctx, http.StatusTooManyRequests, int(utils.ErrRateLimited), uploadErr.Code, nil)
			return
		}
		utils.JSON(ctx, http.StatusUnprocessableEntity, int(utils.ErrInvalidField), uploadErr.Code, nil)
		return
	}
	if errors.Is(err, imagesearchdomain.ErrJobNotFound) {
		utils.Fail(ctx, utils.NewError(utils.ErrNotFound, "image search job not found"))
		return
	}
	utils.Fail(ctx, err)
}

func clientIP(request *http.Request) string {
	if request == nil {
		return ""
	}
	if forwarded := request.Header.Get("X-Forwarded-For"); forwarded != "" {
		host, _, err := net.SplitHostPort(forwarded)
		if err == nil {
			return host
		}
		return forwarded
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}
	return host
}
