# Github Project Migration to Giant Swarm Roadmap

This repository contains a Go tool that replicates the functionality of the original bash script. It migrates GitHub project issues to the Giant Swarm Roadmap board.

## Dependencies
- [gh CLI tool](https://cli.github.com/)
- [jq](https://stedolan.github.io/jq/) (used in original script, not required by this Go tool)

## Build
To build the project run:
```
go build -o project-migration
```

## Usage
```
./project-migration <arguments>
  -h  Help
  -p  Project Number (eg 301)
  -t  Type (eg 'team, sig, wg')
  -n  Name of Team, SIG or WG (eg Rocket)
  -a  Area (eg KaaS)
  -f  Function (eg 'Product Strategy')
```

## Example
### Migrate a team
```bash
./project-migration -p 301 -t team -n Rocket -a KaaS -f 'Product Strategy'
```

### Migrate a SIG
```bash
./project-migration -p 302 -t sig -n 'Docs'
```

### Migrate a WG
```bash
./project-migration -p 303 -t wg -n 'Smart Factory' -a KaaS -f 'Product Strategy'
