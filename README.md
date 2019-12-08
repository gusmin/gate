# Secure Gate - Gate

[![Build Status](https://travis-ci.com/gusmin/gate.svg?token=6WEq9jpFesV2iXzoQsy4&branch=master)](https://travis-ci.com/gusmin/gate)

The gate is an essential part of the project "Secure Gate".

It's the interface used between the user and the other agents.

This project is developped with [Golang](https://golang.org/)
(go minimum version is 1.12)

## Install from release

### Get sources

- From release tarbar  
  [Releases](https://github.com/gusmin/gate/releases)
- From repository

  ```shell
  $ export GOPATH="${HOME}/go"
  $ export PATH="${PATH}:${GOPATH}/bin"
  $ go get -u github.com/magefile/mage
  $ mage -v release:linux
  $ cd release/securegate-gate-{version}
  ````

### Configuration

Copy the `config.json.template` to `config.json`
File `config.json`:

- `backend_uri` -> `String`: uri of the graphql endpoint of the backend
- `ssh_user` -> `String`: user that will be used for the ssh connections

### Install

Run `./install` script

The script:

- create the `secure` user
- install the binary to `/usr/bin`
- set the default shell of `secure` to `/usr/bin/securegate-gate`

## Installation local

At first set your environment:

```bash
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin"
```

Then install mage:

```shell
$ go get -u github.com/magefile/mage
```

Install the gate by running this command in your terminal:

```$shell
$ mage -v install
```

If you want to run unit tests as well as some linters(golint, go vet) you can run this additional command:

```shell
$ mage -v check
```

## Usage

The gate will first ask you some credentials. These are the ones you used for signing in to Secure Gate:

```shell
$ gate
Email: xxxx
Password:
Authentication successful
securegate$
```

### Available ommands

```console
help                ##Help about any command
me                  ## Display informations about the current user
add                 ## Add something
list                ## List all available machines
connect             ## Open SSH connection to a machine
logout              ## Terminate the session
exit                ## Close the shell
```

#### - help

Provides informations about the available commands.
Type wether `help [command]` or `[command] --help` for full details.

```shell
securegate$ help
Secure Gate makes the connection between computers more secure than ever

Usage:
  [command]

Available Commands:
  add       Add object
  connect   Open SSH connection to a machine
  ...
  ...

Use "[command] --help" for more information about this command
```

#### - me

Display all informations about the current user:

- Email
- Firstname
- Lastname
- Job
- ...

```shell
securegate$ me
Email: test@random.com
Firstname: Jean
Lastname: Martin
Job: dev
```

#### - add [command]

Add something(would be better if you try to add something that we can add)

Available things:

- `machine [name] [ip] [port]`

  ```shell
  securegate$ add machine foobar 127.0.0.1 3000
  Machine foobar successfully added
  securegate$ list
  [foobar] => 127.0.0.1:300
  ```

#### - connect [machine]

Create a ssh connection to the given machine(must obviously be an available machine for the user)

```shell
securegate$ connect foobar
dummy@foobar-pc:~$
```

#### - list

List all the machines available for the current user

```shell
securegate$ list
[foo] => 127.0.0.1:3000
[bar] => 127.0.0.2:3000
```

#### - logout

Terminate the active session.

```shell
securegate$ logout
Email:
```

#### - exit

Close the shell with the given exit status(0 by default)

```shell
securegate$ exit 42
$ echo $?
42
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.
