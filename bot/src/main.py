import os
import threading

from dotenv import load_dotenv
from flask import Flask

from src.discord.bot import PulseBot
from src.server.routes import pulse_routes


def _run_flask(app: Flask, port: int) -> None:
    app.run(host="0.0.0.0", port=port, use_reloader=False, debug=False)


def main() -> None:
    load_dotenv()

    token = os.getenv("DISCORD_TOKEN", "")
    channel_id = int(os.getenv("DISCORD_CHANNEL_ID", "0") or "0")
    port = int(os.getenv("BOT_PORT", "5000") or "5000")

    if not token:
        raise SystemExit("DISCORD_TOKEN is required")
    if channel_id == 0:
        raise SystemExit("DISCORD_CHANNEL_ID is required")

    app = Flask(__name__)
    app.register_blueprint(pulse_routes)

    flask_thread = threading.Thread(target=_run_flask, args=(app, port), daemon=True)
    flask_thread.start()

    bot = PulseBot(channel_id=channel_id)

    try:
        bot.run(token)
    except KeyboardInterrupt:
        pass


if __name__ == "__main__":
    main()
