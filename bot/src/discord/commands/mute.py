from discord.ext import commands

from src.discord.bot import PulseBot


class MuteCommand(commands.Cog):
    def __init__(self, bot: PulseBot):
        self.bot = bot

    @commands.command(name="mute")
    async def mute(self, ctx: commands.Context, service: str = "", duration: str = "") -> None:
        _ = (ctx, service, duration)
        return None


async def setup(bot: PulseBot) -> None:
    await bot.add_cog(MuteCommand(bot))
