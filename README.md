# go-dataloader

A clean, safe, user-friendly implementation of [GraphQL's Dataloader](https://github.com/graphql/dataloader), written with generics in go (go1.18beta1).

## Features

- written in generics with strong type
- nearly same interface with original nodejs version dataloader
- promise like thunk design, simply call `val, err := loader.Load(ctx, id).Get(ctx)`
- customizable cache, easily wrap lru or redis.
- customizable scheduler, can manual dispatch, or use time window (default)

## Requirement

- Only support go >= 1.18, currently in beta

## Getting Started

## TODO

- [ ] Docs
- [ ] Support hooks for observability
- [ ] Rewrite tests with mock (waiting for mockgen support 1.18)
