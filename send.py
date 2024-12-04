import pika
import sys
import time

# RabbitMQ configuration
RABBITMQ_HOST = 'localhost'
RABBITMQ_USER = 'guest'
RABBITMQ_PASSWORD = 'guest'
EXCHANGE_NAME = 'homing-pigeon'
ROUTING_KEY = ''  # Leave empty or set appropriately
MESSAGE_COUNT = 100000  # Number of messages to send
MESSAGE_BODY = '{"meta": { "index" : { "_index" : "test", "_id" : "1" } },"data": { "field1" : "value1" }}'

def send_messages():
    try:
        # Connect to RabbitMQ
        connection = pika.BlockingConnection(pika.ConnectionParameters(host=RABBITMQ_HOST, credentials=pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASSWORD)))
        channel = connection.channel()

        # Declare the exchange (optional if already declared)
        channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type='fanout', durable=True)

        print(f"Starting to send {MESSAGE_COUNT} messages to exchange: {EXCHANGE_NAME}")

        for i in range(1, MESSAGE_COUNT + 1):
            # Publish message
            channel.basic_publish(
                exchange=EXCHANGE_NAME,
                routing_key=ROUTING_KEY,
                body=MESSAGE_BODY,
                properties=pika.BasicProperties(delivery_mode=2),  # Make messages persistent
            )
            
            # Print progress every 10,000 messages
            if i % 10000 == 0:
                print(f"Sent {i} messages...")

        print(f"Successfully sent {MESSAGE_COUNT} messages.")

    except KeyboardInterrupt:
        print("Process interrupted.")
        sys.exit(0)

    except Exception as e:
        print(f"An error occurred: {e}")

    finally:
        # Ensure connection is closed
        if 'connection' in locals() and connection.is_open:
            connection.close()

if __name__ == '__main__':
    start_time = time.time()
    send_messages()
    elapsed_time = time.time() - start_time
    print(f"Time taken: {elapsed_time:.2f} seconds")
