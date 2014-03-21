linode-ssh-config
=================

Generate your `~/.ssh/config` file with all your Linode hosts.

Uses the Linode API to retrieve your hosts and generates the correct ssh config.
Instead of connecting like

    ssh bill@my.linode.server.com -i ~/.ssh/linode-rsa
	
you can connect like this

    ssh MyLinodeHost
	
See [Simplify Your Life With an SSH Config File](http://nerderati.com/2011/03/simplify-your-life-with-an-ssh-config-file/) for more information about ssh aliases and configuration options.

## Usage

 * Download the correct `linode-ssh-config` for your computer.
 * Find your [Linode API key](https://manager.linode.com/profile/api_key_create).
 * Create your `linode-ssh-config` config file, editing your API key and other variables.
 
    cp linode-ssh-config.ini.example ~/.linode-ssh-config.ini
 
 * Test the output

    ./linode-ssh-config

 * Make a backup of your exisiting ssh config
 
    cp ~/.ssh/config ~/.ssh/config.bak

 * Generate and overwrite your current config. Do **NOT** write directly to your current config file. `linode-ssh-config` attempts to keep your existing configuration and append to it, but writting directly to the current config file will remove any existing configuration.
 
 	./linode-ssh-config > config.tmp && mv config.tmp ~/.ssh/config