# gocmlclient Changelog

Lists the changes in the gocmlclient package.

## Version 0.0.4

- added additional locking to prevent races
- improved test coverage, esp. w/ cache usage

## Version 0.0.3

- Fixed node tag list update. To delete all tags from a node, an empty tag list
  must be serialized in the `PATCH` JSON.  This was prevented by having
  `omitempty` in the struct.  Fixed  
- Also moved the `ctest` cmd fro the terraform provider repo to the code base.

## Versions prior to 0.0.3

Nothing in particular to be noteworthy -- just huge chunks of initial code
refactoring.
