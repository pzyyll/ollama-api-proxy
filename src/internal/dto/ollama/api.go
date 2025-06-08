package ollama

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"reflect"
	"strings"
	"time"

	"ollama-api-proxy/src/internal/types/model"
)

// StatusError is an error with an HTTP status code and message.
type StatusError struct {
	StatusCode   int
	Status       string
	ErrorMessage string `json:"error"`
}

func (e StatusError) Error() string {
	switch {
	case e.Status != "" && e.ErrorMessage != "":
		return fmt.Sprintf("%s: %s", e.Status, e.ErrorMessage)
	case e.Status != "":
		return e.Status
	case e.ErrorMessage != "":
		return e.ErrorMessage
	default:
		// this should not happen
		return "something went wrong, please see the ollama server logs for details"
	}
}

type ImageData []byte

type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Index     int                       `json:"index,omitempty"`
	Name      string                    `json:"name"`
	Arguments ToolCallFunctionArguments `json:"arguments"`
}

type ToolCallFunctionArguments map[string]any

func (t *ToolCallFunctionArguments) String() string {
	bts, _ := json.Marshal(t)
	return string(bts)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// Thinking contains the text that was inside thinking tags in the
	// original model output when ChatRequest.Think is enabled.
	Thinking  string      `json:"thinking,omitempty"`
	Images    []ImageData `json:"images,omitempty"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

func (m *Message) UnmarshalJSON(b []byte) error {
	type Alias Message
	var a Alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	*m = Message(a)
	m.Role = strings.ToLower(m.Role)
	return nil
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	if d.Duration < 0 {
		return []byte("-1"), nil
	}
	return []byte("\"" + d.Duration.String() + "\""), nil
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	d.Duration = 5 * time.Minute

	switch t := v.(type) {
	case float64:
		if t < 0 {
			d.Duration = time.Duration(math.MaxInt64)
		} else {
			d.Duration = time.Duration(int(t) * int(time.Second))
		}
	case string:
		d.Duration, err = time.ParseDuration(t)
		if err != nil {
			return err
		}
		if d.Duration < 0 {
			d.Duration = time.Duration(math.MaxInt64)
		}
	default:
		return fmt.Errorf("Unsupported type: '%s'", reflect.TypeOf(v))
	}

	return nil
}

type Tool struct {
	Type     string       `json:"type"`
	Items    any          `json:"items,omitempty"`
	Function ToolFunction `json:"function"`
}

func (t Tool) String() string {
	bts, _ := json.Marshal(t)
	return string(bts)
}

type Tools []Tool

func (t Tools) String() string {
	bts, _ := json.Marshal(t)
	return string(bts)
}

// PropertyType can be either a string or an array of strings
type PropertyType []string

// UnmarshalJSON implements the json.Unmarshaler interface
func (pt *PropertyType) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*pt = []string{s}
		return nil
	}

	// If that fails, try to unmarshal as an array of strings
	var a []string
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*pt = a
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (pt PropertyType) MarshalJSON() ([]byte, error) {
	if len(pt) == 1 {
		// If there's only one type, marshal as a string
		return json.Marshal(pt[0])
	}
	// Otherwise marshal as an array
	return json.Marshal([]string(pt))
}

// String returns a string representation of the PropertyType
func (pt PropertyType) String() string {
	if len(pt) == 0 {
		return ""
	}
	if len(pt) == 1 {
		return pt[0]
	}
	return fmt.Sprintf("%v", []string(pt))
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  struct {
		Type       string   `json:"type"`
		Defs       any      `json:"$defs,omitempty"`
		Items      any      `json:"items,omitempty"`
		Required   []string `json:"required"`
		Properties map[string]struct {
			Type        PropertyType `json:"type"`
			Items       any          `json:"items,omitempty"`
			Description string       `json:"description"`
			Enum        []any        `json:"enum,omitempty"`
		} `json:"properties"`
	} `json:"parameters"`
}

func (t *ToolFunction) String() string {
	bts, _ := json.Marshal(t)
	return string(bts)
}

type ChatRequest struct {
	// Model is the model name, as in [GenerateRequest].
	Model string `json:"model"`

	// Messages is the messages of the chat - can be used to keep a chat memory.
	Messages []Message `json:"messages"`

	// Stream enables streaming of returned responses; true by default.
	Stream *bool `json:"stream,omitempty"`

	// Format is the format to return the response in (e.g. "json").
	Format json.RawMessage `json:"format,omitempty"`

	// KeepAlive controls how long the model will stay loaded into memory
	// following the request.
	KeepAlive *Duration `json:"keep_alive,omitempty"`

	// Tools is an optional list of tools the model has access to.
	Tools `json:"tools,omitempty"`

	// Options lists model-specific options.
	Options map[string]any `json:"options"`

	// Think controls whether thinking/reasoning models will think before
	// responding
	Think *bool `json:"think,omitempty"`
}

type ChatResponse struct {
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	Message    Message   `json:"message"`
	DoneReason string    `json:"done_reason,omitempty"`

	Done bool `json:"done"`

	Metrics
}

type Metrics struct {
	TotalDuration      time.Duration `json:"total_duration,omitempty"`
	LoadDuration       time.Duration `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       time.Duration `json:"eval_duration,omitempty"`
}

