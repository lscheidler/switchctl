# switchctl

switchctl is a ssh wrapper for switch script, which makes it easier to switch multiple applications on multiple instances.

## Requirements

* [switch](https://github.com/lscheidler/switch) version >= 0.2.4 deployed and configured on target systems
* ssh access to target system with ssh key
* permissions to run switch on target system
* switchctl configuration in ~/.config/switchctl/config.yml or ./config.yml (see [config.yml.example](config.yml.example))

## Usage

```
switchctl -e <environment> -a <application>:<version> [-a <application>:<version>...]
```

### Example

```
switchctl -e staging -a app1:1.2.0 -a frontend1:2.1.0
```
