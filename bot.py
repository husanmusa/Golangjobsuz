import asyncio
import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, Any

from telegram import (
    InlineKeyboardButton,
    InlineKeyboardMarkup,
    Update,
)
from telegram.constants import ParseMode
from telegram.ext import (
    AIORateLimiter,
    ApplicationBuilder,
    CallbackQueryHandler,
    CommandHandler,
    ContextTypes,
    MessageHandler,
    filters,
)

DATA_FILE = Path("profiles.json")


def load_data() -> Dict[str, Any]:
    if DATA_FILE.exists():
        return json.loads(DATA_FILE.read_text())
    return {}


def save_data(data: Dict[str, Any]) -> None:
    DATA_FILE.write_text(json.dumps(data, indent=2, ensure_ascii=False))


def get_user_record(user_id: int, data: Dict[str, Any]) -> Dict[str, Any]:
    user_key = str(user_id)
    if user_key not in data:
        data[user_key] = {"draft": None, "profiles": []}
    return data[user_key]


def parse_document(file_name: str) -> Dict[str, str]:
    base_name = Path(file_name).stem
    now = datetime.utcnow().strftime("%Y-%m-%d")
    return {
        "full_name": base_name.replace("_", " ") or "Unknown",
        "email": f"{base_name.lower()}@example.com",
        "phone": "+1-555-0100",
        "summary": f"Parsed from {file_name} on {now}.",
    }


def build_review_keyboard(fields: Dict[str, str]) -> InlineKeyboardMarkup:
    buttons = []
    for field in fields:
        buttons.append([
            InlineKeyboardButton(f"Edit {field}", callback_data=f"edit:{field}"),
        ])
    buttons.append(
        [
            InlineKeyboardButton("Re-parse", callback_data="reparse"),
            InlineKeyboardButton("Manual correction", callback_data="manual"),
        ]
    )
    buttons.append([
        InlineKeyboardButton("Confirm profile", callback_data="confirm"),
        InlineKeyboardButton("Discard", callback_data="discard"),
    ])
    return InlineKeyboardMarkup(buttons)


def format_profile(profile: Dict[str, str]) -> str:
    lines = [f"<b>{k.capitalize()}</b>: {v}" for k, v in profile.items()]
    return "\n".join(lines)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    await update.message.reply_text(
        "Send me your CV or profile document to begin. Use /my_profile to view your latest saved profile."
    )


