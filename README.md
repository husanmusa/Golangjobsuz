# Golangjobsuz bot

Telegram bot prototype for managing candidate profiles with a draft-to-confirmation review flow.

## Features
- Upload a document to create a `draft_profile`, trigger parsing, and show a processing message.
- Review extracted fields with inline keyboards to edit, re-parse, or manually correct data.
- Confirm drafts into persisted profiles with version history.
- Use `/my_profile` to view, update (create a new draft from the latest profile), or delete the latest saved profile.

## Running locally
1. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```
2. Set your bot token:
   ```bash
   export TELEGRAM_BOT_TOKEN="<your token>"
   ```
3. Start the bot:
   ```bash
   python bot.py
   ```

Profiles are stored in `profiles.json` in the repository root. Each confirmation appends a new version while keeping drafts separate.
