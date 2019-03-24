package api

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "net/url"
  "strings"

  "github.com/arfeo/go-file"
  _ "github.com/lib/pq"
)

var (
  AppConfig Config
  DBHandle *sql.DB
)

/**
 * Handler function for the host's root
 */
func handler(response http.ResponseWriter, request *http.Request, endpoints []Endpoint) {
  var (
    result string
    outputError error
    responseCode int
    values url.Values
    body io.ReadCloser
  )

  routeSplit := strings.Split(request.URL.Path, "/")

  if len(routeSplit) < 3 {
    result, outputError, responseCode = errorNotFound()

    output(response, result, outputError, responseCode)

    return
  }

  entity := routeSplit[1]
  method := routeSplit[2]

  switch strings.ToUpper(request.Method) {
  case "GET":
    values = request.URL.Query()
    break
  case "DELETE", "POST", "PUT":
    body = request.Body
    break
  default:
    result, outputError, responseCode = errorNotImplemented()
    output(response, result, outputError, responseCode)
    return
  }

  methodFound := false

  for _, v := range endpoints {
    if v.Entity == entity && v.EntityMethod == method {
      methodFound = true

      switch strings.ToUpper(v.RequestMethod) {
      case "GET":
        result, outputError, responseCode = processValues(values, v.Params, v.Query)
        break
      case "DELETE", "POST", "PUT":
        result, outputError, responseCode = processBody(body, v.Params, v.Query)
        break
      default:
        result, outputError, responseCode = errorNotImplemented()
        break
      }
    }
  }

  if !methodFound {
    result, outputError, responseCode = errorNotFound()
  }

  if result != "" || outputError != nil {
    output(response, result, outputError, responseCode)
  }
}

/**
 * Function processes the GET query string parameters and checks them
 * against the params list in the Endpoint's `Params` field (if any);
 * it executes the given database query with the given parameters (if any)
 * and returns its output, error, and the corresponding response code
 * (200 in case of success, otherwise -- 400)
 */
func processValues(values url.Values, params []string, query string) (result string, err error, responseCode int) {
  var row *sql.Row

  if len(params) > 0 {
    payload := make([]interface{}, len(params))

    for i, v := range params {
      isValue := len(values[v]) > 0 && len(values[v][0]) > 0

      if !isValue {
        return errorNoArguments()
      }

      payload[i] = values[v][0]
    }

    row = DBHandle.QueryRow(query, payload...)
  } else {
    row = DBHandle.QueryRow(query)
  }

  return getQueryResult(row)
}

/**
 * Function processes the request body and checks them
 * against the params list in the Endpoint's `Params` field (if any);
 * it executes the given database query with the given parameters (if any)
 * and returns its output, error, and the corresponding response code
 * (200 in case of success, otherwise -- 400)
 */
func processBody(body io.ReadCloser, params []string, query string) (result string, err error, responseCode int) {
  var row *sql.Row

  if len(params) > 0 {
    values := make(map[string]string)
    payload := make([]interface{}, len(params))

    decodeRequestBody(body, &values)

    for i, v := range params {
      isValue := values[v] != ""

      if !isValue {
        return errorNoArguments()
      }

      payload[i] = values[v]
    }

    row = DBHandle.QueryRow(query, payload...)
  } else {
    row = DBHandle.QueryRow(query)
  }

  return getQueryResult(row)
}

/**
 * Function outputs the query execution result
 */
func output(response http.ResponseWriter, result string, err error, responseCode int) {
  if err == nil {
    if _, writerError := fmt.Fprintf(response, result); writerError != nil {
      log.Println(writerError)
    }
  } else {
    var errorMessage string

    errorRaw := err.Error()

    if strings.HasPrefix(errorRaw, "pq: ") && len(errorRaw) > 4 {
      errorMessage = errorRaw[4:]
    } else {
      errorMessage = errorRaw
    }

    response.WriteHeader(responseCode)

    if _, writerError := fmt.Fprintf(response, "{\"error\":\"" + errorMessage + "\"}"); writerError != nil {
      log.Println(writerError)
    }
  }
}

/**
 * Function parses the given config file, opens the database connection,
 * calls the handler function for the host's root, and starts listening
 * on the TCP network address for incoming requests
 */
func Init(configFileName string, endpoints []Endpoint) {
  var dbError error

  // Read and parse `config.json` file
  if !file.Exists(configFileName) {
    log.Fatalln("Fatal error: Cannot find configuration file")
  }

  if config, ok := file.Read(configFileName); ok {
    if err := json.Unmarshal([]byte(config), &AppConfig); err != nil {
      log.Fatalln(err)
    }
  } else {
    log.Fatalln("Fatal error: Cannot read configuration file")
  }

  // Connect to the database
  connStr := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
    AppConfig.DbHost,
    AppConfig.DbPort,
    AppConfig.DbUser,
    AppConfig.DbPassword,
    AppConfig.DbName,
    AppConfig.DbSslMode,
  )

  if DBHandle, dbError = sql.Open("postgres", connStr); dbError != nil {
    panic(dbError)
  }

  defer func() {
    if dbError = DBHandle.Close(); dbError != nil {
      log.Fatalln(dbError)
    }
  }()

  http.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
    handler(response, request, endpoints)
  })

  // Listen on the TCP network address
  tcpAddress := fmt.Sprintf(
    "%s:%s",
    AppConfig.TcpHost,
    AppConfig.TcpPort,
  )

  if listenerError := http.ListenAndServe(tcpAddress, nil); listenerError != nil {
    log.Println(listenerError)
  }
}
