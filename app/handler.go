package main

import (
	"sync"
)

var Handlers = map[string]func([]Value) Value{
	"COMMAND": command,
	"PING":    ping,
	"GET":     get,
	"SET":     set,
	"HGET":    hget,
	"HSET":    hset,
	"HGETALL": hgetall,
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func command(args []Value) Value {
	return Value{typ: "null"}
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}
	return Value{typ: "string", str: args[0].bulk}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key, value := args[0].bulk, args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash, key, value := args[0].bulk, args[1].bulk, args[2].bulk

	HSETsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	hash, key := args[0].bulk, args[1].bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hgetall(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	hash := args[0].bulk

	HSETsMu.RLock()
	keys, ok := HSETs[hash]
	HSETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	values := []Value{}
	for _, val := range keys {
		values = append(values, Value{typ: "string", str: val})
	}

	return Value{typ: "array", array: values}
}
