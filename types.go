package api

type config struct {
  Db                configDb      `json:"db"`
  Tcp               configTcp     `json:"tcp"`
}

type configDb struct {
  Host              string        `json:"host"`
  Port              string        `json:"port"`
  User              string        `json:"user"`
  Password          string        `json:"password"`
  Database          string        `json:"database"`
  SslMode           string        `json:"sslmode"`
}

type configTcp struct {
  Host              string        `json:"host"`
  Port              string        `json:"port"`
}

type Endpoint struct {
  Entity            string
  EntityMethod      string
  RequestMethod     string
  Params            []string
  Query             string
}
