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

var dbHandle *sql.DB

/**
 * Handler function for the host's root
 */
func handler(response http.ResponseWriter, request *http.Request, endpoints []Endpoint) {
  var (
    result          string
    outputError     error
    responseCode    int
    values          url.Values
    body            io.ReadCloser
  )

  routeSplit := strings.Split(request.URL.Path, "/")

  // If the URL path doesn't contain at least two sections divided by `/` symbol (/entity/entityMethod),
  // we assume that the resource can not be found
  if len(routeSplit) < 3 {
    result, outputError, responseCode = errorNotFound()

    output(response, result, outputError, responseCode)

    return
  }

  entity := routeSplit[1]
  entityMethod := routeSplit[2]

  // If the request method is known,
  // get the query string parameters for GET,
  // the request body for DELETE, POST and PUT
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

  entityMethodFound := false

  // Search for the given entity and entity method
  // in the endpoints structure; if found -- process parameters
  // depending on the request method
  for _, v := range endpoints {
    if v.Entity == entity && v.EntityMethod == entityMethod {
      entityMethodFound = true

      switch strings.ToUpper(v.RequestMethod) {
      case "GET":
        result, outputError, responseCode = processQueryString(values, v.Params, v.Query)
        break
      case "DELETE", "POST", "PUT":
        result, outputError, responseCode = processRequestBody(body, v.Params, v.Query)
        break
      default:
        result, outputError, responseCode = errorNotImplemented()
        break
      }
    }
  }

  if !entityMethodFound {
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
func processQueryString(values url.Values, params []string, query string) (result string, err error, responseCode int) {
  var row *sql.Row

  if len(params) > 0 {
    payload := make([]interface{}, len(params))

    // Iterate over `params` (as they are given in the endpoints structure)
    // and check, whether they are passed in the GET query string;
    // store passed parameters in `payload` in case of success,
    // otherwise return with the Bad Request error
    for i, v := range params {
      isValue := len(values[v]) > 0 && len(values[v][0]) > 0

      if !isValue {
        return errorNoArguments()
      }

      payload[i] = values[v][0]
    }

    row = dbHandle.QueryRow(query, payload...)
  } else {
    row = dbHandle.QueryRow(query)
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
func processRequestBody(body io.ReadCloser, params []string, query string) (result string, err error, responseCode int) {
  var row *sql.Row

  if len(params) > 0 {
    values := make(map[string]interface{})
    payload := make([]interface{}, len(params))

    decodeRequestBody(body, &values)

    // Iterate over `params` (as they are given in the endpoints structure)
    // and check, whether they are passed in the request body;
    // store passed parameters in `payload` in case of success,
    // otherwise return with the Bad Request error
    for i, v := range params {
      isValue := values[v] != ""

      if !isValue {
        return errorNoArguments()
      }

      payload[i] = values[v]
    }

    row = dbHandle.QueryRow(query, payload...)
  } else {
    row = dbHandle.QueryRow(query)
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

    // Remove `pq:` prefix for the database generated errors (RAISE EXCEPTION)
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
func Init(configFile string, endpoints []Endpoint) {
  var (
    config          config
    dbError         error
  )

  // Read and parse `config.json` file
  if !file.Exists(configFile) {
    log.Fatalln("Fatal error: Cannot find configuration file")
  }

  if configData, ok := file.Read(configFile); ok {
    if jsonDecodeError := json.Unmarshal([]byte(configData), &config); jsonDecodeError != nil {
      log.Fatalln(jsonDecodeError)
    }
  } else {
    log.Fatalln("Fatal error: Cannot read configuration file")
  }

  // Connect to the database
  connStr := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
    config.Db.Host,
    config.Db.Port,
    config.Db.User,
    config.Db.Password,
    config.Db.Database,
    config.Db.SslMode,
  )

  if dbHandle, dbError = sql.Open("postgres", connStr); dbError != nil {
    panic(dbError)
  }

  defer func() {
    if dbError = dbHandle.Close(); dbError != nil {
      log.Fatalln(dbError)
    }
  }()

  http.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
    handler(response, request, endpoints)
  })

  // Listen on the TCP network address
  tcpAddress := fmt.Sprintf(
    "%s:%s",
    config.Tcp.Host,
    config.Tcp.Port,
  )

  if listenerError := http.ListenAndServe(tcpAddress, nil); listenerError != nil {
    log.Println(listenerError)
  }
}
