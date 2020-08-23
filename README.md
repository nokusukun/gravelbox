# gravelbox

### Setup
* Install Docker
* `git pull github.com/nokusutwo/gravelbox`
* `docker build atom\`
* `go run .`

### Usage
* GET `/api/version` to get the docker version.
* GET `/api/atoms/create/:name` to create a new atom.
    * *The first atom usually takes forever to create*
* GET `/api/atoms/delete/:name` to remove atom.
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
	"atom": "mono",
	"timeout": "20s",
    "network": false,
    "read_only": true
}
```
*Items with \* are required*.

* *`binaries`: the array of files to send to the sandbox
    * *`name`: name of the file
    * *`data`: file contents, (could be JSON, string or base64 string)
    * `resolve`: change replace all instances of `{path}` in the data
    * `decode_b64`: treat the data as a base64 string and decode before saving the binary
    * `decode_json`: treat the data as a JSON object and marshal before saving the binary.
* *`command`: command to run inside of the atom
* *`atom`: name of the atom
* *`timeout`: container timeout
* `network`: enable/disable network access (default: false)
* `read_only`: enable disable writing to the filesystem (write access is required especially in compiled programs) (default: false)

### Executor
Atoms now have a built in utility named `executor` this facilitates fine grained program execution within the atom.
Executor works by following a `.execute` file. The working path is also changed to where the `.execute` file lives.

Sample `.execute` file
```json
{
    "commands": [{
        "command": "mcs",
        "args": ["-out:test.exe", "test.cs"]
    },{
        "command": "mono",
        "args": ["test.exe"],
        "timeout": "100ms"
    }]
}
```

* *`commands`: an array of commands to execute
    * *`command`: the main executable
    * `args`: an array of arguments
    * `timeout`: a go `time.Duration` string

This example file builds `test.cs` and then runs it with a maximum execution time of `100ms`.

---

### Sample

This example uses the `executor` utility in order to build the C# source code first.

POST `http://localhost:12375/api/atoms/execute`
```json
{
	"binaries": [
		{
			"name": ".execute", 
			"data": {
				"commands": [{
					"command": "mcs",
					"args": ["-out:test.exe", "test.cs"]
				},{
					"command": "mono",
					"args": ["test.exe"],
					"timeout": "100ms"
				}]
			},
			"decode_json": true
		},
		{
			"name": "test.cs", "data": "dXNpbmcgU3lzdGVtOwoKcHVibGljIGNsYXNzIEhlbGxvV29ybGQKewogICAgcHVibGljIHN0YXRpYyB2b2lkIE1haW4oc3RyaW5nW10gYXJncykKICAgIHsKICAgICAgICBDb25zb2xlLldyaXRlTGluZSAoIkhlbGxvIE1vbm8gV29ybGQiKTsKICAgIH0KfQ==",
			"decode_b64": true
		}
	],
	"command": ["executor", "{path}/.execute"],
	"atom": "v2",
	"timeout": "20s"
}
```
Output
```json
{
  "data": {
    "output": "Hello Mono World",
    "runtime": "bt1b67d6t3iklc53046g"
  },
  "error": null
}
```

---

This example just executes the JS script.

POST `http://localhost:12375/api/atoms/execute`
```json
{
	"binaries": [
		{
			"name": "test.js", "data": "console.log('Hello universe')"
		}
	],
	"command": ["node", "{path}/test.js"],
	"atom": "v2",
	"timeout": "1s"
}
```
Output
```json
{
  "data": {
    "output": "Hello universe",
    "runtime": "bt1bbu56t3iklc53047g"
  },
  "error": null
}
```

