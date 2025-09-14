package features

import (
	"testing"
)

func TestErrorHandling(t *testing.T) {
	ctx := NewBDDTestContext(t)
	defer ctx.Cleanup()

	t.Run("Handle resource not found", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenNoResourceExistsWithID("nonexistent-resource")
		ctx.whenITryToRetrieveNonexistentResource()
		ctx.thenIShouldReceiveStatusCode(404)
		ctx.thenTheErrorMessageShouldIndicate("not found")
	})

	t.Run("Handle unsupported format request", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveStoredRDFDataInFormat("Turtle")
		ctx.whenIRequestResourceWithAcceptHeader("application/unsupported")
		ctx.thenIShouldReceiveStatusCode(406)
		ctx.thenTheErrorMessageShouldIndicate("not acceptable")
	})

	t.Run("Handle invalid resource data", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()

		// Store invalid RDF data
		ctx.testData["rdfData"] = "invalid rdf data @#$%"
		ctx.testData["contentType"] = "text/turtle"
		ctx.whenIStoreTheResourceWithContentType("text/turtle")
		ctx.thenIShouldReceiveStatusCode(400)
		ctx.thenTheErrorMessageShouldIndicate("bad request")
	})
}
