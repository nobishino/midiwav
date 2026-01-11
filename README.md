# Setup

```sh
export MIDIWAV_DISCORD_WEBHOOK=~~~
export MIDIWAV_DIR=~~~
```

## Testing

Run tests:
```sh
go test
```

Update golden files (testdata/*.wav) with new output:
```sh
go test -update
```