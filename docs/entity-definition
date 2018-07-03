### What is an entity?
 
With V3, we wanted to offer support for remote entities. To understand what we mean, let’s start by defining what we mean by an entity. **Entity** is a general term and refers to is a specific thing we collect data about. We used this term generally because we want to monitor different services and will need to support different parts such as hosts, pods, load balancers, DBs, etc. In the previous SDK versions (v1 & v2) the entity was local and just one, the host. However, as the entity does not necessarily have to be a host, we have broadened our support to include remote entities.
 
### What do we mean by remote entity?

We have defined an entity is defined as anything we collect data about. Previously, the entity was the local machine. In the new version, the **local entity** is the the local host. This leaves us to define what we mean by **remote entities**.
 
So what do we mean by remote entities? A remote entity is an entity, as understood above, that 1) is not the local host where either the agent and integration live or 2) is not equivalent to a “host”. In this way, we offer support to monitor a whole host of things, forgive the bad pun, that we might not otherwise be able to, due to being able to get an agent on the host or the architecture itself requiring the concept of multiple “entities”.
 
For Example:
 
An engineer installs the infrastructure agent on host1 and configures the out-of-the-box mysql integration to monitor a MySQL server running on host2. Host1 would be an entity as far as the Infrastructure entity is concerned and the MySQL servers on Host2 would be a “remote entity”.
 
An engineer installs the infrastructure agent on host1 and configures the out-of-the-box kubernetes integration and collects metrics and inventory about the whole cluster, replica sets, pods and nodes.