async def handle_document(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not update.message:
        return

    document = update.message.document
    data = load_data()
    record = get_user_record(update.effective_user.id, data)

    record["draft"] = {
        "status": "processing",
        "file_name": document.file_name or "uploaded_file",
        "fields": {},
    }
    save_data(data)

    processing_message = await update.message.reply_text(
        "Your document is being processed. Please wait…"
    )

    extracted = parse_document(record["draft"]["file_name"])
    record["draft"].update({"status": "review", "fields": extracted})
    save_data(data)

    await processing_message.edit_text(
        text=(
            "I extracted the following fields. Use the buttons below to edit or confirm."
        ),
        reply_markup=build_review_keyboard(extracted),
        parse_mode=ParseMode.HTML,
    )

    context.user_data["awaiting_field"] = None


async def callback_handler(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    query = update.callback_query
    await query.answer()

    data = load_data()
    record = get_user_record(update.effective_user.id, data)
    draft = record.get("draft")

    if not draft:
        await query.edit_message_text("No draft to review. Send a document first.")
        return

    if query.data == "reparse":
        draft["fields"] = parse_document(draft.get("file_name", "uploaded_file"))
        draft["status"] = "review"
        save_data(data)
        await query.edit_message_text(
            "Re-parsed your document. Review the fields again:",
            reply_markup=build_review_keyboard(draft["fields"]),
        )
        return

    if query.data == "manual":
        context.user_data["awaiting_field"] = "manual"
        await query.edit_message_text(
            "Send updates as `field: value` lines. When finished, type /done.",
            parse_mode=ParseMode.MARKDOWN,
        )
        return

    if query.data == "confirm":
        draft["status"] = "confirmed"
        version = len(record["profiles"]) + 1
        record["profiles"].append(
            {
                "version": version,
                "confirmed_at": datetime.utcnow().isoformat(),
                "fields": draft["fields"],
            }
        )
        record["draft"] = None
        save_data(data)
        await query.edit_message_text(
            f"Profile saved as version {version}. Use /my_profile to view or update."
        )
        return

    if query.data == "discard":
        record["draft"] = None
        save_data(data)
        await query.edit_message_text("Draft discarded.")
        return

    if query.data.startswith("edit:"):
        field = query.data.split(":", 1)[1]
        context.user_data["awaiting_field"] = field
        await query.edit_message_text(
            f"Send the new value for {field}."
        )
        return


async def handle_text(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not update.message:
        return

    awaiting_field = context.user_data.get("awaiting_field")
    if not awaiting_field:
        return

    data = load_data()
    record = get_user_record(update.effective_user.id, data)
    draft = record.get("draft")
    if not draft:
        await update.message.reply_text("No draft in progress.")
        return

    if awaiting_field == "manual":
        lines = update.message.text.splitlines()
        for line in lines:
            if ":" in line:
                key, value = line.split(":", 1)
                draft["fields"][key.strip()] = value.strip()
        save_data(data)
        await update.message.reply_text(
            "Applied manual updates. Use the buttons in your last message to continue."
        )
        return

    draft["fields"][awaiting_field] = update.message.text
    save_data(data)
    context.user_data["awaiting_field"] = None
    await update.message.reply_text(
        f"Updated {awaiting_field}. Return to the buttons above to continue."
    )


async def done(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    awaiting_field = context.user_data.get("awaiting_field")
    if awaiting_field:
        context.user_data["awaiting_field"] = None
        await update.message.reply_text("Stopped editing. Use the inline buttons to continue.")
    else:
        await update.message.reply_text("Nothing to finish right now.")


async def my_profile(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    data = load_data()
    record = get_user_record(update.effective_user.id, data)
    latest = record.get("profiles", [])[-1] if record.get("profiles") else None

    if not latest:
        await update.message.reply_text("No saved profiles yet. Send a document to start.")
        return

    keyboard = InlineKeyboardMarkup(
        [
            [InlineKeyboardButton("Update (new draft)", callback_data="start_update")],
            [InlineKeyboardButton("Delete profile", callback_data="delete_latest")],
        ]
    )

    await update.message.reply_text(
        f"Latest profile (v{latest['version']}):\n{format_profile(latest['fields'])}",
        parse_mode=ParseMode.HTML,
        reply_markup=keyboard,
    )


async def profile_actions(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    query = update.callback_query
    await query.answer()

    data = load_data()
    record = get_user_record(update.effective_user.id, data)

    if query.data == "start_update":
        draft_fields = record.get("profiles", [])[-1]["fields"].copy()
        record["draft"] = {
            "status": "review",
            "fields": draft_fields,
            "file_name": "manual_update",
        }
        save_data(data)
        await query.edit_message_text(
            "Started a new draft based on your latest profile.",
            reply_markup=build_review_keyboard(draft_fields),
        )
        return

    if query.data == "delete_latest":
        if record.get("profiles"):
            record["profiles"].pop()
            save_data(data)
            await query.edit_message_text("Latest profile deleted.")
        else:
            await query.edit_message_text("No profile to delete.")
        return


def build_app(token: str):
    return (
        ApplicationBuilder()
        .token(token)
        .rate_limiter(AIORateLimiter())
        .build()
    )


async def main() -> None:
    token = os.getenv("TELEGRAM_BOT_TOKEN")
    if not token:
        raise RuntimeError("TELEGRAM_BOT_TOKEN is not set")

    application = build_app(token)
    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("my_profile", my_profile))
    application.add_handler(CommandHandler("done", done))
    application.add_handler(MessageHandler(filters.Document.ALL, handle_document))
    application.add_handler(CallbackQueryHandler(callback_handler, pattern="^(edit:|reparse|manual|confirm|discard)$"))
    application.add_handler(CallbackQueryHandler(profile_actions, pattern="^(start_update|delete_latest)$"))
    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_text))

    await application.initialize()
    await application.start()
    await application.updater.start_polling()

    try:
        await asyncio.Event().wait()
    finally:
        await application.updater.stop()
        await application.stop()
        await application.shutdown()


if __name__ == "__main__":
    asyncio.run(main())
