import os
from datetime import datetime
from urllib.parse import quote

import discord
import httpx
from discord import app_commands
from discord.ext import commands


PAGE_SIZE = 5
MAX_INCIDENTS = 10


def _api_url() -> str:
    return os.getenv("PULSE_API_URL", "http://localhost:8080").rstrip("/")


def _error_embed(message: str) -> discord.Embed:
    return discord.Embed(title="⚠️ Error", description=message, color=0xED4245)


def _to_unix(iso: str):
    if not iso:
        return None
    try:
        return int(datetime.fromisoformat(iso.replace("Z", "+00:00")).timestamp())
    except ValueError:
        return None


def _format_duration_ns(ns) -> str:
    if not ns:
        return "ongoing"
    try:
        seconds = int(ns) / 1_000_000_000
    except (TypeError, ValueError):
        return "—"
    if seconds < 60:
        return f"{seconds:.0f}s"
    minutes = seconds / 60
    if minutes < 60:
        return f"{minutes:.0f}m"
    return f"{minutes / 60:.1f}h"


def _build_pages(service_name: str, incidents: list) -> list:
    if not incidents:
        empty = discord.Embed(
            title=f"📜 History — {service_name}",
            description="No incidents recorded.",
            color=0x5865F2,
        )
        return [empty]

    incidents = incidents[:MAX_INCIDENTS]
    total_pages = (len(incidents) + PAGE_SIZE - 1) // PAGE_SIZE
    pages = []

    for start in range(0, len(incidents), PAGE_SIZE):
        chunk = incidents[start:start + PAGE_SIZE]
        embed = discord.Embed(
            title=f"📜 History — {service_name}",
            color=0x5865F2,
        )
        for inc in chunk:
            unix = _to_unix(str(inc.get("started_at", "")))
            started = f"<t:{unix}:f>" if unix else str(inc.get("started_at") or "—")
            duration = _format_duration_ns(inc.get("duration"))
            cause = (inc.get("cause") or "—")[:200]
            embed.add_field(
                name=started,
                value=f"Duration: `{duration}`\nCause: `{cause}`",
                inline=False,
            )
        embed.set_footer(text=f"Page {len(pages) + 1}/{total_pages}")
        pages.append(embed)

    return pages


class HistoryView(discord.ui.View):
    def __init__(self, pages: list):
        super().__init__(timeout=120.0)
        self.pages = pages
        self.index = 0
        self._sync()

    def _sync(self) -> None:
        self.prev.disabled = self.index == 0
        self.next.disabled = self.index >= len(self.pages) - 1

    @discord.ui.button(label="◀", style=discord.ButtonStyle.secondary)
    async def prev(self, interaction: discord.Interaction, button: discord.ui.Button) -> None:
        self.index = max(0, self.index - 1)
        self._sync()
        await interaction.response.edit_message(embed=self.pages[self.index], view=self)

    @discord.ui.button(label="▶", style=discord.ButtonStyle.secondary)
    async def next(self, interaction: discord.Interaction, button: discord.ui.Button) -> None:
        self.index = min(len(self.pages) - 1, self.index + 1)
        self._sync()
        await interaction.response.edit_message(embed=self.pages[self.index], view=self)


class HistoryCog(commands.Cog):
    def __init__(self, bot: commands.Bot):
        self.bot = bot

    @app_commands.command(name="history", description="Show recent incidents for a service")
    @app_commands.describe(service_name="Name of the service")
    async def history(self, interaction: discord.Interaction, service_name: str) -> None:
        await interaction.response.defer(ephemeral=False)

        url = f"{_api_url()}/api/services/{quote(service_name, safe='')}/incidents"
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                resp = await client.get(url)
                resp.raise_for_status()
                incidents = resp.json()
        except httpx.HTTPError as exc:
            await interaction.followup.send(embed=_error_embed(f"Failed to fetch incidents: {exc}"))
            return

        if not isinstance(incidents, list):
            incidents = []

        pages = _build_pages(service_name, incidents)
        if len(pages) > 1:
            await interaction.followup.send(embed=pages[0], view=HistoryView(pages))
        else:
            await interaction.followup.send(embed=pages[0])


async def setup(bot: commands.Bot) -> None:
    await bot.add_cog(HistoryCog(bot))
