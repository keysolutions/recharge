# Recharge

A utility written in Go to provide reloading for Go HTTP applications.

## About

`recharge` watches a Go project for changes and rebuilds the target application when changes occur. `recharge` operates by providing a proxy through to th target application which coordinates the building and running of the target application while being transparent to the user.

## Installing

`go get -u github.com/keysolutions/recharge`

## Configuration

`recharge` uses a TOML configuation file to assign project-specific options. An example configuration file is found in this repository under the name `recharge.conf`.

    RootDir = "."
    Build = "go build -o ./demo/demo ./demo"
    Run = "./demo/demo"
    SourceAddr = ":3000"
    TargetAddr = ":3001"

Possible configuration options:

| Attribute  | Description                                  | Default | Required |
| ---------- | -------------------------------------------- | ------- | -------- |
| RootDir    | The project directory                        | .       |          |
| Build      | The command used to build the project        |         | \_       |
| Run        | The command used to run the project          |         | \_       |
| SourceAddr | The HTTP listen address                      | :3000   |          |
| TargetAddr | The listen address of the target application | :3001   |          |

If `bash` or `sh` is available on the system the build and run commands will be executed as `bash -c "$COMMAND"` to give greater flexibility in creating build and run command rules. When they are not available the command will fall back to being executed directly using Go's exec.Command functionality. Authors of `recharge.conf` files should be aware of the limitations of various systems that their projects may be run on.

## Example

`recharge` contains a demo application in the `demo` directory which is ready to be used with the included `recharge.conf`. Simply start the `recharge` application in its project root to see it in action.
