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
			"data": "....", 
            "resolve": true,
            "decode_b64": true
		},
		{
			"name": "test.cs", "data": "..."
		}
	],
	"command": ["sh", "{path}/exec.sh"],
	"atom": "atom-mono",
	"timeout": "20s",
    "network": false,
    "read_only": true
}
```
* `binaries`: the array of files to send to the sandbox
    * `name`: name of the file
    * `data`: file contents
    * `resolve`: change replace all instances of `{path}` in the data
    * `decode_b64`: treat the data as a base64 string and decode before saving the binary
* `name`: binary destination filename
* `command`: command to run inside of the atom
* `atom`: name of the atom
* `timeout`: timeout...?
* `network`: enable/disable network access (default: false)
* `read_only`: enable disable writing to the filesystem (write access is required especially in compiled programs) (default: false)