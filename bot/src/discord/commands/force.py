from discord.ext import commands

from src.discord.bot import PulseBot


class ForceCommand(commands.Cog):
    def __init__(self, bot: PulseBot):
        self.bot = bot

    @commands.command(name="force")
    async def force(self, ctx: commands.Context, service: str = "") -> None:
        _ = (ctx, service)
        return None


async def setup(bot: PulseBot) -> None:
    await bot.add_cog(ForceCommand(bot))
