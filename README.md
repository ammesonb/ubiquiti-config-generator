# ubiquiti-config-generator
<p align="center">
  <a href="https://travis-ci.com/ammesonb/ubiquiti-config-generator"><img alt="Build Status" src="https://travis-ci.com/ammesonb/ubiquiti-config-generator.svg?branch=main"></a>
  <a href="https://codecov.io/gh/ammesonb/ubiquiti-config-generator">
    <img src="https://codecov.io/gh/ammesonb/ubiquiti-config-generator/branch/main/graph/badge.svg" />
  </a>
  <a href="https://pyup.io/repos/github/ammesonb/ubiquiti-config-generator/"><img src="https://pyup.io/repos/github/ammesonb/ubiquiti-config-generator/shield.svg" alt="Updates" /></a>
  <a href="https://github.com/psf/black"><img alt="Code style: black" src="https://img.shields.io/badge/code%20style-black-000000.svg"></a>
  <a href="https://github.com/ammesonb/ubiquiti-config-generator/blob/trunk/LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-purple.svg"></a>
</p>

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
**It may be helpful to reference the sample router config included on this repo, which is also used for automated testing.**

First, there are a few auxiliary schemas to call out up front:
- Some various global settings can be set in the global.yaml file as you see fit - none are required.
    - Use `/` to denote arbitrarily-nested configuration paths in the yaml keys
- Port groups are mapped separately, since they are by definition shared amongst multiple hosts and cannot easily be dynamically created via the paradigm in use.
- External addresses is a list of any outside IP addresses, accessible from the world
    - These are used in case of hairpin NAT, where you need to redirect a request to an external domain back inside your network
- NAT is defined globally, with any rules for masquerade etc. defined explicitly in the nat/#.yaml files

The bulk of the configuration follows.
- A network is the parent node, which defines a subnet and interface (and any other dhcp server options).
    - Currently only ONE subnet per network is supported, but it should be fairly trivial to extend the YAML to allow for subnets to be a list
    - Interface details are rolled up to be include in the network, including VIF
- Firewalls belong to the network, and have directions, default actions, and how of a gap to leave when creating rule numbers (e.g. 10, 20, 30, ...).
    - Firewall rules will be automatically created based on host configurations, but can be explicitly defined as well
    - Auto-incrementing numbering will skip any which are defined manually
- A host belongs to a network (not an interface, since interfaces _also_ map to networks), and has many of the properties you would expect for the address/firewall mappings you would expect.
    - Hosts are more complex, so will be better-documented in the next section

## Hosts
The host is the a principal item in this configuration.

You must:
- For dynamic firewall configuration, specify either an address group and/or IP address belonging to the host
- Provide numeric ports, or names of valid port groups

You can:
- Provide ports to forward (NAT)
- Provide ports to redirect via hairpin (back to local network instead of exiting the router through WAN, NAT)
- Specify addresses/ports to allow inbound requests to/for (firewall)
    - These should be lists of address/port combinations

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
### Configuring this repository
First, fork or clone this repository.
You will likely want to move this to a self-hosted or at the very least private repository as it will contain references to your router configuration and details about connecting to your router.
Fully fill out the deploy configuration in `deploy.yaml`, as it contains necessary details for connecting to your router and deploy configurations, such as auto-restart times if you do not confirm changes!

Next, we will set up your router configurations.

### Router configuration
As mentioned above, see `sample_router_config` for a working example.
These files MUST be stored in a separate repository, which can be cloned independently of this codebase.

The file structure is as follows:
1. Put any (if applicable) external addresses in `external_addresses.yaml`
2. Fill out any configurations you wish to set in the rest of the top-level files
    a. Port groups, global settings, NAT
3. For each network:
    1. Add a new folder under `networks` with the desired network name
    2. Create a `config.yaml` file
        a. The contents will be key-value pairs that should map to Ubiquiti key names
        b. There is validation of those key names, but it may not be complete
    3. Optionally, add a `firewalls` folder for each of the `in`/`out`/`local`, as desired
        a. For each, add a folder with the firewall's name, and a `config.yaml` file under it
        b. If any omitted, placeholders will be created with a generic default of `accept`, since required for hosts to (potentially) add firewall rules to it
    4. Create a `hosts` folder (if there will be hosts statically mapped to this network)

### Setting up GitHub integrations
#### The GitHub App
1. Go to the [new app page](https://github.com/settings/apps/new).
2. Add a title and/or description that makes sense to you
3. Ensure a webhook URL is present, and active is ticked
    a. Provide a secret (recommended)
    b. TODO: which files need to be present at the webhook URL, at what addresses?
4. Set **repository** permissions to the following (TODO: is this exhaustive?)
    a. Checks: read/write
        i. To ensure that configuration is valid, automatically, for PRs
    b. Contents: read
        i. In order to pull configurations from GitHub, and deploy them
    c. Deployments: read/write
        i. To report status of deploying configurations
    d. Pull requests: read/write
        i. To add details to PRs about commands that would run, configuration state, etc.
    e. Commit statuses: read/write
        i. To report state of a given commit on the primary branch
5. Set event subscriptions to the following (TODO: is this exhaustive?)
    a. Check run
        i. To perform checks on the configuration
    b. Check suite
        i. To schedule checks on the configuration (TODO: this should schedule the check runs?)
    c. Deployment (TODO: and/or status?)
        i. To interact with the deployment status
    d. Pull request (TODO: is this needed?)
        i. MAYBE for checks or something? Push may be the better thing here
    e. Push
        i. To know when to schedule check runs or deployments
6. Ensure "only this account" is ticked for where the app can be installed
7. Create app
