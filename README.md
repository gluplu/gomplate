# Gomplate



This is a fork of https://github.com/hairyhenderson/gomplate that adds support for azure metadata.

It will read metadata keys as it is presented by the azure metadata service

To see what metadata is available you can run this on an azure VM

```
curl "http://169.254.169.254/metadata/instance/?api-version=2017-08-01&format=json" -H 'Metadata: true'|jq


```

However, the meta uses the text format when calling the metadata service in order to align better with the GCP meta.
It also allows - similarily to the GCP metadata service- retrieving an individual tag.

Azure does not provide the external IP of an instance in the instance metadata ( see https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service?tabs=linux#sample-7-retrieve-public-ip-address ) and but in the loadbalancer metadata. This uses a different endpoint with a different formatting but the code accounts for that 

## Build
```
go mod tidy
go build -o gomplate cmd/gomplate/main.go

```

## Example usage

```

 echo '{{azure.Meta "network/interface/0/ipv4/ipAddress/0/privateIpAddress"}}' | ./gomplate
10.1.0.4
 echo '{{azure.Meta "compute/tags/role"}}' | ./gomplate
chat
#Get external IP address
echo '{{azure.Meta "loadbalancer/publicIpAddresses/0/frontendIpAddress"}}' | ./gomplate
20.55.51.220
```                              
