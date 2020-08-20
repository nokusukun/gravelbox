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
	"binaries": [
		{
			"name": "exec.sh", 
			"data": "....", "resolve": true
		},
		{
			"name": "test.cs", "data": "..."
		}
	],
	"command": ["sh", "{path}/exec.sh"],
	"atom": "atom-mono",
	"timeout": "20s"
}
```
* `binaries`: the array of files to send to the sandbox
    * `name`: name of the file
    * `data`: base64 encoded contents
    * `resolve`: change replace all instances of `{path}` in the data
* `name`: binary destination filename
* `command`: command to run inside of the atom
* `atom`: name of the atom
* `timeout`: timeout...?