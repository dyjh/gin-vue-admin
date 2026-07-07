# Front Swagger Docs

This directory is reserved for generated swag documentation files for the front API group.

Run from `server`:

```bash
swag init -g main.go -d . -o front/docs --instanceName front --tags FrontTest --parseGoList=false
```
