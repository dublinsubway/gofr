package gofr

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"gofr.dev/pkg/errors"
	"gofr.dev/pkg/gofr/template"
	"gofr.dev/pkg/gofr/types"
	"gofr.dev/pkg/middleware"
)

// Handler responds to HTTP request
// It takes in a custom Context type which holds all the information related to the incoming HTTP request
// It puts out an interface and error as output depending on what needs to be responded
type Handler func(c *Context) (interface{}, error)

type prometheusLabel struct {
	labelType string
	path      string
	method    string
}

// ServeHTTP processes incoming HTTP requests. It extracts the request context, handles errors,
// determines appropriate responses based on the data type, and sends the response back to the client.
// The method dynamically handles various response formats, such as custom types, templates, and raw data.
func (h Handler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	c, _ := r.Context().Value(gofrContextkey).(*Context)

	data, err := h(c)

	route := mux.CurrentRoute(r)
	path, _ := route.GetPathTemplate()
	// remove the trailing slash
	path = strings.TrimSuffix(path, "/")

	var errorResp error

	if _, ok := err.(errors.EntityAlreadyExists); ok || err == nil {
		errorResp = err
	} else {
		isPartialResponse := data != nil // since err!=nil we can check if data is not nil
		errorResp = processErrors(err, path, r.Method, isPartialResponse, c)

		// set the error in the context, which can be fetched in the logging middleware
		ctx := context.WithValue(r.Context(), middleware.ErrorMessage, err.Error())
		*r = *r.Clone(ctx)
	}

	switch res := data.(type) {
	case types.Response:
		c.resp.Respond(&res, errorResp)
	case template.Template, template.File, *types.Response, types.RawWithOptions:
		c.resp.Respond(res, errorResp)
	case types.Raw:
		c.resp.Respond(res.Data, errorResp)
	default:
		res = &types.Response{Data: data}
		c.resp.Respond(res, errorResp)
	}
}

//nolint:gocyclo // cannot be simplified further without hurting readability
func processErrors(err error, path, method string, isPartialError bool, c *Context) errors.MultipleErrors {
	var errResp errors.Response

	errResp.Value, errResp.TimeZone = evaluateTimeAndTimeZone()
	errResp.Reason = err.Error()

	switch v := err.(type) {
	case errors.InvalidParam:
		errResp.StatusCode = http.StatusBadRequest
		errResp.Code = "Invalid Parameter"
	case errors.MissingParam:
		errResp.StatusCode = http.StatusBadRequest
		errResp.Code = "Missing Parameter"
	case errors.EntityNotFound:
		errResp.StatusCode = http.StatusNotFound
		errResp.Code = "Entity Not Found"
	case errors.FileNotFound:
		errResp.StatusCode = http.StatusNotFound
		errResp.Code = "File Not Found"
	case errors.MethodMissing:
		errResp.StatusCode = http.StatusMethodNotAllowed
		errResp.Code = "Method not allowed"
	case *errors.Response:
		if v.DateTime.Value == "" {
			v.DateTime = errResp.DateTime
		}
		// pushing error type to prometheus
		incrPrometheusCounter(isPartialError, c, &errResp, prometheusLabel{
			labelType: "Unknown Error",
			path:      path,
			method:    method,
		})

		errResp = *v
	case errors.MultipleErrors:
		var finalErr errors.MultipleErrors
		finalErr.StatusCode = v.StatusCode
		now := time.Now()
		timeZone, _ := now.Zone()

		for _, v := range v.Errors {
			resp := errors.Response{}
			resp.TimeZone = timeZone
			resp.Value = now.UTC().Format(time.RFC3339)

			errs := processErrors(v, path, method, isPartialError, c)

			finalErr.Errors = append(finalErr.Errors, errs.Errors...)
		}

		return finalErr
	case errors.DB:
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Code = "Internal Server Error"
		errResp.Reason = "DB Error"

		c.Logger.Errorf("DB error occurred %v", err)

		// pushing error type to prometheus
		incrPrometheusCounter(false, c, &errResp, prometheusLabel{
			labelType: "DB error",
			path:      path,
			method:    method,
		})
	case errors.Raw:
		return errors.MultipleErrors{StatusCode: v.StatusCode, Errors: []error{v}}
	default:
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Code = "Internal Server Error"
		// pushing error type to prometheus
		incrPrometheusCounter(isPartialError, c, &errResp, prometheusLabel{
			labelType: "DB error",
			path:      path,
			method:    method,
		})
	}

	return errors.MultipleErrors{StatusCode: errResp.StatusCode, Errors: []error{&errResp}}
}

func incrPrometheusCounter(isPartialError bool, c *Context, errResp *errors.Response, label prometheusLabel) {
	if (errResp.StatusCode == http.StatusInternalServerError || errResp.StatusCode == 0) && !isPartialError {
		c.Logger.Errorf("error occurred %v", errResp.Reason)
		middleware.ErrorTypesStats.With(prometheus.Labels{"type": label.labelType, "path": label.path, "method": label.method}).Inc()
	}
}

func evaluateTimeAndTimeZone() (formattedTime, timeZone string) {
	now := time.Now()
	formattedTime = now.UTC().Format(time.RFC3339)
	timeZone, _ = now.Zone()

	return formattedTime, timeZone
}
