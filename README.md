linode-ssh-config
=================

**Modifies your `~/.ssh/config` file, preserving your exising entries, and appending your Linode hosts.**

    # Instead of connecting like
    ssh ubuntu@my.linode.server.com -i ~/.ssh/linode-rsa
    	
    # Connect like this
    ssh MyLinodeLabel
	
See [Simplify Your Life With an SSH Config File](http://nerderati.com/2011/03/simplify-your-life-with-an-ssh-config-file/) for more information about ssh aliases and configuration options.

## Quickstart

 * [Download](https://github.com/awilliams/linode-ssh-config/releases) the `linode-ssh-config` binary.

 * Create a [Linode API key](https://manager.linode.com/profile/api_key_create). (Click on `my profile` -> `API Keys`)
 
 * Create your `linode-ssh-config` config file, editing your API key and other variables.

  `cp linode-ssh-config.ini.example ~/.linode-ssh-config.ini`

 * First test the output.

  `./linode-ssh-config`

 * Make a backup of your exisiting ssh config
  
  `cp ~/.ssh/config ~/.ssh/config.bak`

 * Update the `~/.ssh/config` file.

  `./linode-ssh-config --update`
  
## Usage

The update command can be run repeatedly, as it replaces any previously generated configuration with the update configuration. 

    linode-ssh-config --update
    
## Bash Autocomplete

To enable `ssh` autocompletion, add the following to your bash profile

    # Add tab completion for SSH hostnames based on ~/.ssh/config, ignoring wildcards
    [ -e "$HOME/.ssh/config" ] && complete -o "default" -o "nospace" -W "$(grep "^Host" ~/.ssh/config | grep -v "[?*]" | cut -d " " -f2 | tr ' ' '\n')" scp sftp ssh

## Pretty Print

See a nicely formatted list of your linodes.

    ./linode-ssh-config --pp

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
