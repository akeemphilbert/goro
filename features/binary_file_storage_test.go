package features

import (
	"testing"
)

func TestBinaryFileStorage(t *testing.T) {
	ctx := NewBDDTestContext(t)
	defer ctx.Cleanup()

	t.Run("Store and retrieve a binary image file", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveBinaryFile("test.jpg")
		ctx.whenIUploadFileWithContentType("image/jpeg")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Retrieve and verify
		ctx.whenIRequestResourceWithAcceptHeader("image/jpeg")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenIShouldRetrieveExactOriginalContent()
		ctx.thenTheMIMETypeShouldBePreserved("image/jpeg")
	})

	t.Run("Store and retrieve a binary document file", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveBinaryFile("document.pdf")
		ctx.whenIUploadFileWithContentType("application/pdf")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Retrieve and verify
		ctx.whenIRequestResourceWithAcceptHeader("application/pdf")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenIShouldRetrieveExactOriginalContent()
		ctx.thenTheMIMETypeShouldBePreserved("application/pdf")
	})

	t.Run("Store large binary file with streaming", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveLargeBinaryFile(10) // 10MB
		ctx.whenIUploadFileWithContentType("application/octet-stream")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Retrieve and verify
		ctx.whenIRequestResourceWithAcceptHeader("application/octet-stream")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenIShouldRetrieveExactOriginalContent()
	})

	t.Run("Verify binary file integrity", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveBinaryFile("test.jpg")
		ctx.whenIUploadFileWithContentType("image/jpeg")
		ctx.thenTheResourceShouldBeStoredSuccessfully()

		// Retrieve and verify checksum
		ctx.whenIRequestResourceWithAcceptHeader("image/jpeg")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheChecksumShouldMatch()
	})
}
