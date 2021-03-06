# go-api

Quick API prototyping with no ass pain (PostgreSQL + Go).

## Installation

```bash
$ go get github.com/arfeo/go-api
```

## Usage

### Preamble

Assume that you have some tables in your `PostgreSQL` database, e.g.:

* `users`

```
 id | username |       full_name        |      email      | is_activated | is_deleted |         registered_at         |          updated_at           
----+----------+------------------------+-----------------+--------------+------------+-------------------------------+-------------------------------
  1 | arfeo    | Arfeo                  | mail@mail.net   | t            | f          | 2019-03-17 20:24:36.489189+03 | 2019-03-24 13:20:35.792225+03
```

* `classes`

```
 id |    name    
----+------------
  1 | Private
  2 | Technician
  3 | Shooter
  4 | Officer
  5 | General
  6 | King
```

You also have functions that implement some actions on those tables and return data in JSON format:

```sql
CREATE FUNCTION users_change_full_name(_user_id integer, _full_name varchar) RETURNS json AS
  $$
  DECLARE
    result      users;
  BEGIN
    IF (_full_name = '') THEN
      RAISE EXCEPTION 'Full name can not be empty';
    END IF;

    UPDATE users SET full_name = _full_name, updated_at = current_timestamp WHERE id = _user_id;

    SELECT * INTO result FROM users WHERE id = _user_id;

    RETURN (
      SELECT json_build_object(
        'id', result.id,
        'username', result.username,
        'full_name', result.full_name,
        'email', result.email,
        'is_activated', result.is_activated,
        'is_deleted', result.is_deleted,
        'registered_at', result.registered_at,
        'updated_at', result.updated_at
      )
    );
  END;
  $$
LANGUAGE 'plpgsql';
```

and

```sql
CREATE FUNCTION classes_get_list() RETURNS json AS
  $$
  BEGIN
    RETURN (SELECT array_to_json(array_agg(row_to_json(s))) FROM classes s);
  END;
  $$
LANGUAGE 'plpgsql';
```

### Server side scripts

Create a directory of your project.

Create `config.json` file in the project's directory. It should have the next structure:

```json
{
	"db": {
		"host": "",
		"port": "",
		"user": "",
		"password": "",
		"database": "",
		"sslmode": ""
	},
	"tcp": {
		"host": "",
		"port": ""
	}
}
```

Specify the actual database connection settings (`db`) and TCP settings (`tcp`). Say the last ones are `"host": ""` and `"port": "8100"`.

Create `main.go` file in the project's directory. Put the next lines to it:

```go
package main

import (
  "github.com/arfeo/go-api"
)

func main() {
  api.Init("config.json", []api.Endpoint{
    {
      Entity: "classes",
      EntityMethod: "list",
      RequestMethod: "get",
      Query: "select classes_get_list()",
    },
    {
      Entity: "users",
      EntityMethod: "change_full_name",
      RequestMethod: "post",
      Params: []string{
        "user_id",
        "full_name",
      },
      Query: "select users_change_full_name($1, $2)",
    },
  })
}
```

Run the program by executing the following in the project's directory:

```bash
$ go run main.go
```

That's all. As easy as pie!

You can now test your API service:

```bash
$ curl -X GET localhost:8100/classes/list
[{"id":1,"name":"Private"},{"id":2,"name":"Technician"},{"id":3,"name":"Shooter"},{"id":4,"name":"Officer"},{"id":5,"name":"General"},{"id":6,"name":"King"}]
```

```bash
$ curl -X POST localhost:8100/users/change_full_name/ -d '{"user_id":1,"full_name":"Leonid Belikov"}'
{"id" : 1, "username" : "arfeo", "full_name" : "Leonid Belikov", "email" : "mail@mail.net", "is_activated" : true, "is_deleted" : false, "registered_at" : "2019-03-17T20:24:36.489189+03:00", "updated_at" : "2019-03-24T16:55:01.418285+03:00"}
```

## TODO

- [ ] Add authorization support
