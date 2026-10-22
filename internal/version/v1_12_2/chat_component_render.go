package v1_12_2

import (
	"strconv"
	"strings"

	"github.com/Mcd0LUO/gomc/internal/exec"
	"github.com/Mcd0LUO/gomc/internal/world"
)

func renderChatComponent(ctx *exec.Context, raw string, recipient *world.Entity) (string, error) {
	c, err := parseChatComponent(raw)
	if err != nil {
		return "", err
	}
	return c.render(ctx, recipient)
}

// render returns the plain text of the component and its siblings.
func (c *chatComponent) render(ctx *exec.Context, recipient *world.Entity) (string, error) {
	var b strings.Builder
	switch {
	case c.text != nil:
		b.WriteString(*c.text)
	case c.translate != "":
		text, err := c.renderTranslate(ctx, recipient)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
	case c.score != nil:
		text, err := c.renderScore(ctx, recipient)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
	case c.selector != "":
		text, err := renderSelectorComponent(ctx, c.selector)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
	case c.keybind != "":
		// A client shows the key bound to the action. gomc has no
		// key bindings, so the key itself is the closest it gets.
		b.WriteString(c.keybind)
	}
	for _, child := range c.extra {
		text, err := child.render(ctx, recipient)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
	}
	return b.String(), nil
}

// renderScore renders a "score" component.
//
// `ChatComponentUtils.filterForDisplay` resolves the name first: a selector
// has to match exactly one entity (otherwise the command fails with
// `Selector '<x>' found nothing`), a player gives its name and any other
// entity gives its UUID. Then a name of "*" is replaced by the name of the
// player the message goes to -- it is **not** an aggregation over the
// objective. Finally the score is looked up: a holder without an entry
// renders as the empty string (`ChatComponentScore.a`).
func (c *chatComponent) renderScore(ctx *exec.Context, recipient *world.Entity) (string, error) {
	name := c.score.name
	if IsSelector(name) {
		entities, err := resolveSelectorNames(ctx, name)
		if err != nil {
			return "", err
		}
		if len(entities) != 1 {
			return "", exec.Errorf("Selector '%s' found nothing", name)
		}
		e := entities[0]
		if e.IsPlayer {
			name = e.DisplayName()
		} else {
			name = e.EnsureUUID()
		}
	}
	if name == "*" && recipient != nil {
		name = recipient.DisplayName()
	}
	value, ok := ctx.World.Scoreboard.Score(name, c.score.objective)
	if !ok {
		return "", nil
	}
	return strconv.Itoa(value), nil
}

// renderSelector renders a "selector" component.
//
// `PlayerSelector.getPlayerNames` resolves the selector and joins the
// display names; a selector that matches nothing gives an empty component.
func renderSelectorComponent(ctx *exec.Context, text string) (string, error) {
	entities, err := resolveSelectorNames(ctx, text)
	if err != nil {
		return "", err
	}
	if len(entities) == 0 {
		return "", nil
	}
	names := make([]string, 0, len(entities))
	for _, e := range entities {
		names = append(names, scoreboardDisplayName(e))
	}
	return joinVanillaList(names), nil
}

// resolveSelectorNames resolves a selector the way a chat component does.
func resolveSelectorNames(ctx *exec.Context, text string) ([]*world.Entity, error) {
	sel, err := ParseSelector(text)
	if err != nil {
		return nil, exec.Errorf("%v", err)
	}
	resolver := &SelectorResolver{World: ctx.World}
	entities, err := resolver.Resolve(ctx.Source, sel)
	selectorWarnings(ctx, resolver)
	if err != nil {
		return nil, exec.Errorf("%v", err)
	}
	return entities, nil
}

// renderTranslate renders a "translate" component.
//
// The key is looked up in the language table; a key the table does not know
// renders as itself, which is what 1.12.2 does with a missing translation.
// The arguments are formatted into it with the rules of `String.format`.
func (c *chatComponent) renderTranslate(ctx *exec.Context, recipient *world.Entity) (string, error) {
	format, ok := chatTranslation(c.translate)
	if !ok {
		format = c.translate
	}
	args := make([]string, 0, len(c.with))
	for _, arg := range c.with {
		text, err := arg.render(ctx, recipient)
		if err != nil {
			return "", err
		}
		args = append(args, text)
	}
	return javaFormat(format, args), nil
}

// chatTranslation looks a language key up.
//
// gomc carries no language file: the messages it prints are copied into the
// code. A `translate` component can name any key, so the table covers the
// keys a command message uses plus the entity, item and block names gomc
// knows. Everything else renders as the key.
func chatTranslation(key string) (string, bool) {
	switch key {
	case "chat.type.text":
		return "<%s> %s", true
	case "chat.type.announcement":
		return "[%s] %s", true
	}
	if name, ok := chatNamedKey(key, "entity.", ".name"); ok {
		return world.EntityTypeMeta(name).DisplayName, true
	}
	if name, ok := chatNamedKey(key, "item.", ".name"); ok {
		if display, ok := world.ItemDisplayName("minecraft:" + name); ok {
			return display, true
		}
	}
	if name, ok := chatNamedKey(key, "tile.", ".name"); ok {
		return world.BlockDisplayName(name), true
	}
	return "", false
}

// chatNamedKey strips a prefix and a suffix off a language key.
func chatNamedKey(key, prefix, suffix string) (string, bool) {
	if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
		return "", false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(key, prefix), suffix)
	return name, name != ""
}

// javaFormat applies the subset of `String.format` a chat message uses:
// `%s`, `%d`, `%%`, an explicit argument index (`%1$s`) and the flags that
// come before the conversion. A conversion with no argument left is left
// as it is.
func javaFormat(format string, args []string) string {
	var b strings.Builder
	next := 0
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b.WriteByte(format[i])
			continue
		}
		if i+1 < len(format) && format[i+1] == '%' {
			b.WriteByte('%')
			i++
			continue
		}
		j := i + 1
		index := -1
		start := j
		for j < len(format) && format[j] >= '0' && format[j] <= '9' {
			j++
		}
		if j > start && j < len(format) && format[j] == '$' {
			n, _ := strconv.Atoi(format[start:j])
			index = n - 1
			j++
		} else {
			j = start
		}
		for j < len(format) && strings.ContainsRune("-#+ 0,(", rune(format[j])) {
			j++
		}
		for j < len(format) && (format[j] >= '0' && format[j] <= '9' || format[j] == '.') {
			j++
		}
		if j >= len(format) {
			b.WriteString(format[i:])
			break
		}
		conversion := format[j]
		if conversion != 's' && conversion != 'd' {
			b.WriteString(format[i : j+1])
			i = j
			continue
		}
		arg := ""
		found := false
		if index >= 0 {
			if index < len(args) {
				arg, found = args[index], true
			}
		} else if next < len(args) {
			arg, found = args[next], true
			next++
		}
		if !found {
			b.WriteString(format[i : j+1])
		} else {
			b.WriteString(arg)
		}
		i = j
	}
	return b.String()
}

// chatComponentJSONError reports the parse failure of a JSON message.
//
// It is used by `/title`, which keeps the message as the JSON the client
// gets, but still has to reject one that 1.12.2 would reject.
func chatComponentJSONError(raw string) error {
	_, err := parseChatComponent(raw)
	return err
}
