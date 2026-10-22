package v1_12_2

import (
	"encoding/json"
	"strconv"

	"github.com/Mcd0LUO/gomc/internal/exec"
)

// chatComponent is a 1.12.2 chat component.
//
// 1.12.2 deserializes the JSON of `/tellraw` and `/title` into a tree of
// `IChatBaseComponent` (`IChatBaseComponent.ChatSerializer`), and the server
// resolves the parts that depend on the world before it sends the message:
// `ChatComponentUtils.filterForDisplay` replaces a `score` with the score of
// the holder and a `selector` with the names it matches.
//
// gomc has no client, so the tree is rendered to plain text. The styles
// (colour, bold, ...) are not part of the text; `clickEvent` and
// `hoverEvent` are parsed and kept, but they have no text of their own and
// 1.12.2 does not check their action (see buildClickEvent).
type chatComponent struct {
	// text is set for a component with a "text" field. A nil text means
	// the field was not there.
	text *string
	// translate is the key of a "translate" component.
	translate string
	// with are the arguments of a "translate" component.
	with []*chatComponent
	// score is the "score" component.
	score *chatScore
	// selector is the "selector" component.
	selector string
	// keybind is the "keybind" component.
	keybind string
	// extra are the siblings that come after this component.
	extra []*chatComponent
	// click and hover are the events, when the component has them.
	click *chatClickEvent
	hover *chatHoverEvent
}

// chatScore is the payload of a "score" component.
type chatScore struct {
	name      string
	objective string
}

// chatClickEvent is a "clickEvent".
//
// The action is kept as it was written: 1.12.2 does not check it.
type chatClickEvent struct {
	action string
	value  string
}

// chatHoverEvent is a "hoverEvent". The value is a component for
// `show_text`; for the other actions 1.12.2 keeps it as an item stack or an
// entity description, which gomc only keeps as the raw JSON.
type chatHoverEvent struct {
	action string
	value  *chatComponent
}

// parseChatComponent builds the tree of a JSON message.
//
// The shape follows `IChatBaseComponent.ChatSerializer`: a bare string is a
// text component, an array is a component whose first element is the root
// and whose other elements are its siblings, and an object is whichever
// component its first known field names -- `text`, `translate`, `score`,
// `selector`, `keybind` -- or an empty text component.
func parseChatComponent(raw string) (*chatComponent, error) {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, exec.Errorf("Invalid json: %s", err)
	}
	return buildChatComponent(v)
}

// buildChatComponent builds one component out of a decoded JSON value.
func buildChatComponent(v any) (*chatComponent, error) {
	switch value := v.(type) {
	case string:
		text := value
		return &chatComponent{text: &text}, nil
	case float64:
		// A JSON primitive is a text component: the adapter calls
		// `getAsString` on it.
		text := strconv.FormatFloat(value, 'f', -1, 64)
		return &chatComponent{text: &text}, nil
	case bool:
		text := strconv.FormatBool(value)
		return &chatComponent{text: &text}, nil
	case []any:
		if len(value) == 0 {
			return &chatComponent{}, nil
		}
		root, err := buildChatComponent(value[0])
		if err != nil {
			return nil, err
		}
		for _, item := range value[1:] {
			child, err := buildChatComponent(item)
			if err != nil {
				return nil, err
			}
			root.extra = append(root.extra, child)
		}
		return root, nil
	case map[string]any:
		return buildChatComponentObject(value)
	}
	return nil, exec.Errorf("Invalid json: unexpected message")
}

