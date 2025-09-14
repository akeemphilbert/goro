package features

import (
	"testing"
)

func TestRDFFormatSupport(t *testing.T) {
	ctx := NewBDDTestContext(t)
	defer ctx.Cleanup()

	t.Run("Store and retrieve RDF data in Turtle format", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveValidRDFDataInFormat("Turtle")
		ctx.whenIStoreTheResourceWithContentType("text/turtle")
		ctx.thenTheResourceShouldBeStoredSuccessfully()
		ctx.thenIShouldBeAbleToRetrieveItInFormat("Turtle")
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Store and retrieve RDF data in JSON-LD format", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveValidRDFDataInFormat("JSON-LD")
		ctx.whenIStoreTheResourceWithContentType("application/ld+json")
		ctx.thenTheResourceShouldBeStoredSuccessfully()
		ctx.thenIShouldBeAbleToRetrieveItInFormat("JSON-LD")
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Store and retrieve RDF data in RDF/XML format", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveValidRDFDataInFormat("RDF/XML")
		ctx.whenIStoreTheResourceWithContentType("application/rdf+xml")
		ctx.thenTheResourceShouldBeStoredSuccessfully()
		ctx.thenIShouldBeAbleToRetrieveItInFormat("RDF/XML")
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Content negotiation for RDF formats", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveStoredRDFDataInFormat("Turtle")
		ctx.whenIRequestResourceWithAcceptHeader("application/ld+json")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Convert between all supported RDF formats", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveStoredRDFDataInFormat("JSON-LD")

		// Test conversion to Turtle
		ctx.whenIRequestResourceWithAcceptHeader("text/turtle")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheSemanticMeaningShouldBePreserved()

		// Test conversion to RDF/XML
		ctx.whenIRequestResourceWithAcceptHeader("application/rdf+xml")
		ctx.thenIShouldReceiveStatusCode(200)
		ctx.thenTheSemanticMeaningShouldBePreserved()
	})

	t.Run("Reject unsupported RDF format", func(t *testing.T) {
		ctx.givenTheStorageSystemIsRunning()
		ctx.givenThePodStorageIsAvailable()
		ctx.givenIHaveStoredRDFDataInFormat("Turtle")
		ctx.whenIRequestResourceWithAcceptHeader("application/n-triples")
		ctx.thenIShouldReceiveStatusCode(406)
		ctx.thenTheErrorMessageShouldIndicate("unsupported format")
	})
}
