# Homing pigeon

Deliver messages from an input interface to an output interface.

![](homing_pigeon.png)

### Overview

Homing pigeon will listen to incoming messages from a reader interface and bring them to the requested storage (writer) interface.
The storage interface will be able to acknowledge (or not) the messages, so that this information
can be brought back to the reader and handled accordingly.

Currently implemented interfaces:

#### Read interfaces

##### RabbitMQ
Reader interface can handle acks and nacks. Nacks will be automatically sent to dead letter channel withouth retrying

#### Write interfaces

##### Elasticsearch with bulk API
Failed messages will be nacked, and successful messages will be acked.
It supports a well defined JSON format, which of course reminds of elasticsearch Bulk API:

```json
{
  "meta": { "index" : { "_index" : "test", "_id" : "1" } },
  "data": { "field1" : "value1" }
}
```
More info can be found at [elasticsearch's official doc](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html)

### Example

We own `example.com`, and we decide to track multiple user events (user clicking on `example` button, user filling up our `example` form, etc)
We want to be able to graph those events quickly.
So we decide to deploy homing pigeon, linked to existing rabbitmq and elasticsearch clusters.
Once deployed, we can start sending messages to a well defined exchange (with a well defined format) from our website,
and automatically they will be persisted in elasticsearch. All we need todo now is deploy a kibana instance, and graph the data!

### Usage

Running the binary file will start up listen interface.

```bash
$ ./homing-pigeon
```

In order to start up correctly, it needs well defined environment variables:

| Name                                 | Value                                                              |
| ------------------------------------ | ------------------------------------------------------------------ |
| ELASTICSEARCH_URL                    | Elasticsearch url string                                           |
| ELASTICSEARCH_FLUSH_MAX_SIZE         | Elasticsearch flush to bulk API maximum size                       |
| ELASTICSEARCH_FLUSH_MAX_INTERVAL_MS  | Elasticsearch flush to bulk API max interval time, in milliseconds |
| RABBITMQ_URL                         | RabbitMQ url string                                                |
| RABBITMQ_DLX_NAME                    | RabbitMQ dead letters exchange name                                |
| RABBITMQ_DLX_QUEUE_NAME              | RabbitMQ dead letters exchange's queue name                        |
| RABBITMQ_EXCHANGE_NAME               | RabbitMQ messaging exchange name                                   |
| RABBITMQ_QUEUE_NAME                  | RabbitMQ messaging exchange's queue name                           |
| RABBITMQ_CONSUMER_NAME               | Name for RabbitMQ's consumer (optional, defaults to HOSTNAME)      |
| RABBITMQ_QOS_PREFETCH_COUNT          | RabbitMQ QoS prefetch count (defaults to 0)                        |
| MESSAGE_BUFFER_LENGTH                | Buffer length for internal golang channel used for messaging       |
| ACK_BUFFER_LENGTH                    | Buffer length for internal golang channel used for acks            |

### Development

To run docker build:
```bash
$ make docker-build
```

To run the application:
```bash
$ docker compose up -d
```

To run tests:
```bash
$ make test
```

### Roadmap

* Implement interface for transforming messages after reader and before writer
* Add possibility to define username and password outside URLs for adapters