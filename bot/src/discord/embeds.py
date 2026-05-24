from datetime import datetime, timezone

import discord

from src.core.bridge import Alert


FOOTER = "Pulse Monitor • use /status to view all services"


def _to_unix(iso: str) -> int:
    if iso:
        try:
            return int(datetime.fromisoformat(iso.replace("Z", "+00:00")).timestamp())
        except ValueError:
            pass
    return int(datetime.now(timezone.utc).timestamp())


def _status_icon(status: str) -> str:
    s = (status or "").lower()
    if s in ("healthy", "up"):
        return "🟢"
    if s in ("degraded", "slow"):
        return "🟡"
    if s == "down":
        return "🔴"
    return "⚪"


def build_down_embed(alert: Alert) -> discord.Embed:
    embed = discord.Embed(
        title=f"🔴 SERVICE DOWN — {alert.service_name}",
        color=0xED4245,
    )
    embed.add_field(name="URL", value=alert.url or "—", inline=False)
    embed.add_field(name="Error", value=alert.error_message or "—", inline=False)

    sent_unix = _to_unix(alert.sent_at)
    down_since = sent_unix - max(alert.downtime_minutes, 0) * 60
    embed.add_field(name="Down since", value=f"<t:{down_since}:R>", inline=False)

    embed.set_footer(text=FOOTER)
    return embed


def build_slow_embed(alert: Alert) -> discord.Embed:
    embed = discord.Embed(
        title=f"🟡 DEGRADED — {alert.service_name}",
        color=0xFEE75C,
    )
    embed.add_field(name="URL", value=alert.url or "—", inline=False)
    embed.add_field(name="Current latency", value=f"{alert.latency_ms} ms", inline=True)
    embed.add_field(name="Normal latency", value=f"{alert.normal_latency_ms} ms", inline=True)

    if alert.normal_latency_ms > 0:
        deviation = (alert.latency_ms - alert.normal_latency_ms) / alert.normal_latency_ms * 100.0
        sign = "+" if deviation >= 0 else ""
        embed.add_field(name="Deviation", value=f"{sign}{deviation:.0f}%", inline=True)
    else:
        embed.add_field(name="Deviation", value="—", inline=True)

    embed.set_footer(text=FOOTER)
    return embed


def build_recovered_embed(alert: Alert) -> discord.Embed:
    embed = discord.Embed(
        title=f"✅ RECOVERED — {alert.service_name}",
        color=0x57F287,
    )
    embed.add_field(name="URL", value=alert.url or "—", inline=False)
    downtime = alert.total_downtime or f"{alert.downtime_minutes} min"
    embed.add_field(name="Total downtime", value=downtime, inline=False)
    embed.set_footer(text=FOOTER)
    return embed


def build_ssl_embed(alert: Alert) -> discord.Embed:
    embed = discord.Embed(
        title=f"⚠️ SSL EXPIRING — {alert.service_name}",
        color=0xE67E22,
    )
    if alert.url:
        embed.add_field(name="URL", value=alert.url, inline=False)
    if alert.error_message:
        embed.add_field(name="Details", value=alert.error_message, inline=False)
    embed.set_footer(text=FOOTER)
    return embed


def build_status_embed(services: list[dict]) -> discord.Embed:
    embed = discord.Embed(
        title="🫀 Pulse — Service Status",
        color=0x5865F2,
    )

    if not services:
        embed.description = "No services configured."

    for svc in services:
        name = str(svc.get("name", "?"))
        status = str(svc.get("status", "unknown"))
        icon = _status_icon(status)
        uptime = svc.get("uptime_24h", 0.0)
        latency = svc.get("latency_ms", 0)

        try:
            uptime_str = f"{float(uptime):.2f}%"
        except (TypeError, ValueError):
            uptime_str = "—"

        value = f"{icon} `{status}` • uptime {uptime_str} • {latency} ms"
        embed.add_field(name=name, value=value, inline=False)

    now_unix = int(datetime.now(timezone.utc).timestamp())
    embed.set_footer(text=f"Last updated: <t:{now_unix}:R>")
    return embed
