# Github project migration to Giant Swarm Roadmap
This repository contains a script to migrate Giant Swarm Github projects to the Giant Swarm Roadmap project.

# Dependencies
* gh cli tool (https://cli.github.com/)
* jq (https://stedolan.github.io/jq/)

# Usage
```
./project-migration <arguments>
  -h  Help
  -p  Project Number (eg 301)
  -t  Type (eg 'team, sig, wg')
  -n  Name of Team, SIG or WG (eg Rocket)
  -a  Area (eg KaaS)
  -f  Function (eg 'Product Strategy')
```

# Example
## Migrate a team
```bash
./project-migration -p 301 -t team -n Rocket -a KaaS -f 'Product Strategy'
```

## Migrate a SIG
```bash
./project-migration -p 302 -t sig -n 'Docs'
```

## Migration a WG
```bash
./project-migration -p 303 -t wg -n 'Smart Factory' -a KaaS -f 'Product Strategy'
```
