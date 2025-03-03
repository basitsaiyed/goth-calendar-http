package calendar

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func GetCalendarEvents(token string) ([]*calendar.Event, error) {
	ctx := context.Background()
	client := option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))

	srv, err := calendar.NewService(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	events, err := srv.Events.List("primary").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}

	return events.Items, nil
}

// GetUpcomingWeekEvents retrieves events for the upcoming week only
func GetUpcomingWeekEvents(token string) ([]*calendar.Event, error) {
	ctx := context.Background()
	client := option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))

	srv, err := calendar.NewService(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	// Calculate time bounds for the next week
	now := time.Now()
	oneWeekLater := now.AddDate(0, 0, 7)

	// Format times in RFC3339 format as required by Google Calendar API
	timeMin := now.Format(time.RFC3339)
	timeMax := oneWeekLater.Format(time.RFC3339)

	// Query events within the time range
	events, err := srv.Events.List("primary").
		TimeMin(timeMin).
		TimeMax(timeMax).
		OrderBy("startTime").
		SingleEvents(true). // Expand recurring events
		Do()

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}

	return events.Items, nil
}

func CreateEvent(token string, title, startTime, endTime, description string, attendees []string) error {
	ctx := context.Background()
	client := option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))

	srv, err := calendar.NewService(ctx, client)
	if err != nil {
		fmt.Println("Error creating Calendar service:", err)
		return fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	event := &calendar.Event{
		Summary:     title,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime,
			TimeZone: "Asia/Kolkata",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime,
			TimeZone: "Asia/Kolkata",
		},
	}

	var attendeeList []*calendar.EventAttendee
	for _, email := range attendees {
		attendeeList = append(attendeeList, &calendar.EventAttendee{Email: email})
	}

	event.Attendees = attendeeList
	createdEvent, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		fmt.Println("Error inserting event into calendar:", err)
		return fmt.Errorf("unable to create event: %v", err)
	}

	fmt.Println("Meeting Created:", createdEvent.HtmlLink)
	return nil
}
