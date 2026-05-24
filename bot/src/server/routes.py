from flask import Flask, request, jsonify

from src.discord.bot import PulseBot


def create_app(bot: PulseBot, secret: str) -> Flask:
    app = Flask(__name__)

    @app.route("/alert", methods=["POST"])
    def alert():
        _ = request
        _ = bot
        _ = secret
        return jsonify({"ok": True})

    @app.route("/health", methods=["GET"])
    def health():
        return jsonify({"status": "ok"})

    return app
