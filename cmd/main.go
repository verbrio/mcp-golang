package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metoro-io/mcp-golang"
	"io"
	"net/http"
	"strings"
)

type Content struct {
	Title       string  `json:"title" jsonschema:"description=The title to submit"`
	Description *string `json:"description,omitempty" jsonschema:"description=The description to submit"`
}
type MyFunctionsArguments struct {
	Submitter string  `json:"submitter" jsonschema:"description=The name of the thing calling this tool (openai, google, claude, etc)"`
	Content   Content `json:"content" jsonschema:"description=The content of the message"`
}

type ToggleLights struct {
	EntityID string `json:"entity_id,omitempty"`
}

type None struct{}

func main() {
	done := make(chan struct{})

	s := mcp.NewServer(mcp.NewStdioServerTransport())
	err := s.Tool("hello", "Say hello to a person", func(arguments MyFunctionsArguments) (mcp.ToolResponse, error) {
		return mcp.ToolResponse{Content: []mcp.Content{{Type: "text", Text: "Hello, " + arguments.Submitter + "!"}}}, nil
	})
	if err != nil {
		panic(err)
	}

	err = s.Tool("get_lights", "Get the state of all lights in the house", func(arguments None) (mcp.ToolResponse, error) {
		// Configuration
		hassURL := "http://home.net:8123"
		// Replace with your actual token
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiI1YmZkOGMwMmJjNGI0Y2ZkODdiZmExMTA5ZDQwZTg5YSIsImlhdCI6MTczMzY5NTc3MSwiZXhwIjoyMDQ5MDU1NzcxfQ.cCmgbnC_kXOIXgrmf59GPw8cYZYGx6pHjzQIZkYc72Q"

		lights, err := getLights(hassURL, token)
		if err != nil {
			return mcp.ToolResponse{}, err
		}
		output := displayLights(lights)
		return mcp.ToolResponse{Content: []mcp.Content{{Type: "text", Text: output}}}, nil
	})

	// Tool function to toggle lights
	err = s.Tool("control_light", "Toggle a specific light", func(arguments ToggleLights) (mcp.ToolResponse, error) {
		hassURL := "http://home.net:8123"
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiI1YmZkOGMwMmJjNGI0Y2ZkODdiZmExMTA5ZDQwZTg5YSIsImlhdCI6MTczMzY5NTc3MSwiZXhwIjoyMDQ5MDU1NzcxfQ.cCmgbnC_kXOIXgrmf59GPw8cYZYGx6pHjzQIZkYc72Q"

		err := toggleLight(hassURL, token, arguments.EntityID)
		if err != nil {
			return mcp.ToolResponse{}, err
		}

		output := fmt.Sprintf("Successfully toggled %s", arguments.EntityID)
		return mcp.ToolResponse{Content: []mcp.Content{{Type: "text", Text: output}}}, nil
	})

	err = s.Serve()
	if err != nil {
		panic(err)
	}

	<-done
}

type Entity struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
}

func controlLight(baseURL, token, entityID, state string, brightness int) error {
	if !strings.HasPrefix(entityID, "light.") {
		return fmt.Errorf("invalid entity ID format. Must start with 'light.'")
	}

	state = strings.ToLower(state)
	if state != "on" && state != "off" {
		return fmt.Errorf("invalid state. Must be 'on' or 'off'")
	}

	service := "turn_" + state
	endpoint := fmt.Sprintf("%s/api/services/light/%s", baseURL, service)

	command := struct {
		Entity_ID string                 `json:"entity_id"`
		Data      map[string]interface{} `json:"data,omitempty"`
	}{
		Entity_ID: entityID,
	}

	if state == "on" && brightness >= 0 {
		if brightness < 0 || brightness > 100 {
			return fmt.Errorf("brightness must be between 0 and 100")
		}
		hassbrightness := int(float64(brightness) / 100 * 255)
		command.Data = map[string]interface{}{
			"brightness": hassbrightness,
		}
	}

	jsonData, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Parse the error message
		body := resp.Body
		all, err := io.ReadAll(body)
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected status code: %d. Response body: %s. Request body: %s", resp.StatusCode, string(all), string(jsonData))
	}

	return nil
}

func getLights(baseURL, token string) ([]Entity, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/api/states", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var entities []Entity
	if err := json.Unmarshal(body, &entities); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Filter only light entities
	var lights []Entity
	for _, entity := range entities {
		if strings.HasPrefix(entity.EntityID, "light.") {
			lights = append(lights, entity)
		}
	}

	return lights, nil
}

func displayLights(lights []Entity) string {
	var output strings.Builder

	output.WriteString("Active Lights Status:\n")
	output.WriteString("=====================\n")

	for _, light := range lights {
		name := light.Attributes["friendly_name"]
		brightness := light.Attributes["brightness"]

		output.WriteString(fmt.Sprintf("\nLight: %s (%s)\n", name, light.EntityID))
		output.WriteString(fmt.Sprintf("State: %s\n", light.State))

		if brightness != nil {
			output.WriteString(fmt.Sprintf("Brightness: %.0f%%\n", float64(brightness.(float64))/255*100))
		}

		if light.State == "on" {
			if colorMode, ok := light.Attributes["color_mode"].(string); ok {
				output.WriteString(fmt.Sprintf("Color Mode: %s\n", colorMode))
			}

			if rgb, ok := light.Attributes["rgb_color"].([]interface{}); ok {
				output.WriteString(fmt.Sprintf("RGB Color: R:%v G:%v B:%v\n",
					int(rgb[0].(float64)),
					int(rgb[1].(float64)),
					int(rgb[2].(float64))))
			}
		}
		output.WriteString(fmt.Sprintf("Last Changed: %s\n", light.LastChanged))
	}

	return output.String()
}

func toggleLight(baseURL, token, entityID string) error {
	if !strings.HasPrefix(entityID, "light.") {
		return fmt.Errorf("invalid entity ID format. Must start with 'light.'")
	}

	endpoint := fmt.Sprintf("%s/api/services/light/toggle", baseURL)

	command := struct {
		Entity_ID string `json:"entity_id"`
	}{
		Entity_ID: entityID,
	}

	jsonData, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
