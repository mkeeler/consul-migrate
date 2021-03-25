# consul-migrate
Migrate Consul data from one datacenter to another

## Building

`go build ./cmd/consul-migrate`

## Exporting Data

To export Consul data to a file called data.json run the following:

`consul-migrate export -output data.json`

## Importing Data

To import Consul data from a file called data.json run the following:

`consul-migrate import -input data.json`