# ubiquiti-config-generator
This application will dynamically generate and deploy configuration changes for Ubiquiti routers based on local configuration abstractions.
Its focus is on host-centric home networks, with VLAN network segmentation and strict firewalls blocking access to/from networks.
As such, complex _network_ configurations, such as shared subnets across interfaces, is not something it is currently set up for.

That said, it is intended to reduce human error to a minimum, with validations being run automatically to flag errors such as host addresses not being in the designated subnets, multiple hosts assigned the same IP, and more (see below for a complete list).
It will also automatically figure out firewall mappings based on your inputs, port forwards, NAT translations, and more.

NOTE: configurations are NOT applied comprehensively; only changes will be applied to minimize risk of catastrophically damaging your router.
That means you still can manually enter anything you would like, BUT it will not have the collision/safety detection this application aims to provide.

----

## Data Schema
This [diagram.net](https://app.diagrams.net/?src=about#G1Lw4wh8zmSl0JGgrkhEczMQtAOhgKbUKq) shows the data architecture in use.
First, there are a few auxiliary schemas to call out up front:
- Some various global settings can be set in the global.yaml file as you see fit - none are required.
- Port groups are mapped separately, since they are by definition shared amongst multiple hosts and cannot easily be dynamically created via the paradigm in use.
- External addresses is a list of any outside IP addresses, accessible from the world
    - These are used in case of hairpin NAT, where you need to redirect a request to an external domain back inside your network

The bulk of the configuration follows.
- A network is the parent node, which defines a subnet (and any other dhcp server options).
- Underneath that lies an interface, which maps to a physical interface, an address, and any number of virtual interfaces recursively.
- Firewalls belong to interfaces, and have directions, default actions, and how many numbers to automatically increment.
    - Firewall rules will be automatically created based on host configurations.
- A host belongs to a network (not an interface, since interfaces _also_ map to networks), and has many of the properties you would expect for the address/firewall mappings you would expect.
    - Hosts are more complex, so will be better-documented in the next section

## Hosts
The host is the principal item in this configuration.

You must:
- For dynamic firewall configuration, specify either an address group and/or IP address
- Provide numeric ports, or names of valid port groups

You can:
- Provide ports to forward (NAT)
- Provide ports to redirect via hairpin (back to local network instead of exiting the router through WAN, NAT)
- Specify addresses/ports to allow inbound requests for (firewall)
- Specify addresses/ports to allow outbound requests to (firewall)

## Automatic validations
The automatic checks for configuration consistency are as follows:
- Since file names have to be unique, you cannot have two of the same network, interface, firewall, etc
- Address and port group names must actually exist to be used
- Reduction of human error in typing, since only one instance of any given data point
    - Device name or address, port (group) definitions, etc
- The same address cannot be used by multiple hosts
- The same address range cannot be shared across networks
- Firewall rules cannot be merged - e.g. rule 80 and 90 can't be mixed together by accident
    - Unless you manually change it in your configuration
- enable/disable keywords can be checked before commit/save called

## Committing changes
When you commit your changes to the repo, the post-merge hook will execute using the details in `router_connection_config.yaml`.
The hook will create commands using your diff, and apply them.
Unless you have specified the `autosave` configuration flag, it will NOT save the changes to disk, allowing for a reboot should something unexpected occur.
Otherwise, you will run `save` yourself after verifying the changes worked as expected.

## Getting started
First, fork or clone this repository.
You will likely want to move this to a self-hosted or at the very least private repository as it will contain your router configuration and details about connecting to your router.
Then, remove the last line of the .gitignore file, as it will prevent your configuration from being committed.
Next, we will set up your router configurations.

---

*All following configuration will be done under the `router_config` directory*

1. Fill out the router config for `example_router_connection_config.yaml` and rename it to `router_connection_config.yaml`
2. Put any (if applicable) external addresses in `external_addresses.yaml`
3. Fill out any configurations you wish to set in the rest of the top-level files
4. For each network:
    1. Add a new folder with the desired network name
    2. Create a `interfaces` folder (if an interface is applied to the network)
    3. Create a `hosts` folder (if there will be hosts statically mapped to this network)
    4. Create a `config.yaml` file
        a. The contents will be key-value pairs that should map to Ubiquiti key names
        b. There is validation of those key names, but it may not be complete
    5. For each interface:
        1. Create a firewall folder
        2. Populate the firewall folder with a `config.yaml` file defining its properties
            a. There must be a firewall created in the respective interface/s for any host that has "connect" ports set, both source and destination interfaces


---

Finally, set up automatic validations and committing to the router.
You will need to set up Travis CI or some similar service to perform checks on your merge/pull requests.
This is usually pretty easy and there are many guides on how to do this, such as [this one](https://docs.travis-ci.com/user/tutorial/#to-get-started-with-travis-ci-using-github).
After configuration with Travis CI, validation will automatically work on the repo and your configuration files.
    - The commands that would be run to modify the router's configuration will be printed in the test output!
*After* the initial configuration has been merged, set up the Git hook to load your configuration.
    Ensure this is only done after merging the initial configuration, since you (probably) don't want to reload all of it, again.

Then, you should be all set!
