package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	capi "github.com/hashicorp/consul/api"
)

// Config consul 地址配置
type Config struct {
	Address string
}

// DefaultConfig ...
func DefaultConfig() *Config {
	c := &Config{}
	c.Address = "127.0.0.1:8500"
	return c
}

// Gosh ...
type Gosh struct {
	client *capi.Client
}

// NewClient ...
func NewClient(c *Config) (*Gosh, error) {
	config := capi.DefaultConfig()
	config.Address = c.Address
	consul, err := capi.NewClient(config)
	if err != nil {
		return nil, err
	}

	g := &Gosh{}
	g.client = consul
	return g, nil
}

// Unmarshal ...
func (g *Gosh) Unmarshal(v interface{}) error {
	var ops capi.KVTxnOps

	rv := reflect.ValueOf(v).Elem()
	if !rv.IsValid() {
		return errors.New("参数未初始化")
	}

	rt := reflect.TypeOf(v).Elem()

	ndx := 0
	dt := make(map[int]string)

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		key := field.Tag.Get("consul")
		if key != "" {
			op := capi.KVTxnOp{
				Verb: capi.KVGet,
				Key:  key,
			}
			ops = append(ops, &op)
			dt[ndx] = field.Name
			ndx++
		}
	}

	ok, res, _, err := g.client.KV().Txn(ops, nil)
	if err != nil {
		return err
	}

	if !ok {
		str := ""
		for i := 0; i < len(res.Errors); i++ {
			str += fmt.Sprintf("%v\n", res.Errors[i].What)
		}
		return errors.New(str)
	}

	obj := make(map[string]interface{})

	for i, each := range res.Results {
		key := dt[i]
		switch f := rv.FieldByName(key); f.Kind() {
		case reflect.String:
			obj[key] = string(each.Value)

		case
			reflect.Interface,
			reflect.Map,
			reflect.Slice,
			reflect.Ptr,
			reflect.Struct,
			reflect.Bool:
			var x interface{}
			err := json.Unmarshal(each.Value, &x)
			if err != nil {
				return err
			}
			obj[key] = x

		case
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			n, err := strconv.ParseInt(string(each.Value), 10, 64)
			if err != nil || f.OverflowInt(n) {
				return err
			}
			obj[key] = n

		case
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr:
			n, err := strconv.ParseUint(string(each.Value), 10, 64)
			if err != nil || f.OverflowUint(n) {
				return err
			}

			obj[key] = n

		case
			reflect.Float32,
			reflect.Float64:
			n, err := strconv.ParseFloat(string(each.Value), f.Type().Bits())
			if err != nil || f.OverflowFloat(n) {
				return err
			}

			obj[key] = n

		default:
			return fmt.Errorf("not support %v", f.Type())
		}

	}

	b, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil
}
