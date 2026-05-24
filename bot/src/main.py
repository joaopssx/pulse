import os
from dotenv import load_dotenv

from src.discord.bot import PulseBot
from src.server.routes import create_app
from src.core.bridge import Bridge


def main():
    load_dotenv()

    token = os.getenv("DISCORD_TOKEN", "")
    channel_id = int(os.getenv("DISCORD_CHANNEL_ID", "0"))
    secret = os.getenv("PULSE_SECRET", "")
    api_url = os.getenv("PULSE_API_URL", "http://localhost:8080")
    port = int(os.getenv("BOT_PORT", "5000"))

    bridge = Bridge(api_url=api_url)
    bot = PulseBot(channel_id=channel_id, bridge=bridge)
    app = create_app(bot=bot, secret=secret)

    _ = (token, port, app)


if __name__ == "__main__":
    main()
