# Server Agent

## Disclaimer 🚨

The code of "Server-Agent" is fully open sourced and has __NO WARRANTY__ of its safety. The code is not fully tested so may contain critical vulnerable bugs, which will make your server dangerous or even __BEING HIJACKED__ by attackers. __Use it at your own risk! ☠️__

## Setup

1. Choose your agent build version, you can get some common build versions from ["Github Release"](https://github.com/arctome/server-agent/releases/).

2. The executable file, `server-agent` needs an config file using YAML, here's an example of the config.

3. Download the latest UI from `./pages`. Now, your folder of agent should be like this:

```
- server-agent
- conf.yml
- pages/
  - index.html
  - ...
```

## Usage

```bash
nohup ./server-agent > /dev/null 2>&1 &
# Or print log to file
nohup ./server-agent > server-agent.out 2>&1 &
```

if you want to check the agent is running or kill it, run command below:

```bash
# Command below will echo the PID of server-agent
ps -aux | grep server-agent | grep -v grep
# If you want to kill it
kill -9 {PID}
```

## Build

You can compile the executable file yourself, like other common go package.

```bash
git clone https://github.com/arctome/server-agent.git
# Get all package locally and...
go build
```

## Document

"server-agent" is a server-client isomorphism program, use the same agent with different config.

### Agent

The basic mode is running in "agent" mode, with the easist YAML config:

```yaml
basic:
  # monitor/agent
  mode: agent
  # the running port of agent
  port: 
  # valiate agent password
  pass: 
  # The shared salt between "server" & "agent"
  salt: 
  # Is the agent running behind TLS & SSL ? It's recommanded to run behind SSL for encryption.
  enable_ssl: true
  # TODO: info/warning/error/none
  log_level: none
```

### Monitor

"monitor" mode depends on dashboard pages. When "server-agent" is running in "monitor" node, additional handlers will be activate. The metrics will be fetched via WebSocket connections.

An example config is here:

```yaml
basic:
  # monitor/agent
  mode: monitor
  port: 8989
  pass: 
  salt: 
  enable_ssl: true
  # TODO: info/warning/error/none
  log_level: none

servers:
  - name: server name
    host: host address of server, use domain or ip address
    port: port of agent
    pass: password of the agent
    enable_ssl: false
    enable_ssh: true
    ssh_user: for example, root
    ssh_port: ssh port
    location: 2 code or full 
    # ws connection cannot be established from HTTPS page, config this to `true` to use monitor's proxy socket instead.
    use_proxy: true
```

### Additional Features

"server-agent" uses [gofiber](https://github.com/gofiber/fiber) as its core, so you can enable some features of `gofiber` in features section:

```yaml
feature:
  # Enable gofiber's prefork feature
  prefork: false
  # Decided by your `./pages`, if you use `react` or `vue`, please set this to `true`
  spa_ui: true
  # See [pull request #1155](https://github.com/gofiber/fiber/pull/1155), for ipv6 only server.
  use_ipv6: true
```
