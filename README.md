# Webhook Dispatcher

Webhook Dispatcher is a simple PoC for now.

Its objectives are to stand between sources and sinks, routing incoming requests to the sinks that have subscribed to
then.

A source is a service where one can configure webhooks.
A sink is a service that receives http requests.

## Why

This project was born from two issues found when working in projects that heavily interact with third-party services
that allow one to register webhooks (e.g. jira, gitlab):

- registering and removing webhooks in services tends to be a manual process, locked behind high profiles with
  admin/almost admin-like permissions. It tends to be a time consuming task that drags multiple people into the process
  while adding almost nothing;
- if there are multiple slow sinks consuming the same webhooks from the same source, that source's performance will be
  degraded.

This project provides a solution to both issues by becoming the single entity responsible for managing webhooks.

- registration in sources only has to happen once per service, after that, sink subscriptions can be created and removed
  with simple http requests;
- the performance of core services sending webhooks is no longer affected, since the only consumer -
  webhooks-dispatcher - returns a 200 immediately and processes the request outside the request's routine.

## Important Notes

For now, this is nothing but a PoC.
And is not intended for real use.
