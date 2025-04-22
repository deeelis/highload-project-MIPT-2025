from random import randint
import aiofiles
import base64
import re

MAX_IMAGE_SIZE = 10
MAX_IMAGE_SIZE = MAX_IMAGE_SIZE * 1000000


async def save_base64_image(base64_str):
    try:
        image_bytes = base64.b64decode(base64_str)

        if len(image_bytes) > MAX_IMAGE_SIZE:
            print("Image too large.")
            return False

        file_name = f"{randint(6969, 6999)}.png"
        async with aiofiles.open(file_name, mode='wb') as f:
            await f.write(image_bytes)

        return file_name

    except Exception as e:
        print(f"Failed to save image from base64: {e}")
        return False


"""
async def save_base64_image(base64_str):
    try:
        match = re.match(r'data:image/(?P<ext>[^;]+);base64,(?P<data>.+)', base64_str)
        if not match:
            print("Invalid base64 format.")
            return False

        ext = match.group('ext')
        base64_data = match.group('data')
        image_bytes = base64.b64decode(base64_data)

        if len(image_bytes) > MAX_IMAGE_SIZE:
            return False
        
        file_name = f"{randint(6969, 6999)}.{ext}"
        async with aiofiles.open(file_name, mode='wb') as f:
            await f.write(image_bytes)

        return file_name

    except Exception as e:
        print(f"Failed to save image from base64: {e}")
        return False
"""