# LinkOrDie

A self-hosted link graveyard. Save links, and a background worker checks them
forever: healthy links stay alive, flaky ones get a warning, and links that keep
failing are declared dead and moved to the cemetery. It's built in Go with Gin,
Postgres, templ templates, and Google/password login.

## TODO

- [ ] User interface: pages, components, and user flows
- [ ] Handlers: routes, request/response shapes, and auth requirements
