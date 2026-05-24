import asyncio
from datetime import datetime, timezone

import discord
from discord.ext import commands

from src.core import bridge
from src.core.bridge import Alert
from src.discord.commands.mute import is_muted
from src.discord.embeds import (
    build_down_embed,
    build_recovered_embed,
    build_slow_embed,
    build_ssl_embed,
)


COG_MODULES = [
    "src.discord.commands.status",
    "src.discord.commands.history",
    "src.discord.commands.mute",
    "src.discord.commands.force",
]


class PulseBot(commands.Bot):
    def __init__(self, channel_id: int):
        intents = discord.Intents.default()
        intents.message_content = True
        super().__init__(command_prefix="!", intents=intents)
        self.channel_id = channel_id

    async def setup_hook(self) -> None:
        for module in COG_MODULES:
            await self.load_extension(module)
        await self.tree.sync()

    async def on_ready(self) -> None:
        print(f"pulse-bot connected as {self.user} ({self.user.id})")
        bridge.set_runtime(asyncio.get_running_loop(), self)

    async def send_alert(self, alert: Alert) -> None:
        if is_muted(alert.service_name):
            return

        channel = self.get_channel(self.channel_id)
        if channel is None:
            try:
                channel = await self.fetch_channel(self.channel_id)
            except discord.DiscordException:
                return

        embed = self._build_embed(alert)
        if embed is None:
            return

        message = await channel.send(embed=embed)

        if alert.status == "down":
            ts = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
            try:
                await message.create_thread(name=f"Incident — {alert.service_name} — {ts}")
            except discord.DiscordException:
                pass

    def _build_embed(self, alert: Alert):
        if alert.status == "down":
            return build_down_embed(alert)
        if alert.status == "slow":
            return build_slow_embed(alert)
        if alert.status == "recovered":
            return build_recovered_embed(alert)
        if alert.status == "ssl_expiring":
            return build_ssl_embed(alert)
        return None
