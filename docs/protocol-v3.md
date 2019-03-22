# Integration Protocol v3

This version is basically the same as **v2**, but it improves **remote entities uniqueness** by adding "local address replacement on entity names" at agent level.


## Agent support

This integration protocol is supported by New Relic infrastructure agent since version `1.2.25`.


## Name local address replacement

When several remote entities have their `name` based on an endpoint (either ip or hostname), and this name contains a 
[loopback addresses](https://en.wikipedia.org/wiki/Localhost#Name_resolution), we have 2 issues:

- This localhost value does not provide valuable info without more context.
- Name could collide with other service being named with a local address.

This happens when:

- Endpoints names are like `localhost:port`.
- Ports tend be the same for a given service. Ie: 3306 for Mysql.


### Enabling replacement for protocol v2

Agent enables loopback-address replacement on the entity `name` (and therefor `key`) **automatically for v3** 
integration protocol.

It's possible to enable this for **v2** via agent configuration flag `replace_v2_loopback_entity_names`.

> In this case all the integrations being run by the agent using **v2** will have their names replaced whenever they carry 
> a local address.

