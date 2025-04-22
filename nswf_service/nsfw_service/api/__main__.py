# model = predict.load_model('nsfw_detector/nsfw_model.h5')

# KAFKA_BOOTSTRAP_SERVERS = "kafka:9093"
# TOPIC_INPUT = "nsfw_requests"
# TOPIC_OUTPUT = "nsfw_results"

import asyncio
import json
import logging
import os
from omegaconf import OmegaConf
from aiokafka import AIOKafkaConsumer, AIOKafkaProducer
from api.functions import save_base64_image
from nsfw_detector import predict


logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

config = OmegaConf.load("config.yaml")
model = predict.load_model(config.MODEL_PATH)


async def connect_to_kafka():
    consumer = None
    producer = None

    while not consumer or not producer:
        try:
            logger.info("⏳ Attempting to connect to Kafka...")
            consumer = AIOKafkaConsumer(
                config.KAFKA.TOPIC_INPUT,
                bootstrap_servers=config.KAFKA.BOOTSTRAP_SERVERS,
                group_id=config.KAFKA.GROUP_ID,
            )
            await consumer.start()

            producer = AIOKafkaProducer(
                bootstrap_servers=config.KAFKA.BOOTSTRAP_SERVERS,
            )
            await producer.start()

            logger.info("✅ Kafka connected successfully!")

        except Exception as e:
            logger.error(f"Kafka connection failed: {e}. Retrying in 5 seconds...")
            await asyncio.sleep(5)

    return consumer, producer


async def process_image(payload):
    try:
        image_id = payload.get("id")
        user_id = payload.get("user_id")
        base64_data = payload.get("data")

        if not base64_data:
            return {"id": image_id, "user_id": user_id, "error": "IMAGE DATA EMPTY"}

        image_path = await save_base64_image(base64_data)
        if not image_path:
            return {
                "id": image_id,
                "user_id": user_id,
                "error": "IMAGE SIZE TOO LARGE OR INVALID BASE64 DATA"
            }

        results = predict.classify(model, image_path)
        os.remove(image_path)

        hentai = results['data']['hentai']
        sexy = results['data']['sexy']
        porn = results['data']['porn']
        drawings = results['data']['drawings']
        neutral = results['data']['neutral']

        if neutral >= 25:
            is_nsfw = False
        elif (sexy + porn + hentai) >= 70:
            is_nsfw = True
        elif drawings >= 40:
            is_nsfw = False
        else:
            is_nsfw = False

        return {
            "id": image_id,
            "user_id": user_id,
            "nsfw_scores": results['data'],
            "is_nsfw": is_nsfw
        }

    except Exception as e:
        logging.error(f"Exception during processing: {str(e)}")
        return {"error": f"Exception during processing: {str(e)}"}


async def consume_and_produce():
    consumer, producer = await connect_to_kafka()
    logging.info("NSFW detection service is running...")

    try:
        while True:
            async for msg in consumer:
                try:
                    payload = json.loads(msg.value.decode("utf-8"))
                    logging.info(f"Processing image ID: {payload.get('id')}")
                    result = await process_image(payload)
                    await producer.send_and_wait(
                        config.KAFKA.TOPIC_OUTPUT,
                        json.dumps(result).encode("utf-8")
                    )
                    logging.info(f"✔️ Image processed: {payload.get('id')}")
                except Exception as e:
                    logging.error(f"Error during message handling: {e}")
    finally:
        logging.info("Shutting down cleanly")
        await consumer.stop()
        await producer.stop()


if __name__ == "__main__":
    try:
        asyncio.run(consume_and_produce())
    except KeyboardInterrupt:
        logger.info("🔌 Service stopped manually.")
