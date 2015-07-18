package checks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/daemonl/informer/reporter"
)

type JSONCheck struct {
	Request
	Time   []JSONTime   `xml:"time"`
	Number []JSONNumber `xml:"number"`
}

type JSONElem struct {
	Key string `xml:"key,attr"`
}

func (je JSONElem) GetValue(ctx *JSONContext) interface{} {
	return walk(je.Key, ctx.Data)
}

func (je JSONElem) GetString(ctx *JSONContext) string {
	return fmt.Sprint(je.GetValue(ctx))
}

func (je JSONElem) GetInt(ctx *JSONContext) (int64, error) {
	switch val := je.GetValue(ctx).(type) {
	case string:
		return strconv.ParseInt(val, 10, 64)
	case float64:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("Can't convert %s (%T) to Int", je.Key, val)
	}
}

func (je JSONElem) GetFloat(ctx *JSONContext) (float64, error) {
	switch val := je.GetValue(ctx).(type) {
	case string:
		return strconv.ParseFloat(val, 64)
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("Can't convert %s (%T) to Float", je.Key, val)
	}
}

func walk(key string, data interface{}) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		return m[key]
	}
	nextVal := m[parts[0]]
	nextKey := strings.Join(parts[1:], ".")
	return walk(nextKey, nextVal)
}

type JSONTime struct {
	JSONElem
	Age    string `xml:"age"`
	Format string `xml:"format"`
}

func (jt *JSONTime) Check(ctx *JSONContext) error {
	var t time.Time
	var err error
	if jt.Format == "UNIX" {
		i, err := jt.GetInt(ctx)
		if err != nil {
			return err
		}
		t = time.Unix(i, 0)
	} else {
		t, err = time.Parse(jt.Format, jt.GetString(ctx))
		if err != nil {
			return err
		}
	}

	if len(jt.Age) > 0 {
		duration, err := time.ParseDuration(jt.Age)
		if err != nil {
			return err
		}
		since := time.Since(t)
		if since > duration {
			ctx.Fail("Elem %s was %s ago, > %s", jt.Key, since, jt.Age)
		}
	}
	return nil
}

type JSONNumber struct {
	JSONElem
	Min *float64 `xml:"min"`
	Max *float64 `xml:"max"`
	Eq  *float64 `xml:"eq"`
	Neq *float64 `xml:"neq"`
}

func (jt *JSONNumber) Check(ctx *JSONContext) error {
	val, err := jt.GetFloat(ctx)
	if err != nil {
		return err
	}
	if jt.Min != nil && val < *jt.Min {
		ctx.Fail("Elem %s: %f < %f", jt.Key, val, *jt.Min)
	}
	if jt.Max != nil && val > *jt.Max {
		ctx.Fail("Elem %s: %f > %f", jt.Key, val, *jt.Max)
	}
	if jt.Eq != nil && val != *jt.Eq {
		ctx.Fail("Elem %s: %f != %f", jt.Key, val, *jt.Eq)
	}
	if jt.Neq != nil && val == *jt.Neq {
		ctx.Fail("Elem %s: %f == %f", jt.Key, val, *jt.Neq)
	}
	return nil
}

type JSONContext struct {
	Data map[string]interface{}
	*reporter.Report
}

func (t *JSONCheck) RunCheck(r *reporter.Reporter) error {
	res := r.Report("JSON CHECK %s", t.Url)

	resp, err := t.Request.DoRequest()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	data := map[string]interface{}{}
	dec.Decode(&data)
	ctx := &JSONContext{
		Data:   data,
		Report: res,
	}

	for _, test := range t.Number {
		err := test.Check(ctx)
		if err != nil {
			return err
		}
	}
	for _, test := range t.Time {
		err := test.Check(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
