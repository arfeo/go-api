package api

import (
  "database/sql"
  "encoding/json"
  "io"
  "log"
  "net/http"
)

/**
 * Function scans columns from the given row and return the result,
 * error (if any) and the corresponding response code
 * (200 in case of success, otherwise -- 400)
 */
func getQueryResult(row *sql.Row) (result string, err error, responseCode int) {
  if err = row.Scan(&result); err == nil {
    responseCode = http.StatusOK
  } else {
    responseCode = http.StatusBadRequest
  }

  return
}

/**
 * Function reads JSON-encoded values from the given request body
 * and stores them in value pointed to by target
 */
func decodeRequestBody(body io.ReadCloser, target interface{}) {
  data := json.NewDecoder(body)
  data.DisallowUnknownFields()

  if err := data.Decode(&target); err != nil {
    log.Println(err)
  }
}
