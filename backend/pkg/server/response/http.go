package response

import (
	"fmt"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/version"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HttpError struct {
	message  string
	code     string
	httpCode int
}

func (h *HttpError) Code() string {
	return h.code
}

func (h *HttpError) HttpCode() int {
	return h.httpCode
}

func (h *HttpError) Msg() string {
	return h.message
}

func NewHttpError(httpCode int, code, message string) *HttpError {
	return &HttpError{httpCode: httpCode, message: message, code: code}
}

func (h *HttpError) Error() string {
	return fmt.Sprintf("%s: %s", h.code, h.message)
}

func Error(c *gin.Context, err *HttpError, original error) {
	body := gin.H{"status": "error", "code": err.Code()}

	if version.IsDevelopMode() {
		body["msg"] = err.Msg()
		if original != nil {
			body["error"] = original.Error()
		}
	}

	fields := logrus.Fields{
		"code":    err.HttpCode(),
		"message": err.Msg(),
	}
	logger.FromContext(c).WithFields(fields).WithError(original).Error("api error")

	c.AbortWithStatusJSON(err.HttpCode(), body)
}

func Success(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{"status": "success", "data": data})
}

//lint:ignore U1000 successResp
type successResp struct {
	Status string      `json:"status" example:"success"`
	Data   interface{} `json:"data" swaggertype:"object"`
} // @name SuccessResponse

//lint:ignore U1000 errorResp
type errorResp struct {
	Status string `json:"status" example:"error"`
	Code   string `json:"code" example:"Internal"`
	Msg    string `json:"msg,omitempty" example:"internal server error"`
	Error  string `json:"error,omitempty" example:"original server error message"`
} // @name ErrorResponse
