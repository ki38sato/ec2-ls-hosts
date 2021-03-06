ec2-ls-hosts: an alternative tool for list ec2 instances
====

`ls-hosts` is a simple cli-tool for describing ec2 instances.
This tool will simplify the operation to describe instances.
You can integrate this tool with unix tools (eg: awk, ssh, peco, and so on.)

Usage
----

```bash
$ ls-hosts
i-00000001 10.0.0.1 app01 running
i-00000002 10.0.0.2 app02 running
i-00000003 10.0.0.3 app03 running
```

CLI Options
----

### with options

`-filters`

ec2 filter  [(Docs)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-DescribeInstancesInput)

```
-filters=key1:value1,key2:value2...
```

`-tags`

tag filter

```
-tags=key1:value1,key2:value2...
```

`-fields`

support fields

- instance-id
- private-ip
- public-ip
- tag:*

```
-fields=c1,c2,c3,...
```

`-region`

AWS region

`-creds`

support credentials

- env
- shared
- ec2

`-noheader`

Not Display field headers, if set true.

### with config file

1. ~/.ls-hosts
1. /etc/ls-hosts.conf

```
[options]
creds    = shared
region   = ap-northeast-1
tags     = Role:app,Env:production
fields   = instance-id,tag:Name,public-ip,private-ip
noheader = true
```

Integration with zsh and peco
----

[![https://gyazo.com/ae45206ad8215934f5e0a897b91b3d2a](https://i.gyazo.com/ae45206ad8215934f5e0a897b91b3d2a.gif)](https://gyazo.com/ae45206ad8215934f5e0a897b91b3d2a)

- With this integration, you can ssh login with interactive host selector

Dependencies

- zsh
- [peco](https://github.com/peco/peco)

```~/.zshrc
function peco-ec2-ls-hosts () {
  BUFFER=$(
    /path/to/ls-hosts -fields instance-id,tag:Role,tag:Name | \
    peco --prompt "EC2 >" --query "$LBUFFER" | \
    awk '{printf "echo \"Login:%s"; ssh %s\n", $3,$2}'
  )
  CURSOR=$#BUFFER
  zle accept-line
  zle clear-screen
}
zle -N peco-ec2-ls-hosts
bindkey '^oo' peco-ec2-ls-hosts
```

Build
----

```bash
$ make build (-B)
```

Dependencies

- [aws-sdk-go](https://github.com/aws/aws-sdk-go)
- [go-ini](https://github.com/go-ini/ini)
- [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)

Contribution
----

- Fork (https://github.com/ReSTARTR/ec2-ls-hosts/fork)
- Create a feature branch
- Commit your changes
- Rebase your local changes against the master branch
- Run test suite with the make test command and confirm that it passes
- Create a new Pull Request

Author
----

[ReSTARTR](https://github.com/ReSTARTR)
