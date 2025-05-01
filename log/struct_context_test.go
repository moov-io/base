package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStructContext(t *testing.T) {
	type Address struct {
		Street  string `log:"street"`
		City    string `log:"city"`
		Country string `log:"country"`
		ZipCode string `log:"zip_code,omitempty"`
	}

	type Person struct {
		Name      string    `log:"name"`
		Age       int       `log:"age"`
		Email     string    `log:"email,omitempty"`
		CreatedAt time.Time `log:"created_at"`
		Address   Address   `log:"address"`
		Hidden    string
	}

	now := time.Now()
	p := Person{
		Name:      "John Doe",
		Age:       30,
		Email:     "john@example.com",
		CreatedAt: now,
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
		Hidden: "should not appear",
	}

	// Test basic struct context
	ctx := StructContext(p)
	fields := ctx.Context()

	require.Equal(t, 8, len(fields))
	require.Equal(t, "John Doe", fields["name"].getValue())
	require.Equal(t, int64(30), fields["age"].getValue())
	require.Equal(t, "john@example.com", fields["email"].getValue())
	require.Equal(t, now.Format(time.RFC3339Nano), fields["created_at"].getValue())
	require.Equal(t, "123 Main St", fields["address.street"].getValue())
	require.Equal(t, "New York", fields["address.city"].getValue())
	require.Equal(t, "USA", fields["address.country"].getValue())
	require.Contains(t, fields, "address") // The struct itself is also included
	require.NotContains(t, fields, "Hidden")
	require.NotContains(t, fields, "address.zip_code") // Should be omitted as it's empty

	// Test with prefix
	ctx = StructContext(p, WithPrefix("user"))
	fields = ctx.Context()

	require.Equal(t, 8, len(fields))
	require.Equal(t, "John Doe", fields["user.name"].getValue())
	require.Equal(t, int64(30), fields["user.age"].getValue())
	require.Equal(t, "john@example.com", fields["user.email"].getValue())
	require.Equal(t, now.Format(time.RFC3339Nano), fields["user.created_at"].getValue())
	require.Equal(t, "123 Main St", fields["user.address.street"].getValue())
	require.Equal(t, "New York", fields["user.address.city"].getValue())
	require.Equal(t, "USA", fields["user.address.country"].getValue())
	require.Contains(t, fields, "user.address") // The struct itself is also included

	// Test with nil value
	ctx = StructContext(nil)
	require.Empty(t, ctx.Context())

	// Test omitempty behavior
	p.Email = ""
	ctx = StructContext(p)
	fields = ctx.Context()
	require.NotContains(t, fields, "email") // Should be omitted as it's empty

	// Test with pointer to struct
	ctx = StructContext(&p)
	fields = ctx.Context()
	require.Equal(t, 7, len(fields)) // email is empty and omitted
	require.Equal(t, "John Doe", fields["name"].getValue())

	// Test nested pointer structs
	type Department struct {
		Name string `log:"name"`
	}

	type Company struct {
		Dept *Department `log:"department"`
	}

	type Employee struct {
		Company *Company `log:"company"`
	}

	employee := Employee{
		Company: &Company{
			Dept: &Department{
				Name: "Engineering",
			},
		},
	}

	ctx = StructContext(employee)
	fields = ctx.Context()
	require.Equal(t, 3, len(fields))
	require.Contains(t, fields, "company")
	require.Contains(t, fields, "company.department")
	require.Equal(t, "Engineering", fields["company.department.name"].getValue())

	// Test struct without log tag is not included
	type TeamMember struct {
		Role     string   `log:"role"`
		Employee Employee // No log tag
	}

	team := TeamMember{
		Role: "Developer",
		Employee: Employee{
			Company: &Company{
				Dept: &Department{
					Name: "Engineering",
				},
			},
		},
	}

	ctx = StructContext(team)
	fields = ctx.Context()
	require.Equal(t, 1, len(fields))
	require.Equal(t, "Developer", fields["role"].getValue())
	require.NotContains(t, fields, "employee.company.department.name")

	// Test with various value types
	type AllTypes struct {
		BoolVal    bool    `log:"bool"`
		IntVal     int     `log:"int"`
		Int64Val   int64   `log:"int64"`
		UintVal    uint    `log:"uint"`
		Uint64Val  uint64  `log:"uint64"`
		Float32Val float32 `log:"float32"`
		Float64Val float64 `log:"float64"`
		StringVal  string  `log:"string"`
	}

	allTypes := AllTypes{
		BoolVal:    true,
		IntVal:     42,
		Int64Val:   int64(9223372036854775807),
		UintVal:    42,
		Uint64Val:  uint64(18446744073709551615),
		Float32Val: 3.14,
		Float64Val: 2.71828,
		StringVal:  "hello",
	}

	ctx = StructContext(allTypes)
	fields = ctx.Context()
	require.Equal(t, 8, len(fields))
	require.Equal(t, true, fields["bool"].getValue())
	require.Equal(t, int64(42), fields["int"].getValue())
	require.Equal(t, int64(9223372036854775807), fields["int64"].getValue())
	require.Equal(t, uint64(42), fields["uint"].getValue())
	require.Equal(t, uint64(18446744073709551615), fields["uint64"].getValue())
	require.Equal(t, float32(3.14), fields["float32"].getValue())
	require.Equal(t, float64(2.71828), fields["float64"].getValue())
	require.Equal(t, "hello", fields["string"].getValue())
}

