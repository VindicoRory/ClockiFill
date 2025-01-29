package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const baseURL = "https://api.clockify.me/api/v1"

type ClockifyAPI struct {
	apiKey      string
	workspaceID string
	userID      string
	client      *http.Client
}

type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TimeEntry struct {
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
	ProjectID   string `json:"projectId"`
	TaskID      string `json:"taskId,omitempty"`
	Billable    string `json:"billable"`
}

func NewClockifyAPI() (*ClockifyAPI, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	apiKey := os.Getenv("CLOCKIFY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CLOCKIFY_API_KEY not found in environment variables")
	}

	api := &ClockifyAPI{
		apiKey: apiKey,
		client: &http.Client{},
	}

	var err error
	if api.workspaceID, err = api.getWorkspaceID(); err != nil {
		return nil, err
	}

	if api.userID, err = api.getUserID(); err != nil {
		return nil, err
	}

	return api, nil
}

func (api *ClockifyAPI) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Api-Key", api.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return api.client.Do(req)
}

func (api *ClockifyAPI) getWorkspaceID() (string, error) {
	resp, err := api.makeRequest("GET", "/workspaces", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var workspaces []Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
		return "", err
	}

	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces found")
	}

	return workspaces[0].ID, nil
}

func (api *ClockifyAPI) getUserID() (string, error) {
	resp, err := api.makeRequest("GET", "/user", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (api *ClockifyAPI) getProjects() ([]Project, error) {
	resp, err := api.makeRequest("GET", fmt.Sprintf("/workspaces/%s/projects", api.workspaceID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (api *ClockifyAPI) getTasks(projectID string) ([]Task, error) {
	resp, err := api.makeRequest("GET", fmt.Sprintf("/workspaces/%s/projects/%s/tasks", api.workspaceID, projectID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (api *ClockifyAPI) hasTimeEntry(projectID string, startTime, endTime time.Time) (bool, error) {
	params := fmt.Sprintf("?start=%s&end=%s&project=%s",
		startTime.UTC().Format(time.RFC3339),
		endTime.UTC().Format(time.RFC3339),
		projectID)

	endpoint := fmt.Sprintf("/workspaces/%s/user/%s/time-entries%s",
		api.workspaceID, api.userID, params)

	resp, err := api.makeRequest("GET", endpoint, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %v", err)
	}

	// Handle empty response
	if len(body) == 0 {
		return false, nil
	}

	// Try to decode the response
	var entries []interface{}
	if err := json.Unmarshal(body, &entries); err != nil {
		return false, fmt.Errorf("error decoding response (status %d): %v - body: %s",
			resp.StatusCode, err, string(body))
	}

	return len(entries) > 0, nil
}

func (api *ClockifyAPI) addTimeEntry(projectID string, startTime, endTime time.Time, description string, taskID string, billable bool) error {
	entry := TimeEntry{
		Start:       startTime.UTC().Format(time.RFC3339),
		End:         endTime.UTC().Format(time.RFC3339),
		Description: description,
		ProjectID:   projectID,
		Billable:    strconv.FormatBool(billable),
	}

	if taskID != "" {
		entry.TaskID = taskID
	}

	resp, err := api.makeRequest("POST", fmt.Sprintf("/workspaces/%s/time-entries", api.workspaceID), entry)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create time entry: %s", resp.Status)
	}

	return nil
}

func getWorkingDays(startDate, endDate time.Time) []time.Time {
	var workingDays []time.Time
	currentDate := startDate

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
			workingDays = append(workingDays, currentDate)
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return workingDays
}

func getDescriptionMode() int {
	fmt.Println("\nHow would you like to handle task descriptions?")
	fmt.Println("1. Use default description ('Standard workday') for all entries")
	fmt.Println("2. Set one custom description for all entries")
	fmt.Println("3. Enter custom description for each day")

	var choice int
	for {
		fmt.Print("\nEnter your choice (1-3): ")
		fmt.Scanln(&choice)
		if choice >= 1 && choice <= 3 {
			return choice
		}
		fmt.Println("Please enter a valid choice (1-3)")
	}
}

func getBillablePreference() bool {
	fmt.Print("\nMake entries billable? (y/N): ")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(input)
	return input == "y" || input == "yes"
}

func main() {
	api, err := NewClockifyAPI()
	if err != nil {
		fmt.Printf("Error initializing Clockify API: %v\n", err)
		return
	}

	// Get projects
	projects, err := api.getProjects()
	if err != nil {
		fmt.Printf("Error getting projects: %v\n", err)
		return
	}

	fmt.Println("\nAvailable Projects:")
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project.Name)
	}

	// Select project
	var projectIdx int
	for {
		fmt.Print("\nSelect project number: ")
		fmt.Scanln(&projectIdx)
		projectIdx--
		if projectIdx >= 0 && projectIdx < len(projects) {
			break
		}
		fmt.Printf("Please enter a number between 1 and %d\n", len(projects))
	}

	selectedProject := projects[projectIdx]

	// Get tasks
	tasks, err := api.getTasks(selectedProject.ID)
	if err != nil {
		fmt.Printf("Error getting tasks: %v\n", err)
		return
	}

	var selectedTask *Task
	if len(tasks) > 0 {
		fmt.Println("\nAvailable Tasks:")
		for i, task := range tasks {
			fmt.Printf("%d. %s\n", i+1, task.Name)
		}

		fmt.Print("\nPress Enter to skip task selection or enter a task number: ")
		var taskInput string
		fmt.Scanln(&taskInput)

		if taskInput != "" {
			taskIdx, err := strconv.Atoi(taskInput)
			if err == nil && taskIdx > 0 && taskIdx <= len(tasks) {
				selectedTask = &tasks[taskIdx-1]
			} else {
				fmt.Println("Invalid task number, proceeding without task selection")
			}
		}
	} else {
		fmt.Println("\nNo tasks found for this project, proceeding without task selection")
	}

	descriptionMode := getDescriptionMode()
	billable := getBillablePreference()

	defaultDescription := "Standard workday"
	if descriptionMode == 2 {
		fmt.Print("\nEnter the description to use for all entries: ")
		fmt.Scanln(&defaultDescription)
	}

	// Calculate date range
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 9, 0, 0, 0, now.Location())
	workingDays := getWorkingDays(startOfMonth, now)

	skippedCount := 0
	addedCount := 0

	for _, day := range workingDays {
		startTime := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, day.Location())
		endTime := time.Date(day.Year(), day.Month(), day.Day(), 16, 30, 0, 0, day.Location())

		hasEntry, err := api.hasTimeEntry(selectedProject.ID, startTime, endTime)
		if err != nil {
			fmt.Printf("Error checking time entry for %s: %v\n", day.Format("2006-01-02"), err)
			continue
		}

		if hasEntry {
			fmt.Printf("Skipping %s - Time entry already exists\n", day.Format("2006-01-02"))
			skippedCount++
			continue
		}

		description := defaultDescription
		if descriptionMode == 3 {
			fmt.Printf("\nEnter description for %s: ", day.Format("2006-01-02"))
			fmt.Scanln(&description)
		}

		taskID := ""
		if selectedTask != nil {
			taskID = selectedTask.ID
		}

		if err := api.addTimeEntry(selectedProject.ID, startTime, endTime, description, taskID, billable); err != nil {
			if strings.Contains(err.Error(), "EOF") {
				fmt.Printf("Skipping %s - Unable to verify existing entries\n", day.Format("2006-01-02"))
			} else {
				fmt.Printf("Failed to add time entry for %s: %v\n", day.Format("2006-01-02"), err)
			}
			continue
		}

		fmt.Printf("Added time entry for %s\n", day.Format("2006-01-02"))
		addedCount++
	}

	fmt.Printf("\nSummary: Added %d entries, Skipped %d existing entries\n", addedCount, skippedCount)
}
