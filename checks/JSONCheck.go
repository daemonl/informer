package checks

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/daemonl/informer/reporter"
)

type JSONCheck struct {
	Request
	CheckFields []JSONFieldCheck `xml:",any"`
}

func (t *JSONCheck) GetHash() string {
	return hashFromf("JSON:%s %s", t.CheckFields, t.Request.HashBase())
}

type JSONElemDef struct {
	Key string `xml:"key,attr"`
}

func (je *JSONElemDef) GetName() string {
	return je.Key
}
func (je *JSONElemDef) GetElemDef() *JSONElemDef {
	return je
}

type JSONTimeDef struct {
	*JSONElemDef
	Age    string `xml:"age"`
	Format string `xml:"format,attr"`
}

type JSONNumberDef struct {
	*JSONElemDef
	Min *float64 `xml:"min"`
	Max *float64 `xml:"max"`
	Eq  *float64 `xml:"eq"`
	Neq *float64 `xml:"neq"`
}

type JSONStringDef struct {
	*JSONElemDef
	Eq  *float64 `xml:"eq"`
	Neq *float64 `xml:"neq"`
}

type JSONFieldCheck struct {
	JSONFieldCheckDef
}

type JSONFieldCheckDef interface {
	Check(je *JSONElem) error
	GetName() string
	GetElemDef() *JSONElemDef
}

func (jfc *JSONFieldCheck) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case "time":
		jfc.JSONFieldCheckDef = &JSONTimeDef{}
	case "number":
		jfc.JSONFieldCheckDef = &JSONNumberDef{}
	case "string":
		jfc.JSONFieldCheckDef = &JSONStringDef{}
	default:
		return fmt.Errorf("No JSON Field check type %s", start.Name.Local)
	}
	return d.DecodeElement(jfc.JSONFieldCheckDef, &start)
}

func (jfc *JSONFieldCheck) DoChecks(ctx *JSONContext, r *reporter.Reporter) {
	res := r.Report("Elem: %s", jfc.GetName())
	je := &JSONElem{
		ctx:         ctx,
		Fails:       []string{},
		JSONElemDef: jfc.GetElemDef(),
	}
	err := jfc.JSONFieldCheckDef.Check(je)
	if err != nil {
		res.Fail(err.Error())
		return
	}
	if len(je.Fails) > 0 {
		for _, f := range je.Fails {
			res.Fail(f)
		}
	} else {
		res.Pass(je.GetString())
	}
}

type JSONContext struct {
	Data map[string]interface{}
	*reporter.Reporter
}

type JSONElem struct {
	ctx   *JSONContext
	Fails []string
	*JSONElemDef
}

func (je JSONElem) GetValue() interface{} {
	return walk(je.Key, je.ctx.Data)
}

func (je JSONElem) GetString() string {
	return fmt.Sprint(je.GetValue())
}

func (je JSONElem) GetInt() (int64, error) {
	switch val := je.GetValue().(type) {
	case string:
		return strconv.ParseInt(val, 10, 64)
	case float64:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("Can't convert %s (%T) to Int", je.Key, val)
	}
}

func (je JSONElem) GetFloat() (float64, error) {
	switch val := je.GetValue().(type) {
	case string:
		return strconv.ParseFloat(val, 64)
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("Can't convert %s (%T) to Float", je.Key, val)
	}
}

func (je *JSONElem) Failf(msg string, params ...interface{}) {
	je.Fail(fmt.Sprintf(msg, params...))
}

func (je *JSONElem) Fail(msg string) {
	je.Fails = append(je.Fails, fmt.Sprintf("Elem %s: %s", je.Key, msg))
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

func (jt *JSONTimeDef) Check(je *JSONElem) error {
	var t time.Time
	var err error
	if jt.Format == "UNIX" {
		i, err := je.GetInt()
		if err != nil {
			return err
		}
		t = time.Unix(i, 0)
	} else {
		if jt.Format == "RFC3339" {
			jt.Format = time.RFC3339
		}
		t, err = time.Parse(jt.Format, je.GetString())
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
			je.Failf("Was %s ago, > %s", since, jt.Age)
		}
	}
	return nil
}

func (jt *JSONNumberDef) Check(je *JSONElem) error {
	val, err := je.GetFloat()
	if err != nil {
		return err
	}
	if jt.Min != nil && val < *jt.Min {
		je.Failf("%f < %f", val, *jt.Min)
	}
	if jt.Max != nil && val > *jt.Max {
		je.Failf("%f > %f", val, *jt.Max)
	}
	if jt.Eq != nil && val != *jt.Eq {
		je.Failf("%f != %f", val, *jt.Eq)
	}
	if jt.Neq != nil && val == *jt.Neq {
		je.Failf("%f == %f", val, *jt.Neq)
	}
	return nil
}

func (jt *JSONStringDef) Check(je *JSONElem) error {
	val, err := je.GetFloat()
	if err != nil {
		return err
	}
	if jt.Eq != nil && val != *jt.Eq {
		je.Failf("%f != %f", val, *jt.Eq)
	}
	if jt.Neq != nil && val == *jt.Neq {
		je.Failf("%f == %f", val, *jt.Neq)
	}
	return nil
}
func (t *JSONCheck) RunCheck(r *reporter.Reporter) error {
	rChild := r.Spawn("JSON Check %s", t.GetName())

	reader, err := t.Request.GetReader()
	if err != nil {
		return err
	}
	defer reader.Close()
	dec := json.NewDecoder(reader)
	data := map[string]interface{}{}
	dec.Decode(&data)
	ctx := &JSONContext{
		Data:     data,
		Reporter: rChild,
	}

	for _, test := range t.CheckFields {
		test.DoChecks(ctx, rChild)
	}

	return nil
}