func TestStructContextWithTag(t *testing.T) {
	// Define a struct with otel tags instead of log tags
	type Product struct {
		ID          int       `otel:"product_id"`
		Name        string    `otel:"product_name"`
		Price       float64   `otel:"price"`
		Description string    `otel:"description,omitempty"`
		CreatedAt   time.Time `otel:"created_at"`
	}

	now := time.Now()
	product := Product{
		ID:        123,
		Name:      "Test Product",
		Price:     29.99,
		CreatedAt: now,
	}

	// Use StructContext with WithTag option to use otel tags instead of log tags
	ctx := StructContext(product, WithTag("otel"))
	fields := ctx.Context()

	// Verify that the fields are extracted using otel tags
	require.Contains(t, fields, "product_id")
	require.Contains(t, fields, "product_name")
	require.Contains(t, fields, "price")
	require.Contains(t, fields, "created_at")
	require.NotContains(t, fields, "description") // Should be omitted as it's empty with omitempty

	// Verify values
	require.Equal(t, int64(123), fields["product_id"].getValue())
	require.Equal(t, "Test Product", fields["product_name"].getValue())
	require.Equal(t, float64(29.99), fields["price"].getValue())

	// Test with both custom tag and prefix
	ctx = StructContext(product, WithTag("otel"), WithPrefix("item"))
	fields = ctx.Context()

	require.Contains(t, fields, "item.product_id")
	require.Contains(t, fields, "item.product_name")
	require.Contains(t, fields, "item.price")
	require.Contains(t, fields, "item.created_at")
	require.NotContains(t, fields, "item.description")
}

func TestStructContextWithLogger(t *testing.T) {
	type User struct {
		ID       int    `log:"id"`
		Username string `log:"username"`
		Email    string `log:"email,omitempty"`
	}

	buffer, logger := NewBufferLogger()
	user := User{
		ID:       1,
		Username: "johndoe",
		Email:    "john@example.com",
	}

	// Log with struct context
	logger.With(StructContext(user)).Info().Log("User logged in")

	// Check log output
	output := buffer.String()
	require.Contains(t, output, "username=johndoe")
	require.Contains(t, output, "id=1")
	require.Contains(t, output, "email=john@example.com")
	require.Contains(t, output, "level=info")
	require.Contains(t, output, "msg=\"User logged in\"")

	// Test with prefix
	buffer.Reset()
	logger.With(StructContext(user, WithPrefix("user"))).Info().Log("User details")

	output = buffer.String()
	require.Contains(t, output, "user.username=johndoe")
	require.Contains(t, output, "user.id=1")
	require.Contains(t, output, "user.email=john@example.com")
}
