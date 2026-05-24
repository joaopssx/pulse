import discord
from discord.ext import commands

from src.core.bridge import Bridge


class PulseBot(commands.Bot):
    def __init__(self, channel_id: int, bridge: Bridge):
        intents = discord.Intents.default()
        intents.message_content = True
        super().__init__(command_prefix="!", intents=intents)

        self.channel_id = channel_id
        self.bridge = bridge

    async def setup_hook(self) -> None:
        return None

    async def send_alert(self, payload: dict) -> None:
        _ = payload
        return None
