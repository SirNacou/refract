package idgen_test

import (
	"fmt"
	"log"
	"os"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
)

// Example_basicUsage demonstrates basic Snowflake ID generation
func Example_basicUsage() {
	// Create generator with explicit worker ID
	gen, err := idgen.NewSnowflakeGeneratorWithWorkerID(1)
	if err != nil {
		log.Fatal(err)
	}

	// Generate a unique ID
	id, err := gen.NextID()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated ID: %d\n", id)
	fmt.Printf("Worker ID: %d\n", gen.WorkerID())
}

// Example_fromEnvironment demonstrates creating generator from WORKER_ID env var
func Example_fromEnvironment() {
	// Set worker ID via environment variable
	os.Setenv("WORKER_ID", "42")
	defer os.Unsetenv("WORKER_ID")

	// Create generator (reads from WORKER_ID env var)
	gen, err := idgen.NewSnowflakeGenerator()
	if err != nil {
		log.Fatal(err)
	}

	// Generate IDs
	id1, _ := gen.NextID()
	id2, _ := gen.NextID()

	fmt.Printf("Worker ID from env: %d\n", gen.WorkerID())
	fmt.Printf("Generated IDs are unique: %t\n", id1 != id2)
}

// Example_multipleWorkers demonstrates using different worker IDs for distributed systems
func Example_multipleWorkers() {
	// Simulate multiple API service instances with different worker IDs
	worker1, _ := idgen.NewSnowflakeGeneratorWithWorkerID(0)  // API instance 1
	worker2, _ := idgen.NewSnowflakeGeneratorWithWorkerID(1)  // API instance 2
	worker3, _ := idgen.NewSnowflakeGeneratorWithWorkerID(63) // API instance 64

	id1, _ := worker1.NextID()
	id2, _ := worker2.NextID()
	id3, _ := worker3.NextID()

	fmt.Printf("All IDs are unique: %t\n", id1 != id2 && id2 != id3 && id1 != id3)
}

// Example_interfaceUsage demonstrates using the IDGenerator interface
func Example_interfaceUsage() {
	// Function that accepts any ID generator
	generateMultipleIDs := func(gen idgen.IDGenerator, count int) []uint64 {
		ids := make([]uint64, count)
		for i := 0; i < count; i++ {
			id, _ := gen.NextID()
			ids[i] = id
		}
		return ids
	}

	// Use Snowflake generator via interface
	gen, _ := idgen.NewSnowflakeGeneratorWithWorkerID(10)
	ids := generateMultipleIDs(gen, 5)

	fmt.Printf("Generated %d IDs via interface\n", len(ids))
}

// Example_base62Encoding demonstrates encoding Snowflake IDs to Base62 for URLs
func Example_base62Encoding() {
	// Generate a Snowflake ID
	gen, _ := idgen.NewSnowflakeGeneratorWithWorkerID(0)
	id, _ := gen.NextID()

	// Encode to Base62 for URL-safe short code
	shortCode := idgen.EncodeBase62(id)

	fmt.Printf("Snowflake ID: %d\n", id)
	fmt.Printf("Base62 code: %s\n", shortCode)
	fmt.Printf("Code length: %d characters\n", len(shortCode))

	// Decode back to verify
	decoded, _ := idgen.DecodeBase62(shortCode)
	fmt.Printf("Roundtrip successful: %t\n", decoded == id)
}

// Example_urlShortener demonstrates a complete URL shortener workflow
func Example_urlShortener() {
	// Setup: Create generator
	gen, _ := idgen.NewSnowflakeGeneratorWithWorkerID(0)

	// User submits a long URL
	longURL := "https://example.com/very/long/path?param1=value1&param2=value2"

	// Generate unique ID and create short code
	id, _ := gen.NextID()
	shortCode := idgen.EncodeBase62(id)
	shortURL := fmt.Sprintf("https://short.link/%s", shortCode)

	fmt.Printf("Original: %s\n", longURL)
	fmt.Printf("Shortened: %s\n", shortURL)
	fmt.Printf("Saved %d characters!\n", len(longURL)-len(shortURL))

	// When user visits short URL, decode and redirect
	decodedID, _ := idgen.DecodeBase62(shortCode)
	fmt.Printf("\nDecoded ID: %d (use to lookup in database)\n", decodedID)
}

// Example_base62Properties demonstrates Base62 encoding properties
func Example_base62Properties() {
	// Encode various numbers
	examples := []uint64{
		0,
		123,
		1234567890,
		1234567890123456,
	}

	fmt.Println("Base62 Encoding Examples:")
	for _, num := range examples {
		encoded := idgen.EncodeBase62(num)
		fmt.Printf("%d -> %s\n", num, encoded)
	}

	// Show URL-safe alphabet
	fmt.Println("\nBase62 alphabet: 0-9, a-z, A-Z (62 chars total)")
	fmt.Println("URL-safe: ✓ No special characters")
}
