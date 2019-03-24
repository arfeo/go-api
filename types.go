package api

type Config struct {
  DbHost            string        `json:"db_host"`
  DbPort            string        `json:"db_port"`
  DbUser            string        `json:"db_user"`
  DbPassword        string        `json:"db_password"`
  DbName            string        `json:"db_name"`
  DbSslMode         string        `json:"db_sslmode"`
  TcpHost           string        `json:"tcp_host"`
  TcpPort           string        `json:"tcp_port"`
}

type Endpoint struct {
  Entity            string
  EntityMethod      string
  RequestMethod     string
  Params            []string
  Query             string
}
