import os
import requests
from datetime import datetime, timedelta
from dateutil.relativedelta import relativedelta
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

class ClockifyAPI:
    BASE_URL = "https://api.clockify.me/api/v1"
    
    def __init__(self):
        self.api_key = os.getenv("CLOCKIFY_API_KEY")
        if not self.api_key:
            raise ValueError("CLOCKIFY_API_KEY not found in environment variables")
        
        self.headers = {
            "X-Api-Key": self.api_key,
            "Content-Type": "application/json"
        }
        self.workspace_id = self._get_workspace_id()
        self.user_id = self._get_user_id()

    def _get_workspace_id(self):
        """Get the user's workspace ID"""
        response = requests.get(f"{self.BASE_URL}/workspaces", headers=self.headers)
        response.raise_for_status()
        return response.json()[0]["id"]  # Using the first workspace

    def _get_user_id(self):
        """Get the user's ID"""
        response = requests.get(f"{self.BASE_URL}/user", headers=self.headers)
        response.raise_for_status()
        return response.json()["id"]

    def get_projects(self):
        """Get all projects in the workspace"""
        response = requests.get(
            f"{self.BASE_URL}/workspaces/{self.workspace_id}/projects",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def get_tasks(self, project_id):
        """Get all tasks for a specific project"""
        response = requests.get(
            f"{self.BASE_URL}/workspaces/{self.workspace_id}/projects/{project_id}/tasks",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def has_time_entry(self, project_id, start_time, end_time):
        """Check if there's already a time entry for the given period"""
        # Format dates for the API
        start_str = start_time.strftime("%Y-%m-%dT%H:%M:%SZ")
        end_str = end_time.strftime("%Y-%m-%dT%H:%M:%SZ")
        
        params = {
            "start": start_str,
            "end": end_str,
            "project": project_id
        }
        
        response = requests.get(
            f"{self.BASE_URL}/workspaces/{self.workspace_id}/user/{self.user_id}/time-entries",
            headers=self.headers,
            params=params
        )
        response.raise_for_status()
        entries = response.json()
        
        return len(entries) > 0

    def add_time_entry(self, project_id, start_time, end_time, description="", task_id=None, billable=False):
        """Add a time entry"""
        payload = {
            "start": start_time.isoformat() + "Z",
            "end": end_time.isoformat() + "Z",
            "description": description,
            "projectId": project_id,
            "billable": str(billable).lower()  # API expects "true" or "false" as string
        }
        
        if task_id:
            payload["taskId"] = task_id
        
        response = requests.post(
            f"{self.BASE_URL}/workspaces/{self.workspace_id}/time-entries",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

def get_working_days(start_date, end_date):
    """Get list of working days (Mon-Fri) between start_date and end_date"""
    working_days = []
    current_date = start_date
    
    while current_date <= end_date:
        if current_date.weekday() < 5:  # Monday = 0, Friday = 4
            working_days.append(current_date)
        current_date += timedelta(days=1)
    
    return working_days

def get_description_mode():
    """Get the user's preferred mode for task descriptions"""
    print("\nHow would you like to handle task descriptions?")
    print("1. Use default description ('Standard workday') for all entries")
    print("2. Set one custom description for all entries")
    print("3. Enter custom description for each day")
    
    while True:
        try:
            choice = int(input("\nEnter your choice (1-3): "))
            if 1 <= choice <= 3:
                return choice
        except ValueError:
            pass
        print("Please enter a valid choice (1-3)")

def get_billable_preference():
    """Get user's preference for billable time"""
    while True:
        choice = input("\nMake entries billable? (y/N): ").lower()
        if choice in ['y', 'yes']:
            return True
        if choice in ['', 'n', 'no']:
            return False
        print("Please enter 'y' for yes or 'n' for no (default is no)")

def main():
    # Initialize Clockify API
    api = ClockifyAPI()
    
    # Get all projects
    projects = api.get_projects()
    print("\nAvailable Projects:")
    for idx, project in enumerate(projects):
        print(f"{idx + 1}. {project['name']}")
    
    # Let user select project
    while True:
        try:
            project_idx = int(input("\nSelect project number: ")) - 1
            if 0 <= project_idx < len(projects):
                break
            print(f"Please enter a number between 1 and {len(projects)}")
        except ValueError:
            print("Please enter a valid number")
    
    selected_project = projects[project_idx]
    
    # Get tasks for selected project
    tasks = api.get_tasks(selected_project["id"])
    selected_task = None
    
    if tasks:
        print("\nAvailable Tasks:")
        for idx, task in enumerate(tasks):
            print(f"{idx + 1}. {task['name']}")
        
        print("\nPress Enter to skip task selection or enter a task number:")
        task_input = input()
        
        if task_input.strip():
            try:
                task_idx = int(task_input) - 1
                if 0 <= task_idx < len(tasks):
                    selected_task = tasks[task_idx]
                else:
                    print("Invalid task number, proceeding without task selection")
            except ValueError:
                print("Invalid input, proceeding without task selection")
    else:
        print("\nNo tasks found for this project, proceeding without task selection")
    
    # Get description mode preference
    description_mode = get_description_mode()
    
    # Get billable preference
    billable = get_billable_preference()
    
    # Get custom description if mode 2 is selected
    default_description = "Standard workday"
    if description_mode == 2:
        default_description = input("\nEnter the description to use for all entries: ")
    
    # Calculate date range (start of current month to today)
    today = datetime.now()
    start_of_month = today.replace(day=1, hour=9, minute=0, second=0, microsecond=0)
    
    # Get working days
    working_days = get_working_days(start_of_month, today)
    
    # Add time entries for each working day
    skipped_count = 0
    added_count = 0
    
    for day in working_days:
        start_time = day.replace(hour=9, minute=0, second=0, microsecond=0)
        end_time = day.replace(hour=16, minute=30, second=0, microsecond=0)
        
        # Check if entry already exists
        if api.has_time_entry(selected_project["id"], start_time, end_time):
            print(f"Skipping {day.date()} - Time entry already exists")
            skipped_count += 1
            continue
        
        # Get description based on mode
        description = default_description
        if description_mode == 3:
            description = input(f"\nEnter description for {day.date()}: ")
        
        try:
            api.add_time_entry(
                selected_project["id"],
                start_time,
                end_time,
                description=description,
                task_id=selected_task["id"] if selected_task else None,
                billable=billable
            )
            print(f"Added time entry for {day.date()}")
            added_count += 1
        except requests.exceptions.HTTPError as e:
            print(f"Failed to add time entry for {day.date()}: {str(e)}")
    
    print(f"\nSummary: Added {added_count} entries, Skipped {skipped_count} existing entries")

if __name__ == "__main__":
    main()
