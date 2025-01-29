# Clockify Time Entry Bot

This bot helps you add standard workday time entries (9:00 AM to 4:30 PM) to Clockify for the current month-to-date.

## Setup

1. Install the required dependencies:
```bash
pip install -r requirements.txt
```

2. Get your Clockify API key:
   - Log in to [Clockify](https://clockify.me/)
   - Go to your Profile Settings
   - Copy your API key

3. Create a `.env` file in the project root and add your API key:
```
CLOCKIFY_API_KEY=your_api_key_here
```

## Usage

Run the script:
```bash
python main.py
```

The script will:
1. Show you a list of available projects
2. Let you select a project
3. If the project has tasks:
   - Show you a list of available tasks
   - Let you optionally select a task (press Enter to skip)
4. Ask how you want to handle task descriptions:
   - Use default description ("Standard workday") for all entries
   - Set one custom description for all entries
   - Enter a custom description for each day
5. Ask if entries should be billable (defaults to non-billable)
6. Add time entries (9:00 AM - 4:30 PM) for all working days (Mon-Fri) from the start of the current month until today
   - Automatically skips days that already have time entries
   - Shows a summary of added and skipped entries

## Features

- Automatically fetches projects and tasks from your Clockify workspace
- Adds time entries only for working days (Monday through Friday)
- Uses standard hours: 9:00 AM to 4:30 PM
- Skips weekends
- Shows progress as entries are added
- Flexible task description options:
  - Default description
  - Single custom description
  - Individual descriptions per day
- Works with projects that have no tasks
- Optional task selection
- Configurable billable status (defaults to non-billable)
- Intelligent duplicate prevention:
  - Checks for existing time entries
  - Automatically skips days with existing entries
  - Provides summary of added and skipped entries