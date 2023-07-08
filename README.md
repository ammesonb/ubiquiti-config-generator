# ubiquiti-config-generator
This application will analyze, generate, validate, and deploy configuration changes for Ubiquiti routers based on local configurations.
The idea behind this is to minimize the possiblity of IP address conflicts, repetitive data entry, and provide visibility into the router configuration history via PRs and deployments.
Each change will generate a full config.boot file and auto-add it to your PR so you can see a full diff of changes to ensure that the output is what you expect, and will be preserved as `.config.boot` in the repository for reference.

While there are "smart" features provided via a network-firewall-host abstraction, this program can also read from the node.def VyOS configuration files, so anything that can be represented in that structure can also be used here, just without the abstractions.

----

## Data Schema
This [diagram.net](https://app.diagrams.net/?src=about#G1Lw4wh8zmSl0JGgrkhEczMQtAOhgKbUKq) shows the data architecture in use.
**It may be helpful to reference the sample router config included on this repo, which is also used for automated testing.**

Settings and runtime configurations are found in `config.yaml`.

1. To ensure compatibility with your device, you first must provide a templates directory that can be analyzed
2. Then, provide a list of *.boot files (can be modular) that will analyze general configurations
3. If you wish, you may then create the networks folder containing a list of DHCP services, which have any number of firewalls and hosts
    - A network is the parent node, which defines a subnet and interface (and any other dhcp server options).
        - Currently only ONE subnet per network is supported, but it should be fairly trivial to extend the YAML to allow for subnets to be a list
        - Also limits DHCP ranges to ONE per network, e.g. 200-254, NOT 100-150 + 200-250
        - Interface details are rolled up to be include in the network, including VIF
    - Firewalls belong to the network, and have directions, default actions, and how much of a gap to leave when creating rule numbers (e.g. 10, 20, 30, ...)
        - Firewall rules will be automatically created based on host configurations, but can be explicitly defined as well
        - Auto-incrementing numbering will skip any which are defined manually
    - A host belongs to a network (not an interface, since interfaces _also_ map to networks), and has many of the properties you would expect for the address/firewall mappings you would expect
        - Hosts are more complex, so will be better-documented in the next section

## Hosts
The host is the a principal item in this configuration.

You must:
- For dynamic firewall configuration, specify either an address group and/or IP address belonging to the host
- Provide numeric ports, or names of port groups

You can:
- Provide ports to forward (NAT)
- Provide ports to redirect via hairpin (back to local network instead of exiting the router through WAN, NAT)
- Specify addresses/ports to allow inbound requests to/from (firewall)
    - These should be lists of address/port combinations

## Validations
This program is capable of numerous validations, but broadly falls into two categories: ones from VyOS, and "smart" checks based off of the network/host abstractions to minimize chances of runtime problems like duplicate IP addresses.

The VyOS checks roughly consist of these validations:
- Inside numeric range
- Matches a RegEx pattern
- Value is in a specific options allowlist
- A command succeeds with the value (e.g. check-fw-name, etc)

The "smart" checks for configuration consistency are as follows:
- Since file names have to be unique, you cannot have two of the same network, interface, firewall, etc
- Address and port group names must actually exist to be used
- Reduction of human error in typing, since only one instance of any given data point
    - Device name or address, port (group) definitions, etc
- The same address cannot be used by multiple hosts
- The same address range cannot be shared across networks
- enable/disable keywords can be checked before commit/save called

## Modifying the Ubiquiti Configs
### Validating Changes
When you create a PR, on every change the post-push hook will execute using the details in your YAML config.
The hook will then load your boot configs and merge the abstractions into them, then validate using the template files' and abstraction's rules.
Results of the validations will be posted on the output and to a comment on the PR.
Note some are warnings, some will be blocking errors and must be resolved prior to merging.

You should also check the generated `.config.boot` file to ensure it matches your expectations.

### Deploying
After validations are successful, you can merge the PR.
This will generate a new deployment on the repo which you can use to track status and see errors, if any.

By default, the changes will be done with commit-confirm (unless you set the `save-after-commit: true` setting), allowing you to verify the Ubiquiti router is performing as you expect.
If the configuration fails to apply, a rollback will be performed automatically unless you disable the `rollback-on-failure` setting.

## Getting started
### Configuring this repository
First, fork or clone this repository.
You will likely want to move this to a self-hosted or at the very least private repository as it will contain references to your router configuration and details about connecting to your router.
Fully fill out the deploy configuration in `config.yaml`, as it contains necessary details for connecting to your router and deploy configurations, such as auto-restart times if you do not confirm changes!
Note that credentials must be set via environment variables.

Next, we will set up your router configurations.

### Router configuration
As mentioned above, see `sample_router_config` for a working example.
These files MUST be stored in a separate repository, which can be cloned independently of this codebase.

The file structure is as follows:
1. Add a `templates` directory containing your Ubiquiti product's settings to the repo, and provide that path in `templatesDir` in the `config.yaml` settings.
    - Typically this is something like `/opt/vyatta/share/vyatta-cfg/templates/`
2. Add any `*.boot` files wherever you like in the repo, but ensure the paths are added to `configFiles` in the settings.
3. For each network:
    1. Add a new folder under `networks` with the desired network name
    2. Create a `config.yaml` file
        - The contents will be key-value pairs that should map to Ubiquiti key names
        - Any mis-typed keys will simply be ignored by the parser
    3. Optionally, add a `firewalls` folder for each of the `IN`/`OUT`/`LOCAL`, as desired
        - For each, add a folder with the firewall's name, and a `config.yaml` file under it
        - If any omitted, placeholders will be created with a generic default of `accept`
        - Since network hosts may want to add firewall rules, these will be generated automatically if necessary
    4. Create a `hosts` folder and add any hosts present

### Setting up GitHub integrations
#### The GitHub App
1. Go to the [new app page](https://github.com/settings/apps/new).
2. Add a title and/or description that makes sense to you
3. Ensure a webhook URL is present, and active is ticked
    1. Provide a secret (recommended)
4. Set **repository** permissions to the following
    1. Checks: read/write
        1. To ensure that configuration is valid, automatically, for PRs
    2. Contents: read
        1. In order to pull configurations from GitHub, and deploy them
    3. Deployments: read/write
        1. To report status of deploying configurations
    4. Pull requests: read/write
        1. To add details to PRs about commands that would run, configuration state, etc.
    5. Commit statuses: read/write
        1. To report state of a given commit on the primary branch
5. Set event subscriptions to the following
    1. Check run
        1. To perform checks on the configuration
    2. Check suite
        1. To schedule checks on the configuration
    3. Deployment (TODO: and/or status?)
        1. To interact with the deployment status
    4. Pull request
        1. Needed to comment on/modify pull requests
    5. Push
        1. To know when to schedule check runs or deployments
6. Ensure "only this account" is ticked for where the app can be installed
7. Create the application
8. Optionally, follow [this guide](https://docs.github.com/en/free-pro-team@latest/github/administering-a-repository/enabling-required-status-checks) to require status checks.
    1. This _may_ require Pro version of GitHub (or a public config repo), either of which may be undesirable
    2. This is not required, but obviously would offer stronger protections against accidentally merging a broken configuration
9. Create a private configuration repo, if you have do not have one for your configuration already
10. Install the GitHub app to that repository

#### Configuring the app
1. Note the app ID, and add it in the `config.yaml` file
2. Go to the app's settings page, and generate a new private key.
  - This should be placed somewhere code in this repo can read it, and stored in the environment variable from the settings.
3. Set the folders to clone configs into as desired - THEY MUST BE DIFFERENT LOCATIONS
4. Note the webhook port defaults to 54321 - this should be configured as desired, with a reverse proxy forwarded to that port
