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

Possible configuration options.

| Attribute  | Description                                  | Default | Required |
| ---------- | -------------------------------------------- | ------- | -------- |
| RootDir    | The project directory                        | .       |          |
| Build      | The command used to build the project        |         | \_       |
| Run        | The command used to run the project          |         | \_       |
| SourceAddr | The HTTP listen address                      | :3000   |          |
| TargetAddr | The listen address of the target application | :3001   |          |

## Example

`recharge` contains a demo application in the `demo` directory which is ready to be used with the included `recharge.conf`. Simply start the `recharge` application in its project root to see it in action.
