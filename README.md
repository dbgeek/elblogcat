# ELBLOGCAT

`elblogcat` is tool to list accesslogs for aws loadbalancers and cat them to see the content of them without need to download them local and cat them manual.

## Installation

### Manual installation - compile it

- [Install go](https://golang.org/doc/install)
- Set up `GOPATH` and add `$GOPATH/bin` to your `PATH`
- Run `go get -u github.com/dbgeek/elblogcat`

### Download github release 

```sh
curl https://github.com/dbgeek/elblogcat/releases/download/<release>/elblogcat_<release>_<os>_<arch>.tar.gz --out elblogcat_0.0.1-rc2_darwin_amd64.tar.gz
```

## Usage

### list all accesslog with prefix

```sh
elblogcat list --aws-account-id 1234567890 --s3-prefix-bucket lb-bucket --s3-prefix team-xxx
```

### list all accesslog with without prefix

```sh
elblogcat list --aws-account-id 1234567890 --s3-prefix-bucket lb-bucket
```

### cat accesslog for one load-balancer between a timerange

```sh
elblogcat cat --load-balancer-id load-balancer-id --start-time "2019-03-03 11:00:00" --end-time "2019-03-03 12:00:00"
```