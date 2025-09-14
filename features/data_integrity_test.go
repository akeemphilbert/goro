package features

import (
	"fmt"
	"testing"
)

func TestDataIntegrity(t *testing.T) {
	ctx := NewBDDTestContext(t)
	defer ctx.Cleanup()

	t.Run("Verify successful write operations", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveValidRDFDataInFormat("Turtle")
		ctx.whenIStoreTheResourceWithContentType("text/turtle")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Verify data is immediately retrievable
		ctx.whenIRequestResourceWithAcceptHeader("text/turtle")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Validate data integrity on retrieval", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()

		// Store multiple resources
		for i := 0; i < 5; i++ {
			ctx.givenIHaveValidRDFDataInFormat("Turtle")
			ctx.testData["resourceID"] = fmt.Sprintf("integrity-test-%d", i)
			ctx.whenIStoreTheResourceWithContentType("text/turtle")
			ctx.thenTheResourceShouldBeStoredSuccessfully()
		}

		// Retrieve each resource and verify integrity
		for i := 0; i < 5; i++ {
			ctx.testData["resourceID"] = fmt.Sprintf("integrity-test-%d", i)
			ctx.whenIRequestResourceWithAcceptHeader("text/turtle")
			ctx.thenIShouldReceiveStatusCode(200)
			ctx.thenTheSemanticMeaningShouldBePreserved()
		}
	})

	t.Run("Atomic operations for data consistency", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()

		// Store initial resource
		ctx.givenIHaveValidRDFDataInFormat("Turtle")
		ctx.whenIStoreTheResourceWithContentType("text/turtle")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Update the resource (should be atomic)
		ctx.givenIHaveValidRDFDataInFormat("JSON-LD")
		ctx.whenIStoreTheResourceWithContentType("application/ld+json")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Verify the update was successful and consistent
		ctx.whenIRequestResourceWithAcceptHeader("application/ld+json")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Binary file integrity verification", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveBinaryFile("integrity-test.jpg")
		ctx.whenIUploadFileWithContentType("image/jpeg")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Verify integrity through checksum
		ctx.whenIRequestResourceWithAcceptHeader("image/jpeg")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheChecksumShouldMatch()
		ctx.thenIShouldRetrieveExactOriginalContent()
	})
}
