# What's next

Preá is pre-alpha. The core is in place: preás, components, configuration, the database layer, the CLI and these docs. These are the next steps, roughly in order. There are no dates; each lands when it is solid.

## Testing tools

Testing an app should be as easy as writing it:

- a test database for each test, so tests can run in parallel;
- replacing any component with a fake in a single line;
- helpers for requests and replies, without decoding JSON by hand;
- a generated test for every new preá.

## Security

A complete, safe-by-default security layer, configured from the settings:

- a ready-made authentication middleware;
- revocable sessions, refresh tokens and roles;
- rules for ownership: "only the owner of this list can change it".

## Persistence

A new [persistence layer](/persistence/) with more databases, a query language for your models (PPRQL), specifications and pagination.

## Validation

A validation layer of its own, with clear messages, that also works outside HTTP: in services, in configuration blocks and in background jobs.

## Follow along

- The [changelog](./changelog) lists what each version brought.
- The [repository](https://github.com/prea-framework/prea) is where the work happens.
