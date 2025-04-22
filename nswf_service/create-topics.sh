#!/bin/bash

# Wait for Kafka to fully start
echo "⏳ Waiting for Kafka to be ready..."

# Create topic 1
kafka-topics --create \
  --bootstrap-server kafka:9092 \
  --replication-factor 1 \
  --partitions 1 \
  --topic content.images

# Create topic 2
kafka-topics --create \
  --bootstrap-server kafka:9092 \
  --replication-factor 1 \
  --partitions 1 \
  --topic content.images.result
  
echo "✅ Topics created: content.images, content.images.result"
