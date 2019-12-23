# Gate

[![GoDoc](https://godoc.org/github.com/gusmin/gate/pkg?status.svg)](https://godoc.org/github.com/gusmin/gate/pkg)
[![Build Status](https://travis-ci.com/atrahy/gate.svg?token=xuCuQMuPLxpwzP2jsQ7f&branch=master)](https://travis-ci.com/atrahy/gate)
[![Built with Mage](https://magefile.org/badge.svg)](https://magefile.org)

<img src="https://media.discordapp.net/attachments/433311912281767978/626863798610821130/logo_400dpi.png?width=764&height=884" width=100>

***

Gate is an essential part of the opensource project Secure Gate.
It is an interactive CLI used as an interface between users and [agents](https://github.com/atrahy/agent) in order to secure access permissions within your machines stack.

For more informations about Secure Gate check out our other repositories:

- [Backend](https://github.com/atrahy/backend)
- [Frontend](https://github.com/atrahy/frontend)
- [Agent](https://github.com/atrahy/agent)

***

## :cd: Installation

### From a release

1. Get a release of the project. There are various way to achieve this.

    - Get it from a [releases tarball](https://github.com/gusmin/gate/releases)
    - Package it directly from the repository:

        ```Shell
        $ go get -u github.com/magefile/mage
        $ mage release:linux
        ```

2. Set up the configuration.
  
    `cp` the `config.json.template` as `config.json` and edit it according to your needs.

    |           Setting          |                      Description                     |  Value |
    |:--------------------------:|:----------------------------------------------------:|:------:|
    |         backend_uri        |          URI of your running backend server          | string |
    |          ssh_user          |                 user used for SSH connection                 | string |
    | agent_authentication_token | Bearer token used for authentication on agent's side | string |
    |          language          |              Language of the application             | string |
    |        db_path       |             Path of your database            | string |

3. Install the Gate

    Go to the previously created release directory and **run the installation script**.

    ```Shell
    $ cd release/securegate-gate-{version}
    $ ./install
    ```

    This installation script will do **EVERYTHING** for you and set up the Gate

    - It creates a new user: `secure`
    - Installs the binary `securegate-gate` in `/usr/bin/`
    - Installs your custom configuration
    - Makes a special directory to store your logs in `/var/log/securegate/gate/`
    - Installs the available translations in case you need to change the langage later on.

        _Available languages are English, French and Korean_
    - Set the default shell of the user `secure` to `/usr/bin/securegate-gate`

### With Mage

1. If you don't have [Mage](https://magefile.org) already installed **install it**

    ```Shell
    $ go get -u github.com/magefile/mage
    ```

2. Install the Gate by running this command in your terminal

    ```
    $ mage install
    ```

## :milky_way: Welcome in the Gate

Congrats you finished to install the Gate ! Now let's get started ! :tada:

> :warning: _Note: Make sure an instance of our [Backend]([Backend](https://github.com/atrahy/backend)) is running somewhere and that your configuration file is tied to its URI._

In the first place you'll have to sign up with on your Secure Gate account.

_The credentials are the ones you used during your sign in on the [Frontend](https://github.com/atrahy/frontend)_.

### Commands

```console
help                ## Global help about commands
me                  ## Display informations related to the user
list                ## List all accessible nodes by the user
connect             ## Open SSH connection to a node
logout              ## Terminate the session current session
exit                ## Close the shell
```

#### :books: Help

Provides global informations about  commands in the Gate

> _Note: Althought every commands have their own help option._
>
> _e.g. `[command] --help`_

```
securegate$ help
Secure Gate makes the connection between computers more secure than ever

Usage:
  [command]

Available Commands:
  connect     Open SSH connection to a node
  machine     List all accessible nodes by the user
  ...
  ...

Use "[command] --help" for more information about this command
```

#### :ok_woman: Me

Display informations related to the current user

```
securegate$ me
+--------+-----------+----------+------------+
| EMAIL  | FIRSTNAME | LASTNAME |     JOB    |
+--------+-----------+----------+------------+
| Secure | Gate      | Is       | Wonderfull |
+--------+-----------+----------+------------+
You.
```

#### :scroll: List

List all the accessible machines by the current user

```
securegate$ list
+------------------+----------+-------------------------------------------------------+-------+
|        ID        |   NAME   |                          IP                           | PORT  |
+------------------+----------+-------------------------------------------------------+-------+
| 09gtWjWi9SOVSGb1 | NASA     | localhost                                             |  3002 |
| VexuCBYu0JOHzy84 | AREA-51  | localhost                                             | 62774 |
| kKQSHWF2cl1pjZdp | nowhere  | localhost                                             |  3002 |
+------------------+----------+-------------------------------------------------------+-------+
Available nodes.
```

#### :electric_plug: Connect

Open SSH connection toward the machine given as argument

```Shell
securegate$ connect nowhere
dummy@nowhere-pc:~$
```

#### :walking: Logout

Terminate the current session

```
securegate$ logout
See you soon bro!
Email:
```

#### :running: Exit

Close the shell with the given exit status (0 by default)

```shell
securegate$ exit 42
$ echo $?
42
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.