// buildChatComponentObject builds a component out of a JSON object.
func buildChatComponentObject(obj map[string]any) (*chatComponent, error) {
	c := &chatComponent{}
	switch {
	case obj["text"] != nil:
		text, err := chatPrimitive(obj["text"], "text")
		if err != nil {
			return nil, err
		}
		c.text = &text
	case obj["translate"] != nil:
		key, err := chatString(obj["translate"], "translate")
		if err != nil {
			return nil, err
		}
		c.translate = key
		if with, ok := obj["with"]; ok && with != nil {
			list, ok := with.([]any)
			if !ok {
				return nil, exec.Errorf("Invalid json: with is not a list")
			}
			for _, item := range list {
				arg, err := buildChatComponent(item)
				if err != nil {
					return nil, err
				}
				c.with = append(c.with, arg)
			}
		}
	case obj["score"] != nil:
		score, ok := obj["score"].(map[string]any)
		if !ok {
			return nil, exec.Errorf("Invalid json: score is not an object")
		}
		name, err := chatString(score["name"], "score name")
		if err != nil {
			return nil, err
		}
		objective, err := chatString(score["objective"], "score objective")
		if err != nil {
			return nil, err
		}
		c.score = &chatScore{name: name, objective: objective}
	case obj["selector"] != nil:
		sel, err := chatString(obj["selector"], "selector")
		if err != nil {
			return nil, err
		}
		c.selector = sel
	case obj["keybind"] != nil:
		key, err := chatString(obj["keybind"], "keybind")
		if err != nil {
			return nil, err
		}
		c.keybind = key
	}
	if extra, ok := obj["extra"]; ok && extra != nil {
		list, ok := extra.([]any)
		if !ok {
			return nil, exec.Errorf("Invalid json: extra is not a list")
		}
		for _, item := range list {
			child, err := buildChatComponent(item)
			if err != nil {
				return nil, err
			}
			c.extra = append(c.extra, child)
		}
	}
	if click, ok := obj["clickEvent"]; ok && click != nil {
		event, err := buildClickEvent(click)
		if err != nil {
			return nil, err
		}
		c.click = event
	}
	if hover, ok := obj["hoverEvent"]; ok && hover != nil {
		event, err := buildHoverEvent(hover)
		if err != nil {
			return nil, err
		}
		c.hover = event
	}
	return c, nil
}

// buildClickEvent reads a "clickEvent".
//
// 1.12.2 hands the event to Gson, which builds a `ChatClickable` out of the
// two strings and **does not check the action**: an action it does not know
// becomes a null enum and the message still goes out (the client then has
// nothing to do with it). Measured on a real server: a `tellraw` with
// `{"action":"bogus"}` succeeds.
func buildClickEvent(v any) (*chatClickEvent, error) {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, exec.Errorf("Invalid json: clickEvent is not an object")
	}
	action, err := chatOptionalString(obj["action"], "clickEvent action")
	if err != nil {
		return nil, err
	}
	value, err := chatOptionalString(obj["value"], "clickEvent value")
	if err != nil {
		return nil, err
	}
	return &chatClickEvent{action: action, value: value}, nil
}

// buildHoverEvent reads a "hoverEvent".
//
// Like `clickEvent`, the action is not checked: a name the server does not
// know is accepted and the message goes out.
func buildHoverEvent(v any) (*chatHoverEvent, error) {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, exec.Errorf("Invalid json: hoverEvent is not an object")
	}
	action, err := chatOptionalString(obj["action"], "hoverEvent action")
	if err != nil {
		return nil, err
	}
	event := &chatHoverEvent{action: action}
	if action == "show_text" && obj["value"] != nil {
		value, err := buildChatComponent(obj["value"])
		if err != nil {
			return nil, err
		}
		event.value = value
	}
	return event, nil
}

// chatString reads a string field of a component.
func chatString(v any, field string) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", exec.Errorf("Invalid json: %s is not a string", field)
	}
	return s, nil
}

// chatOptionalString reads a field that may be missing, which reads as the
// empty string.
func chatOptionalString(v any, field string) (string, error) {
	if v == nil {
		return "", nil
	}
	return chatPrimitive(v, field)
}

// chatPrimitive reads a field that Gson accepts as a string, which is a
// string and also a number or a boolean (it calls `getAsString` on the
// primitive).
func chatPrimitive(v any, field string) (string, error) {
	switch value := v.(type) {
	case string:
		return value, nil
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(value), nil
	}
	return "", exec.Errorf("Invalid json: %s is not a string", field)
}

// renderChatComponent renders a JSON message to the text a client shows.
//
// The recipient is the player the message goes to: it is what a `score`