// Options specified in [GenerateRequest].  If you add a new option here, also
// add it to the API docs.
type Options struct {
	Runner

	// Predict options used at runtime
	NumKeep          int      `json:"num_keep,omitempty"`
	Seed             int      `json:"seed,omitempty"`
	NumPredict       int      `json:"num_predict,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	TopP             float32  `json:"top_p,omitempty"`
	MinP             float32  `json:"min_p,omitempty"`
	TypicalP         float32  `json:"typical_p,omitempty"`
	RepeatLastN      int      `json:"repeat_last_n,omitempty"`
	Temperature      float32  `json:"temperature,omitempty"`
	RepeatPenalty    float32  `json:"repeat_penalty,omitempty"`
	PresencePenalty  float32  `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32  `json:"frequency_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
}

// Runner options which must be set when the model is loaded into memory
type Runner struct {
	NumCtx    int   `json:"num_ctx,omitempty"`
	NumBatch  int   `json:"num_batch,omitempty"`
	NumGPU    int   `json:"num_gpu,omitempty"`
	MainGPU   int   `json:"main_gpu,omitempty"`
	UseMMap   *bool `json:"use_mmap,omitempty"`
	NumThread int   `json:"num_thread,omitempty"`
}

func (m *Metrics) Summary() {
	if m.TotalDuration > 0 {
		fmt.Fprintf(os.Stderr, "total duration:       %v\n", m.TotalDuration)
	}

	if m.LoadDuration > 0 {
		fmt.Fprintf(os.Stderr, "load duration:        %v\n", m.LoadDuration)
	}

	if m.PromptEvalCount > 0 {
		fmt.Fprintf(os.Stderr, "prompt eval count:    %d token(s)\n", m.PromptEvalCount)
	}

	if m.PromptEvalDuration > 0 {
		fmt.Fprintf(os.Stderr, "prompt eval duration: %s\n", m.PromptEvalDuration)
		fmt.Fprintf(os.Stderr, "prompt eval rate:     %.2f tokens/s\n", float64(m.PromptEvalCount)/m.PromptEvalDuration.Seconds())
	}

	if m.EvalCount > 0 {
		fmt.Fprintf(os.Stderr, "eval count:           %d token(s)\n", m.EvalCount)
	}

	if m.EvalDuration > 0 {
		fmt.Fprintf(os.Stderr, "eval duration:        %s\n", m.EvalDuration)
		fmt.Fprintf(os.Stderr, "eval rate:            %.2f tokens/s\n", float64(m.EvalCount)/m.EvalDuration.Seconds())
	}
}

func (opts *Options) FromMap(m map[string]any) error {
	valueOpts := reflect.ValueOf(opts).Elem() // names of the fields in the options struct
	typeOpts := reflect.TypeOf(opts).Elem()   // types of the fields in the options struct

	// build map of json struct tags to their types
	jsonOpts := make(map[string]reflect.StructField)
	for _, field := range reflect.VisibleFields(typeOpts) {
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" {
			jsonOpts[jsonTag] = field
		}
	}

	for key, val := range m {
		opt, ok := jsonOpts[key]
		if !ok {
			slog.Warn("invalid option provided", "option", key)
			continue
		}

		field := valueOpts.FieldByName(opt.Name)
		if field.IsValid() && field.CanSet() {
			if val == nil {
				continue
			}

			switch field.Kind() {
			case reflect.Int:
				switch t := val.(type) {
				case int64:
					field.SetInt(t)
				case float64:
					// when JSON unmarshals numbers, it uses float64, not int
					field.SetInt(int64(t))
				default:
					return fmt.Errorf("option %q must be of type integer", key)
				}
			case reflect.Bool:
				val, ok := val.(bool)
				if !ok {
					return fmt.Errorf("option %q must be of type boolean", key)
				}
				field.SetBool(val)
			case reflect.Float32:
				// JSON unmarshals to float64
				val, ok := val.(float64)
				if !ok {
					return fmt.Errorf("option %q must be of type float32", key)
				}
				field.SetFloat(val)
			case reflect.String:
				val, ok := val.(string)
				if !ok {
					return fmt.Errorf("option %q must be of type string", key)
				}
				field.SetString(val)
			case reflect.Slice:
				// JSON unmarshals to []any, not []string
				val, ok := val.([]any)
				if !ok {
					return fmt.Errorf("option %q must be of type array", key)
				}
				// convert []any to []string
				slice := make([]string, len(val))
				for i, item := range val {
					str, ok := item.(string)
					if !ok {
						return fmt.Errorf("option %q must be of an array of strings", key)
					}
					slice[i] = str
				}
				field.Set(reflect.ValueOf(slice))
			case reflect.Pointer:
				var b bool
				if field.Type() == reflect.TypeOf(&b) {
					val, ok := val.(bool)
					if !ok {
						return fmt.Errorf("option %q must be of type boolean", key)
					}
					field.Set(reflect.ValueOf(&val))
				} else {
					return fmt.Errorf("unknown type loading config params: %v %v", field.Kind(), field.Type())
				}
			default:
				return fmt.Errorf("unknown type loading config params: %v", field.Kind())
			}
		}
	}

	return nil
}

// DefaultOptions is the default set of options for [GenerateRequest]; these
// values are used unless the user specifies other values explicitly.
func DefaultOptions() Options {
	return Options{
		// options set on request to runner
		NumPredict: -1,

		// set a minimal num_keep to avoid issues on context shifts
		NumKeep:          4,
		Temperature:      0.8,
		TopK:             40,
		TopP:             0.9,
		TypicalP:         1.0,
		RepeatLastN:      64,
		RepeatPenalty:    1.1,
		PresencePenalty:  0.0,
		FrequencyPenalty: 0.0,
		Seed:             -1,

		Runner: Runner{
			// options set when the model is loaded
			NumCtx:    4096,
			NumBatch:  512,
			NumGPU:    -1, // -1 here indicates that NumGPU should be set dynamically
			NumThread: 0,  // let the runtime decide
			UseMMap:   nil,
		},
	}
}

type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type ListResponse struct {
	Models []ListModelResponse `json:"models"`
}

type ListModelResponse struct {
	Name         string             `json:"name"`
	Model        string             `json:"model"`
	ModifiedAt   time.Time          `json:"modified_at"`
	Size         int64              `json:"size"`
	Digest       string             `json:"digest"`
	Capabilities []model.Capability `json:"capabilities,omitempty"`
	Details      ModelDetails       `json:"details,omitempty"`
}

type ShowRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`

	// Template is deprecated
	Template string `json:"template"`
	Verbose  bool   `json:"verbose"`

	Options map[string]any `json:"options"`

	// Deprecated: set the model name with Model instead
	Name string `json:"name"`
}

type Tensor struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Shape []uint64 `json:"shape"`
}

// ShowResponse is the response returned from [Client.Show].
type ShowResponse struct {
	License       string             `json:"license,omitempty"`
	Modelfile     string             `json:"modelfile,omitempty"`
	Parameters    string             `json:"parameters,omitempty"`
	Template      string             `json:"template,omitempty"`
	System        string             `json:"system,omitempty"`
	Details       ModelDetails       `json:"details,omitempty"`
	Messages      []Message          `json:"messages,omitempty"`
	ModelInfo     map[string]any     `json:"model_info,omitempty"`
	ProjectorInfo map[string]any     `json:"projector_info,omitempty"`
	Tensors       []Tensor           `json:"tensors,omitempty"`
	Capabilities  []model.Capability `json:"capabilities,omitempty"`
	ModifiedAt    time.Time          `json:"modified_at,omitempty"`
}
