from discord.ext import commands

from src.discord.bot import PulseBot


class StatusCommand(commands.Cog):
    def __init__(self, bot: PulseBot):
        self.bot = bot

    @commands.command(name="status")
    async def status(self, ctx: commands.Context) -> None:
        _ = ctx
        return None


async def setup(bot: PulseBot) -> None:
    await bot.add_cog(StatusCommand(bot))
