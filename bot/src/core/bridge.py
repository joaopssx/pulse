import asyncio
from dataclasses import dataclass
from typing import Optional


@dataclass
class Alert:
    service_name: str
    status: str
    url: str
    latency_ms: int
    normal_latency_ms: int
    error_message: str
    downtime_minutes: int
    total_downtime: str
    sent_at: str


bot_loop: Optional[asyncio.AbstractEventLoop] = None
discord_bot = None


def set_runtime(loop: asyncio.AbstractEventLoop, bot) -> None:
    global bot_loop, discord_bot
    bot_loop = loop
    discord_bot = bot


def dispatch(alert: Alert) -> bool:
    if bot_loop is None or discord_bot is None:
        return False
    if not bot_loop.is_running():
        return False
    asyncio.run_coroutine_threadsafe(discord_bot.send_alert(alert), bot_loop)
    return True
