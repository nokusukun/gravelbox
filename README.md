# gravelbox

Gravelbox is a containerized execution platform with a REST API, it's main functions is 
to run unsafe source code (or anything, really) with granular controls such as timeouts and resource
access.

## Sections
* [**Setup**](#setup)
* [**Configuration**](#configuration)
* [**Usage**](#usage)
* [**Executor**](#executor)
* [**Examples**](#examples)

### Setup
* Install Docker
* `git pull github.com/nokusutwo/gravelbox`
* `docker build atom\`
* `go run .`

### Configuration
There's an included `gravel.ini` with rudimentary configuration options.
```ini
[docker]
# the docker command to use
command=docker
# Global docker execution timeout for all docker commands 
# (this includes when sending a /atom/execute) request
timeout=120s

[atom]
# Path to the atom folder
path=atom

[gravelbox]
# Mount point to where the files are saved prior to execution
mountdir=_mount
```

### Usage
See examples below.
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
	"command": ["sh", "exec.sh"],
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

### What's the deal with `{path}`?
Before the container is started, the binaries are saved in a unique folder specified in `gravelbox.mountdir` in `gravel.ini`.
The folder name is random so gravelbox automatically replaces the instances of `{path}` prior to execution in `$.command` as well in
the binaries using the `resolve` flag. 

`"command": ["sh", "{path}/exec.sh"]` is automatically resolved to `sh /mnt/12345ABCDE/exec.sh`.




### Executor
Atoms now have a built in utility named `executor` which facilitates fine grained program execution within the container.
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
* `no_parse`: Disables the attempt to parse each execution as an array. One output string per command. (default: `false`)
* `export_json`: Export the parsed output to JSON. (default: `false`)

This example file builds `test.cs` and then runs it with a maximum execution time of `100ms`.

---

### Examples

This example uses the `executor` utility in order to build the C# source code first.
The binary includes two files namely `.execute` which is a JSON file and `test.cs` which is a base64 encoded source code.

The atom execution is limited to run for 20s in which the rest endpoint returns with an error if it exceeds past that.

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
	"command": ["executor", ".execute"],
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
	"command": ["node", "test.js"],
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

