package rules

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var regexCache sync.Map

type Evaluator struct {
	engine *Engine
}

func NewEvaluator(engine *Engine) *Evaluator {
	return &Evaluator{
		engine: engine,
	}
}

func (e *Evaluator) MatchRules(ctx context.Context, event Event) ([]Rule, error) {
	matched := []Rule{}

	for _, rule := range e.engine.rules {
		if rule.CanHandle(event) {
			matched = append(matched, rule)
		}
	}

	return matched, nil
}

func (e *Evaluator) ExecuteActions(ctx context.Context, event *Event, actions []Action) error {
	for _, action := range actions {
		if err := e.executeAction(ctx, event, action); err != nil {
			return fmt.Errorf("execute action %s: %w", action.Type, err)
		}
	}

	return nil
}

func (e *Evaluator) executeAction(ctx context.Context, event *Event, action Action) error {
	switch action.Type {
	case "log":
		return e.executeLogAction(action)
	case "reject":
		return e.executeRejectAction(action)
	case "modify":
		return e.executeModifyAction(event, action)
	case "approve":
		return nil
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *Evaluator) executeLogAction(action Action) error {
	// TODO: Implement logging action
	// This should log messages to a configured output (e.g., file, stdout, logging service)
	// Currently validates parameters but performs no actual logging
	if action.Comment != "" {
		return nil
	}

	message, ok := action.Params["message"].(string)
	if !ok {
		return fmt.Errorf("log action requires message parameter")
	}

	if message == "" {
		return nil
	}

	// TODO: Actually log the message
	return nil
}

func (e *Evaluator) executeRejectAction(action Action) error {
	reason, ok := action.Params["reason"].(string)
	if !ok || reason == "" {
		if action.Comment != "" {
			reason = action.Comment
		} else {
			reason = "rejected by rule"
		}
	}

	return fmt.Errorf("reject action: %s", reason)
}

func (e *Evaluator) executeModifyAction(event *Event, action Action) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	if event.Context == nil {
		return fmt.Errorf("event context is nil")
	}

	field, ok := action.Params["field"].(string)
	if !ok || field == "" {
		return fmt.Errorf("modify action requires field parameter")
	}

	value, ok := action.Params["value"]
	if !ok {
		return fmt.Errorf("modify action requires value parameter")
	}

	if _, exists := event.Context[field]; !exists {
		return fmt.Errorf("field %s does not exist in event context", field)
	}

	event.Context[field] = value

	if action.Comment != "" {
		fmt.Printf("[MODIFY] %s: set %s = %v\n", action.Comment, field, value)
	} else {
		fmt.Printf("[MODIFY] set %s = %v\n", field, value)
	}

	return nil
}

func EvaluateCondition(event Event, cond Condition) bool {
	var eventValue interface{}

	switch cond.Type {
	case "event_type":
		eventValue = event.Type
	case "source":
		eventValue = event.Source
	default:
		if event.Context != nil {
			if field, ok := cond.Params["field"].(string); ok {
				eventValue = event.Context[field]
			}
		}
	}

	return compareValues(eventValue, cond.Operator, cond.Value)
}

func compareValues(eventValue interface{}, operator string, targetValue interface{}) bool {
	switch operator {
	case "eq":
		return compareEqual(eventValue, targetValue)
	case "ne":
		return !compareEqual(eventValue, targetValue)
	case "gt":
		return compareGreater(eventValue, targetValue)
	case "lt":
		return compareLess(eventValue, targetValue)
	case "gte":
		return compareGreater(eventValue, targetValue) || compareEqual(eventValue, targetValue)
	case "lte":
		return compareLess(eventValue, targetValue) || compareEqual(eventValue, targetValue)
	case "contains":
		return compareContains(eventValue, targetValue)
	case "not_contains":
		return !compareContains(eventValue, targetValue)
	case "regex":
		return compareRegex(eventValue, targetValue)
	case "in":
		return compareIn(eventValue, targetValue)
	default:
		return false
	}
}

func compareEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func compareGreater(a, b interface{}) bool {
	return compareNumeric(a, b, func(x, y float64) bool { return x > y })
}

func compareLess(a, b interface{}) bool {
	return compareNumeric(a, b, func(x, y float64) bool { return x < y })
}

func compareNumeric(a, b interface{}, compareFunc func(float64, float64) bool) bool {
	aFloat, err1 := toFloat(a)
	bFloat, err2 := toFloat(b)

	if err1 != nil || err2 != nil {
		return false
	}

	return compareFunc(aFloat, bFloat)
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return rv.Float(), nil
		}
	}
	return 0, fmt.Errorf("cannot convert to float")
}

func compareContains(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	return strings.Contains(aStr, bStr)
}

func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	if cached, ok := regexCache.Load(pattern); ok {
		if regex, ok := cached.(*regexp.Regexp); ok {
			return regex, nil
		}
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	regexCache.Store(pattern, compiled)
	return compiled, nil
}

func compareRegex(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if bStr == "" {
		return false
	}

	re, err := getCompiledRegex(bStr)
	if err != nil {
		return false
	}

	return re.MatchString(aStr)
}

func compareIn(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)

	switch val := b.(type) {
	case []string:
		for _, item := range val {
			if item == aStr {
				return true
			}
		}
	case []interface{}:
		for _, item := range val {
			if fmt.Sprintf("%v", item) == aStr {
				return true
			}
		}
	}

	return false
}
