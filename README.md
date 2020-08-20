# gravelbox

### Setup
* Install Docker Desktop or whatever
* `docker build atom\`
* `go run .`

### Usage
* GET `/api/version` to get the docker version.
* GET `/api/atoms/create/:name` to create a new atom.
    * *The first atom usually takes forever to create*
* GET `/api/atoms/list` to list current atoms.
* POST `/api/atoms/execute` to execute a thing.
```json
{
	"binary": "Y29uc29sZS5sb2coImhlbGxvIG5lcmQiKQ==",
	"name": "test.js",
	"command": ["node", "{path}/test.js"],
	"atom": "atom-runner",
	"timeout": "20s"
}
```
* `binary`: is the base64 encoded source file
* `name`: binary destination filename
* `command`: command to run inside of the atom
* `atom`: name of the atom
* `timeout`: timeout...?