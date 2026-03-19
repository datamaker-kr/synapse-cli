package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
)

type tableFormatter struct{ w io.Writer }

func (f *tableFormatter) Format(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		_, printErr := fmt.Fprintln(f.w, string(b))
		return printErr
	}

	t := table.NewWriter()
	t.SetOutputMirror(f.w)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"Field", "Value"})
	for k, v := range m {
		t.AppendRow(table.Row{k, fmt.Sprintf("%v", v)})
	}
	t.Render()
	return nil
}

func (f *tableFormatter) FormatList(items interface{}, columns []Column) error {
	val := reflect.ValueOf(items)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice {
		return f.Format(items)
	}

	t := table.NewWriter()
	t.SetOutputMirror(f.w)
	t.SetStyle(table.StyleLight)

	header := make(table.Row, len(columns))
	for i, col := range columns {
		header[i] = col.Header
	}
	t.AppendHeader(header)

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		b, err := json.Marshal(item)
		if err != nil {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			continue
		}
		row := make(table.Row, len(columns))
		for j, col := range columns {
			if v, ok := m[col.Field]; ok {
				row[j] = fmt.Sprintf("%v", v)
			}
		}
		t.AppendRow(row)
	}

	t.Render()
	return nil
}

func (f *tableFormatter) FormatError(statusCode int, body []byte) error {
	_, err := fmt.Fprintf(f.w, "Error (HTTP %d): %s\n", statusCode, string(body))
	return err
}
