import os

import discord
import httpx
from discord import app_commands
from discord.ext import commands

from src.discord.embeds import build_status_embed


def _api_url() -> str:
    return os.getenv("PULSE_API_URL", "http://localhost:8080").rstrip("/")


def _error_embed(message: str) -> discord.Embed:
    return discord.Embed(title="⚠️ Error", description=message, color=0xED4245)


class StatusCog(commands.Cog):
    def __init__(self, bot: commands.Bot):
        self.bot = bot

    @app_commands.command(name="status", description="Show current status of all services")
    async def status(self, interaction: discord.Interaction) -> None:
        await interaction.response.defer(ephemeral=False)

        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                resp = await client.get(f"{_api_url()}/api/services")
                resp.raise_for_status()
                services = resp.json()
        except httpx.HTTPError as exc:
            await interaction.followup.send(embed=_error_embed(f"Failed to fetch services: {exc}"))
            return

        if not isinstance(services, list):
            services = []

        await interaction.followup.send(embed=build_status_embed(services))


async def setup(bot: commands.Bot) -> None:
    await bot.add_cog(StatusCog(bot))
