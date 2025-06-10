# Template Engine

A high-performance, extensible template engine for generating dynamic data in Go. Supports fake data, regex, method chains, argument passing, and custom generators.

---

## Table of Contents
- [Overview](#overview)
- [Key Concepts](#key-concepts)
- [Supported Directives](#supported-directives)
- [Generators](#generators)
- [Methods](#methods)
- [Syntax & Examples](#syntax--examples)
---

## Overview

TemplateEngine is designed for fast, flexible, and modular template processing. It is used to generate test data, mock payloads, and dynamic content for HTTP/API testing and automation.

- **Fake data**: Generate names, emails, numbers, addresses, etc.
- **Regex**: Generate strings matching a regular expression.
- **Method chains**: Transform generated values (e.g. `.ToUpper()`, `.Replace()`)
- **Arguments**: Pass arguments to generators via dot or parentheses syntax.
- **Extensible**: Register your own generators and methods.

---

## Key Concepts

- **Generator**: A function that produces a value (e.g. `Fake.name`, `Regex`).
- **Method**: A function that transforms a value (e.g. `ToUpper`, `Replace`).
- **Directive**: A template expression inside `{{ ... }}`.
- **Arguments**: Values passed to generators or methods.

---

## Supported Directives

### Fake Data Generators
- `Fake.name`
- `Fake.email`
- `Fake.password`
- `Fake.int.min.max` (e.g. `Fake.int.1.10`)
- `Fake.uint.min.max` (e.g. `Fake.uint.10.100`)
- `Fake.float.min.max` (e.g. `Fake.float.0.1.10.5`)
- `Fake.bool`
- `Fake.phone`
- `Fake.address`
- `Fake.company`
- `Fake.date`
- `Fake.uuid`
- `Fake.url`
- `Fake.color`
- `Fake.word`
- `Fake.sentence`
- `Fake.country`
- `Fake.city`

### Regex Generator
`Regex(<pattern>)` — generates a string matching the given regex pattern
    (e.g, `^[a-z]{5,10}@example\\.com$`, `Regex(^[a-z]{5,10}@[a-z]{5,10}\\.(com|net|org)$\)` )

### Body Reference
- `Body.field` — reference to a value in the generated body (for nested templates).

---

## Methods
Methods can be chained to any generator result:
- `ToUpper()` - Formats the value to uppercase.
- `ToLower()` - Formats the value to lowercase.
- `TrimSpace()` - Removes leading and trailing whitespace.
- `Replace(old, new)` - Replaces occurrences of `old` with `new`.
- `PadLeft(width, char)` - Pads the value to the left with the specified character.
- `PadRight(width, char)` - Pads the value to the right with the specified character.
- `Substring(start, length)` - Extracts a substring from the value.
- `Capitalize()` - Capitalizes the first letter of the value.
- `Reverse()` - Reverses the value.
- `RandomCase()` - Randomly capitalizes the first letter of the value.
- `SnakeCase()` - Converts the value to snake case.
- `CamelCase()` - Converts the value to camel case.
- `Split(sep)` - Splits the value by the specified separator.
- `Join(sep)` - Joins the values with the specified separator.
- `Index(idx)` - Returns the value at the specified index.
- `Cut(start, end)` - Extract the specified range from the value.
- `ToString()` - Converts the value to a string.
- `ToInt()` - Convert the value to an integer.
- `ToUint()` - Convert the value to an unsigned integer.
- `ToFloat()` - Convert the value to a float.
- `ToBool()` - Convert the value to a boolean.
- `ToArray()` - Converts the value to an array.

All method does not fail if the value does not match the expected type of the method has not valid arguments.
In case of an error, the method returns the original value.

---

## Syntax & Examples

### Basic Usage
```
{{ Fake.name }}
{{ Fake.email }}
{{ Fake.int.1.100 }}
{{ Regex(\"^[a-z]{5,10}@example\\.com$\") }}
```

### Method Chains
```
{{ Fake.name.ToUpper() }}
{{ Fake.email.Replace('@', '_at_') }}
{{ Fake.word.ToString().PadLeft(10, '-') }}
```

### Arguments
```
{{ Fake.uint.10.100 }}
{{ Fake.float.1.5.2.4 }}
{{ Regex('foo[0-9]{3}') }}
```

### Nested Templates
```
{
  "user": {
    "name": "{{ Fake.name }}",
    "email": "{{ Fake.email.ToLower() }}",
    "age": "{{ Fake.uint.18.99 }}"
  }
}
```
