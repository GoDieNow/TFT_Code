# Design and implementation in Go of a microservices-based accounting system for cloud service providers

This repo contains the reference code for the TFT.


## Structure overview

Following, there's a high level overview of the structure of this repo.

Your main interest probably will be within the collector/services folder or the deploy one since they contain the code of the services and a sample way to run the whole system.

```
.
├── collectors
│   ├── blockstorage
│   │   ├── run/
│   │   ├── Dockerfile
│   │   └── build.sh
│   ├── network
│   ├── objects
│   └── servers
├── services
│   ├── billing
│   │   ├── client/
│   │   ├── models/
│   │   ├── restapi/
│   │   ├── run/
│   │   ├── server/
│   │   ├── build.sh
│   │   ├── Dockerfile
│   │   └── swagger.yaml
│   ├── cdr
│   ├── creditsystem
│   ├── customerdb
│   ├── eventsengin
│   ├── planmanager
│   └── udr
├── deploy
│   ├── prometheus
│   ├── docker-compose.yml
│   ├── planManagerLoader.sh
│   └── *-config.toml
├── templates
│   ├── service
│   └── collector
└── This file!
```


## Building

Each collector and service have their own Dockerfile and a build.sh bash script for building the container.
To do so, get into the folder of the service you want to build and execute:

```
bash build.sh
```

## Runing

To run any of the services you first need to build it, please refer to the previous section.

Within the folder run at the root of each collector and service, in a terminal, execute:

```
docker-compose up -d
```

## Deployment of the full system
In the folder **deploy** you will find all you need to start the full system.

Be aware that parts of it have been commented out since they require external components to work as expected.

The same way as when running an individual piece of the system, within the **deploy** folder, in a terminal, simply execute:

```
docker-compose up -d
```

## Useful links

- Fast overview of golang: https://learnxinyminutes.com/docs/go/
- Swagger generation tool: https://github.com/Stratoscale/swagger
- Golang swagger documentation: https://goswagger.io/

## Acknowledgement and Notice
This repo is just my take of a previous work that can be found [here](https://github.com/Cyclops-Labs/cyclops-4-hpc).

Since I'm the same dev and this repo only reorders the code, expands on the building/deployment, adds some service templates, and ensures that everything can be built, unless requested otherwise, this repo can be considered as a part of it and will keep the same license.