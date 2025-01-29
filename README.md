# ClockiFill

ClockiFill is a command-line tool that helps automate time entry management in Clockify. It's particularly useful for bulk-creating time entries for working days within the current month, using standard working hours (9:00 AM - 4:30 PM).

## Quick Start

1. Download the latest release from [GitHub Releases](https://github.com/VindicoRory/ClockiFill/releases/latest)

2. Get your Clockify API key:
   - Log into [Clockify](https://clockify.me/)
   - Click on your profile picture in the top right
   - Go to "Profile Settings"
   - Scroll down to "API"
   - Copy your API key or click "Generate" to create a new one

3. Create a `.env` file in the same directory as the binary:
   ```
   CLOCKIFY_API_KEY=your_api_key_here
   ```
   Replace `your_api_key_here` with the API key you copied

4. Run the program:
   ```bash
   # Windows
   clockifill.exe

   # Mac/Linux
   ./clockifill
   ```

## Usage Guide

When you run ClockiFill, it will:

1. Show you a list of your Clockify projects
2. Ask you to select a project number
3. If the project has tasks, offer you to select one (optional)
4. Ask how you want to handle descriptions:
   - Option 1: Use "Standard workday" for all entries
   - Option 2: Set one custom description for all entries
   - Option 3: Enter a description for each day
5. Ask if the entries should be billable (y/N)

The program will then create time entries for all working days (Monday-Friday) from the start of the current month up to today, skipping any days that already have entries.

## Features

- Automatically detects working days (Monday-Friday)
- Prevents duplicate entries
- Standard working hours (9:00 AM - 4:30 PM)
- Interactive project and task selection
- Flexible description options
- Billable/non-billable tracking

## Troubleshooting

Common issues and solutions:

- **"API key not found"**: Make sure your `.env` file is in the same directory as the binary
- **"No projects found"**: Verify your API key is correct
- **"EOF error"**: This can occur when checking future dates - it's safe to ignore
- **Rate limiting**: If you see API errors, try running the program again

## Building from Source

If you want to build the program yourself instead of using the pre-built binaries:

1. Install Go 1.19 or later
2. Clone the repository
3. Run:
   ```bash
   go mod init clockifill
   go mod tidy
   go build
   ```
