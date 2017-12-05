docker-swarm-token-manager
==========================

Manage swarm tokens, better?

Usage
-----

Start this on a swarm node, or in a container in the swarm:

```
$ AUTH_WORKER=SEKRETTOKEN dstm
```

Then use curl on a node joining the swarm:

```
$ docker swarm join --token $(curl -H "Authorization: SEKRETTOKEN" http://dstm:8080/v1/token/worker)
```

Configuration
-------------

- `AUTH_WORKER` secret preshared token for workers
- `AUTH_MANAGER` secret preshared token for managers

License
-------
Affero GPL v3
