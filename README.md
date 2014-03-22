linode-ssh-config
=================

Generates a `~/.ssh/config` file with your Linode servers, preserving existing configuration.

Uses the Linode API to retrieve your hosts and generates the correct ssh config.
Instead of connecting like

    ssh ubuntu@my.linode.server.com -i ~/.ssh/linode-rsa
	
you can connect like this

    ssh MyLinodeLabel
	
See [Simplify Your Life With an SSH Config File](http://nerderati.com/2011/03/simplify-your-life-with-an-ssh-config-file/) for more information about ssh aliases and configuration options.

## Usage

 * [Download](https://github.com/awilliams/linode-ssh-config/releases) the correct `linode-ssh-config` for your computer.

 * Create a [Linode API key](https://manager.linode.com/profile/api_key_create).
 
 * Create your `linode-ssh-config` config file, editing your API key and other variables.

  `cp linode-ssh-config.ini.example ~/.linode-ssh-config.ini`

 * The program only outputs to STDOUT, so we can first test the output.

  `./linode-ssh-config`

 * Make a backup of your exisiting ssh config
  
  `cp ~/.ssh/config ~/.ssh/config.bak`

 * Generate and overwrite your current config. Do **NOT** write directly to your current config file. `linode-ssh-config` attempts to keep your existing configuration and append to it, but writting directly to the current config file will remove any existing configuration.
 
  `./linode-ssh-config > config.tmp && mv config.tmp ~/.ssh/config`

## Example output
```
# Existing config
Host devbox
  User vagrant
  IdentityFile ~/.ssh/vagrant.key

Host pi
  User pi
  StrictHostKeyChecking no

##### START GENERATED LINODE-SSH-CONFIG #####

## Production

Host Web1
        # Production | Linode ID 123456 | 1024m Ram
        Hostname 123.45.67.89
        IdentityFile ~/.ssh/linode-rsa
        User ubuntu

Host DB1
        # Production | Linode ID 765432 | 2048m Ram
        Hostname 111.11.111.11
        IdentityFile ~/.ssh/linode-rsa
        User ubuntu

## Staging

Host TestWeb1
        # Staging | Linode ID 112233 | 1024m Ram
        Hostname 128.11.11.11
        IdentityFile ~/.ssh/linode-rsa
        User ubuntu

Host TestDB1
        # Staging | Linode ID 765431 | 2048m Ram
        Hostname 111.11.111.12
        IdentityFile ~/.ssh/linode-rsa
        User ubuntu

##### END GENERATED LINODE-SSH-CONFIG #####
```
