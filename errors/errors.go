package errors

import (
  "net/http"

  "github.com/pkg/errors"
)

/**
 * Function returns empty string, error, and error code 400 when not all or even all needed arguments
 * for the called method passed
 */
func NoArguments() (string, error, int) {
  return "", errors.New("Not all needed arguments passed"), http.StatusBadRequest
}

/**
 * Function returns empty string, error, and error code 404 when URL can not be parsed
 */
func NotFound() (string, error, int)  {
  return "", errors.New("Requested URL not found"), http.StatusNotFound
}

/**
 * Function returns empty string, error, and error code 501 when unknown method is called (neither GET, nor POST)
 */
func NotImplemented() (string, error, int)  {
  return "", errors.New("Method not implemented"), http.StatusNotImplemented
}
