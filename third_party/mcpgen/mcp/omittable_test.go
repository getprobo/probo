package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOmittable_NotSet(t *testing.T) {
	var o Omittable[string]

	assert.False(t, o.IsSet())
	assert.False(t, o.IsNull())

	value, ok := o.Value()
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestOmittable_SetToValue(t *testing.T) {
	o := NewOmittable("hello")

	assert.True(t, o.IsSet())
	assert.False(t, o.IsNull())

	value, ok := o.Value()
	assert.True(t, ok)
	assert.Equal(t, "hello", value)
	assert.Equal(t, "hello", o.ValueOrZero())

	ptr := o.Ptr()
	require.NotNil(t, ptr)
	assert.Equal(t, "hello", *ptr)
}

func TestOmittable_SetToNull(t *testing.T) {
	o := NewOmittableNull[string]()

	assert.True(t, o.IsSet())
	assert.True(t, o.IsNull())

	value, ok := o.Value()
	assert.False(t, ok)
	assert.Equal(t, "", value)
	assert.Equal(t, "", o.ValueOrZero())
	assert.Nil(t, o.Ptr())
}

func TestOmittable_UnmarshalJSON_NotProvided(t *testing.T) {
	type Input struct {
		Name  Omittable[string] `json:"name,omitempty"`
		Email Omittable[string] `json:"email,omitempty"`
	}

	jsonData := `{"name": "John"}`
	var input Input
	require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

	assert.True(t, input.Name.IsSet())
	name, ok := input.Name.Value()
	assert.True(t, ok)
	assert.Equal(t, "John", name)

	assert.False(t, input.Email.IsSet())
}

func TestOmittable_UnmarshalJSON_ExplicitNull(t *testing.T) {
	type Input struct {
		Name  Omittable[string] `json:"name,omitempty"`
		Email Omittable[string] `json:"email,omitempty"`
	}

	jsonData := `{"name": null, "email": "test@example.com"}`
	var input Input
	require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

	assert.True(t, input.Name.IsSet())
	assert.True(t, input.Name.IsNull())

	assert.True(t, input.Email.IsSet())
	email, ok := input.Email.Value()
	assert.True(t, ok)
	assert.Equal(t, "test@example.com", email)
}

func TestOmittable_UnmarshalJSON_WithValue(t *testing.T) {
	type Input struct {
		Count Omittable[int] `json:"count,omitempty"`
	}

	jsonData := `{"count": 42}`
	var input Input
	require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

	assert.True(t, input.Count.IsSet())
	assert.False(t, input.Count.IsNull())
	count, ok := input.Count.Value()
	assert.True(t, ok)
	assert.Equal(t, 42, count)
}

func TestOmittable_MarshalJSON_NotSet(t *testing.T) {
	type Output struct {
		Name Omittable[string] `json:"name,omitempty"`
	}

	output := Output{}
	data, err := json.Marshal(output)
	require.NoError(t, err)

	assert.JSONEq(t, `{"name":null}`, string(data))
}

func TestOmittable_MarshalJSON_Null(t *testing.T) {
	type Output struct {
		Name Omittable[string] `json:"name,omitempty"`
	}

	output := Output{
		Name: NewOmittableNull[string](),
	}
	data, err := json.Marshal(output)
	require.NoError(t, err)

	assert.JSONEq(t, `{"name":null}`, string(data))
}

func TestOmittable_MarshalJSON_WithValue(t *testing.T) {
	type Output struct {
		Name Omittable[string] `json:"name,omitempty"`
	}

	output := Output{
		Name: NewOmittable("Alice"),
	}
	data, err := json.Marshal(output)
	require.NoError(t, err)

	assert.JSONEq(t, `{"name":"Alice"}`, string(data))
}

func TestOmittable_ComplexTypes(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Input struct {
		Person Omittable[Person] `json:"person,omitempty"`
	}

	t.Run("with value", func(t *testing.T) {
		jsonData := `{"person": {"name": "John", "age": 30}}`
		var input Input
		require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

		assert.True(t, input.Person.IsSet())
		person, ok := input.Person.Value()
		assert.True(t, ok)
		assert.Equal(t, "John", person.Name)
		assert.Equal(t, 30, person.Age)
	})

	t.Run("with null", func(t *testing.T) {
		jsonData := `{"person": null}`
		var input Input
		require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

		assert.True(t, input.Person.IsSet())
		assert.True(t, input.Person.IsNull())
	})
}

func TestOmittable_Pointers(t *testing.T) {
	type Input struct {
		Name Omittable[*string] `json:"name,omitempty"`
	}

	t.Run("with value", func(t *testing.T) {
		jsonData := `{"name": "hello"}`
		var input Input
		require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

		assert.True(t, input.Name.IsSet())
		value, ok := input.Name.Value()
		assert.True(t, ok)
		require.NotNil(t, value)
		assert.Equal(t, "hello", *value)
	})

	t.Run("with null", func(t *testing.T) {
		jsonData := `{"name": null}`
		var input Input
		require.NoError(t, json.Unmarshal([]byte(jsonData), &input))

		assert.True(t, input.Name.IsSet())
		assert.True(t, input.Name.IsNull())
	})
}
