import os

from flask import Blueprint, Flask, jsonify, request

from src.core.bridge import Alert, dispatch
from src.discord.bot import PulseBot


pulse_routes = Blueprint("pulse_routes", __name__)


def _parse_alert(data: dict) -> Alert:
    return Alert(
        service_name=str(data.get("service_name", "")),
        status=str(data.get("status", "")),
        url=str(data.get("url", "")),
        latency_ms=int(data.get("latency_ms") or 0),
        normal_latency_ms=int(data.get("normal_latency_ms") or 0),
        error_message=str(data.get("error_message", "")),
        downtime_minutes=int(data.get("downtime_minutes") or 0),
        total_downtime=str(data.get("total_downtime", "")),
        sent_at=str(data.get("sent_at", "")),
    )


@pulse_routes.route("/alert", methods=["POST"])
def alert():
    expected = os.getenv("PULSE_SECRET", "")
    provided = request.headers.get("X-Pulse-Secret", "")
    if not expected or provided != expected:
        return jsonify({"error": "unauthorized"}), 401

    data = request.get_json(silent=True)
    if not isinstance(data, dict):
        return jsonify({"error": "invalid json"}), 400

    try:
        alert_obj = _parse_alert(data)
    except (TypeError, ValueError):
        return jsonify({"error": "invalid payload"}), 400

    dispatch(alert_obj)
    return jsonify({"ok": True}), 200


@pulse_routes.route("/health", methods=["GET"])
def health():
    return jsonify({"status": "ok"})


def create_app(bot: PulseBot, secret: str) -> Flask:
    app = Flask(__name__)
    if secret:
        os.environ.setdefault("PULSE_SECRET", secret)
    app.config["PULSE_BOT"] = bot
    app.register_blueprint(pulse_routes)
    return app
