package googlesheet

import (
	"fmt"
	"strings"
	"sync"
)

type UserData struct {
	mu    sync.Mutex
	users map[string]string
}

var Users UserData

func init() {
	Users = UserData{users: make(map[string]string)}
}

func SetUserData(gs *GoogleSheets) {
	Users.mu.Lock()
	defer Users.mu.Unlock()

	data, err := gs.ReadRangeValues("1HL0DrSMNX7xOvHl57zpYwJ36ikSJjrRo5gzQMbWNbp4", "Оповещение", "A:C")
	if err != nil {
		gs.Error("Ошибка при  получении данных username: ", err)
		return
	}
	err = processUserData(data)
	if err != nil {
		gs.Error("Ошибка при  обработке данных, полученных из гугл таблиц: ", err)
		return
	}
}

func processUserData(data [][]string) error {
	newUsers := make(map[string]string)

	for _, row := range data {
		if len(row) != 3 {
			continue // Skip invalid rows
		}

		id := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		username := strings.TrimSpace(row[2])

		if id == "" || name == "" || username == "" {
			continue // Skip rows with empty fields
		}

		// Remove '@' from the beginning of the username if present
		username = strings.TrimPrefix(username, "@")

		// Add to the new map
		newUsers[username] = name
	}
	// Replace the old map with the new one
	Users.users = newUsers

	fmt.Printf("Processed %d valid user entries\n", len(newUsers))
	return nil
}

func GetUserName(userName string) string {
	Users.mu.Lock()
	defer Users.mu.Unlock()

	name, ok := Users.users[userName]
	if !ok {
		return ""
	}

	return name
}
