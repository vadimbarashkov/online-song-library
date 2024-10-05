package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

type songSchema struct {
	ID         uuid.UUID        `json:"id"`
	GroupName  string           `json:"groupName"`
	Name       string           `json:"name"`
	SongDetail songDetailSchema `json:"songDetail"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type songDetailSchema struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type songWithVersesSchema struct {
	ID        uuid.UUID `json:"id"`
	GroupName string    `json:"groupName"`
	Name      string    `json:"name"`
	Verses    []string  `json:"verses"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type paginationSchema struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	Items  uint64 `json:"items"`
	Total  uint64 `json:"total"`
}

type addSongRequest struct {
	Group string `json:"group" validate:"required"`
	Song  string `json:"song" validate:"required"`
}

type updateSongRequest struct {
	GroupName   string `json:"groupName"`
	Name        string `json:"name"`
	ReleaseDate string `json:"releaseDate" validate:"omitempty,releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link" validate:"url,omitempty"`
}

type songsResponse struct {
	Songs      []songSchema     `json:"songs"`
	Pagination paginationSchema `json:"pagination"`
}

type songWithVersesResponse struct {
	Song       songWithVersesSchema `json:"song"`
	Pagination paginationSchema     `json:"pagination"`
}

func parsePagination(r *http.Request) entity.Pagination {
	getUintQueryParam := func(key string, defaultValue uint64) uint64 {
		param := r.URL.Query().Get(key)
		if param != "" {
			value, err := strconv.ParseUint(param, 10, 64)
			if err == nil {
				return value
			}
		}

		return defaultValue
	}

	pagination := entity.Pagination{
		Offset: getUintQueryParam("offset", entity.DefaultOffset),
		Limit:  getUintQueryParam("limit", entity.DefaultLimit),
	}

	return pagination
}

func parseSongFilters(r *http.Request) []entity.SongFilter {
	var filters []entity.SongFilter

	addStringFilter := func(paramValue string, field entity.SongFilterField) {
		if paramValue != "" {
			filters = append(filters, entity.SongFilter{
				Field: field,
				Value: paramValue,
			})
		}
	}

	addDateFilter := func(paramValue string, field entity.SongFilterField) {
		if paramValue != "" {
			value, err := time.Parse("02.01.2006", paramValue)
			if err == nil {
				filters = append(filters, entity.SongFilter{
					Field: field,
					Value: value,
				})
			}
		}
	}

	query := r.URL.Query()

	addStringFilter(query.Get("groupName"), entity.SongGroupNameFilterField)
	addStringFilter(query.Get("name"), entity.SongNameFilterField)
	addDateFilter(query.Get("releaseYear"), entity.SongReleaseYearFilterField)
	addDateFilter(query.Get("releaseDate"), entity.SongReleaseDateFilterField)
	addDateFilter(query.Get("releaseDateAfter"), entity.SongReleaseDateAfterFilterField)
	addDateFilter(query.Get("releaseDateBefore"), entity.SongReleaseDateBeforeFilterField)
	addStringFilter(query.Get("text"), entity.SongTextFilterField)

	return filters
}

const statusError = "error"

type errorResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

var (
	emptyRequestBodyResp = errorResponse{
		Status:  statusError,
		Message: "empty request body",
	}

	invalidRequestBodyResp = errorResponse{
		Status:  statusError,
		Message: "invalid request body",
	}

	invalidSongIDParamResp = errorResponse{
		Status:  statusError,
		Message: "invalid song id param",
	}

	songNotFoundErrResp = errorResponse{
		Status:  statusError,
		Message: "song not found",
	}

	serverErrResp = errorResponse{
		Status:  statusError,
		Message: "server error occurred",
	}
)

func messageForValidateTag(tag string) string {
	switch tag {
	case "required":
		return "required field"
	case "releaseDate":
		return "invalid format, must be like '02.01.2006'"
	case "url":
		return "invalid url"
	default:
		return "invalid value"
	}
}

func getValidationErrorDetails(err error) []string {
	var details []string

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			msg := messageForValidateTag(e.Tag())

			details = append(details, fmt.Sprintf("%s: %s", field, msg))
		}
	}

	return details
}

func validationError(err error) errorResponse {
	return errorResponse{
		Status:  statusError,
		Message: "validation error",
		Details: getValidationErrorDetails(err),
	}
}
