import os
from urllib.parse import quote

import discord
import httpx
from discord import app_commands
from discord.ext import commands


def _api_url() -> str:
    return os.getenv("PULSE_API_URL", "http://localhost:8080").rstrip("/")


def _secret() -> str:
    return os.getenv("PULSE_SECRET", "")


def _error_embed(message: str) -> discord.Embed:
    return discord.Embed(title="⚠️ Error", description=message, color=0xED4245)


class ForceCog(commands.Cog):
    def __init__(self, bot: commands.Bot):
        self.bot = bot

    @app_commands.command(name="force", description="Run an immediate check for a service")
    @app_commands.describe(service_name="Name of the service")
    async def force(self, interaction: discord.Interaction, service_name: str) -> None:
        await interaction.response.defer(ephemeral=False)

        url = f"{_api_url()}/api/services/{quote(service_name, safe='')}/force-check"
        headers = {"X-Pulse-Secret": _secret()}
        try:
            async with httpx.AsyncClient(timeout=15.0) as client:
                resp = await client.post(url, headers=headers)
                resp.raise_for_status()
                data = resp.json()
        except httpx.HTTPError as exc:
            await interaction.followup.send(embed=_error_embed(f"Failed to force check: {exc}"))
            return

        if not isinstance(data, dict):
            data = {}

        success = bool(data.get("success", False))
        status_code = data.get("status_code", "—")
        latency = data.get("latency_ms", "—")
        err = str(data.get("error") or "")

        embed = discord.Embed(
            title=("✅ Force check — " if success else "❌ Force check — ") + service_name,
            color=0x57F287 if success else 0xED4245,
        )
        embed.add_field(name="Status code", value=str(status_code), inline=True)
        embed.add_field(name="Latency", value=f"{latency} ms", inline=True)
        if err:
            embed.add_field(name="Error", value=err[:500], inline=False)

        await interaction.followup.send(embed=embed)


async def setup(bot: commands.Bot) -> None:
    await bot.add_cog(ForceCog(bot))
