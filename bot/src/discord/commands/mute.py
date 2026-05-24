from datetime import datetime, timedelta, timezone

import discord
from discord import app_commands
from discord.ext import commands


_muted: dict = {}


def is_muted(service_name: str) -> bool:
    expire = _muted.get(service_name)
    if expire is None:
        return False
    if datetime.now(timezone.utc) >= expire:
        _muted.pop(service_name, None)
        return False
    return True


class MuteCog(commands.Cog):
    def __init__(self, bot: commands.Bot):
        self.bot = bot

    @app_commands.command(name="mute", description="Suppress alerts for a service")
    @app_commands.describe(service_name="Name of the service", minutes="Mute duration in minutes")
    async def mute(self, interaction: discord.Interaction, service_name: str, minutes: int) -> None:
        if minutes <= 0:
            embed = discord.Embed(
                title="⚠️ Invalid duration",
                description="Minutes must be greater than 0.",
                color=0xED4245,
            )
            await interaction.response.send_message(embed=embed, ephemeral=False)
            return

        until = datetime.now(timezone.utc) + timedelta(minutes=minutes)
        _muted[service_name] = until

        embed = discord.Embed(
            title=f"🔇 Muted — {service_name}",
            description=f"Alerts suppressed until <t:{int(until.timestamp())}:R>",
            color=0xFEE75C,
        )
        await interaction.response.send_message(embed=embed, ephemeral=False)

    @app_commands.command(name="unmute", description="Resume alerts for a service")
    @app_commands.describe(service_name="Name of the service")
    async def unmute(self, interaction: discord.Interaction, service_name: str) -> None:
        existed = _muted.pop(service_name, None) is not None
        if existed:
            embed = discord.Embed(
                title=f"🔔 Unmuted — {service_name}",
                description="Alerts resumed.",
                color=0x57F287,
            )
        else:
            embed = discord.Embed(
                title=f"🔔 {service_name}",
                description="Service was not muted.",
                color=0x5865F2,
            )
        await interaction.response.send_message(embed=embed, ephemeral=False)


async def setup(bot: commands.Bot) -> None:
    await bot.add_cog(MuteCog(bot))
