package features

import (
	"testing"
)

// TestContainerCreationAndHierarchyManagement tests BDD scenarios for container creation and hierarchy
func TestContainerCreationAndHierarchyManagement(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("CreateBasicContainer", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		// And the server supports container operations
		ctx.givenTheServerSupportsContainerOperations()

		// When I create a container with ID "documents"
		ctx.whenICreateAContainerWithID("documents")

		// Then the container should be created successfully
		ctx.thenTheContainerShouldBeCreatedSuccessfully()
		// And the container should have type "BasicContainer"
		ctx.thenTheContainerShouldHaveType("BasicContainer")
		// And the container should be empty
		ctx.thenTheContainerShouldBeEmpty()
	})

	t.Run("CreateNestedContainerHierarchy", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I create a container "images" inside "documents"
		ctx.whenICreateAContainerInside("images", "documents")

		// Then the container "images" should be created successfully
		ctx.thenTheContainerShouldBeCreatedSuccessfully()
		// And the container "images" should have parent "documents"
		ctx.thenTheContainerShouldHaveParent("images", "documents")
		// And the container "documents" should contain "images"
		ctx.thenTheContainerShouldContain("documents", "images")
	})

	t.Run("CreateDeepHierarchy", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "root" exists
		ctx.givenAContainerExists("root")
		// And a container "level1" exists inside "root"
		ctx.givenAContainerExistsInside("level1", "root")

		// When I create a container "level2" inside "level1"
		ctx.whenICreateAContainerInside("level2", "level1")

		// Then the container "level2" should be created successfully
		ctx.thenTheContainerShouldBeCreatedSuccessfully()
		// And the hierarchy path should be "root/level1/level2"
		ctx.thenTheHierarchyPathShouldBe("root/level1/level2")
	})

	t.Run("PreventCircularReferences", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "parent" exists
		ctx.givenAContainerExists("parent")
		// And a container "child" exists inside "parent"
		ctx.givenAContainerExistsInside("child", "parent")

		// When I try to move container "parent" inside "child"
		ctx.whenITryToMoveContainerInside("parent", "child")

		// Then the operation should fail with "circular reference" error
		ctx.thenTheOperationShouldFailWithError("circular reference")
	})

	t.Run("ContainerCreationWithMetadata", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()

		// When I create a container with ID "photos" and title "Photo Collection"
		ctx.whenICreateAContainerWithIDAndTitle("photos", "Photo Collection")

		// Then the container should be created successfully
		ctx.thenTheContainerShouldBeCreatedSuccessfully()
		// And the container should have title "Photo Collection"
		ctx.thenTheContainerShouldHaveTitle("Photo Collection")
		// And the container should have creation timestamp
		ctx.thenTheContainerShouldHaveCreationTimestamp()
		// And the container should have Dublin Core metadata
		ctx.thenTheContainerShouldHaveDublinCoreMetadata()
	})

	t.Run("ContainerCreationValidation", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()

		// When I try to create a container with invalid ID ""
		ctx.whenITryToCreateAContainerWithInvalidID("")

		// Then the operation should fail with "invalid container ID" error
		ctx.thenTheOperationShouldFailWithError("invalid container ID")
	})

	t.Run("DuplicateContainerPrevention", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I try to create another container with ID "documents"
		ctx.whenITryToCreateAnotherContainerWithID("documents")

		// Then the operation should fail with "container already exists" error
		ctx.thenTheOperationShouldFailWithError("container already exists")
	})
}

// TestContainerMembershipOperations tests BDD scenarios for container membership
func TestContainerMembershipOperations(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("AddResourceToContainer", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And a resource "doc1.ttl" exists
		ctx.givenAResourceExists("doc1.ttl")

		// When I add resource "doc1.ttl" to container "documents"
		ctx.whenIAddResourceToContainer("doc1.ttl", "documents")

		// Then the container "documents" should contain resource "doc1.ttl"
		ctx.thenTheContainerShouldContainResource("documents", "doc1.ttl")
		// And the membership should be recorded in the index
		ctx.thenTheMembershipShouldBeRecordedInTheIndex()
		// And the container should emit "member_added" event
		ctx.thenTheContainerShouldEmitEvent("member_added")
	})

	t.Run("RemoveResourceFromContainer", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And a resource "doc1.ttl" exists in container "documents"
		ctx.givenAResourceExistsInContainer("doc1.ttl", "documents")

		// When I remove resource "doc1.ttl" from container "documents"
		ctx.whenIRemoveResourceFromContainer("doc1.ttl", "documents")

		// Then the container "documents" should not contain resource "doc1.ttl"
		ctx.thenTheContainerShouldNotContainResource("documents", "doc1.ttl")
		// And the membership should be removed from the index
		ctx.thenTheMembershipShouldBeRemovedFromTheIndex()
		// And the container should emit "member_removed" event
		ctx.thenTheContainerShouldEmitEvent("member_removed")
	})

	t.Run("ListContainerMembers", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And resources "doc1.ttl", "doc2.ttl", "doc3.ttl" exist in container "documents"
		ctx.givenResourcesExistInContainer("doc1.ttl, doc2.ttl, doc3.ttl", "documents")

		// When I list the members of container "documents"
		ctx.whenIListTheMembersOfContainer("documents")

		// Then I should get 3 members
		ctx.thenIShouldGetMembers(3)
		// And the members should include "doc1.ttl", "doc2.ttl", "doc3.ttl"
		ctx.thenTheMembersShouldInclude("doc1.ttl, doc2.ttl, doc3.ttl")
		// And each member should have type information
		ctx.thenEachMemberShouldHaveTypeInformation()
	})

	t.Run("ContainerWithMixedContentTypes", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "mixed" exists
		ctx.givenAContainerExists("mixed")
		// And a RDF resource "data.ttl" exists in container "mixed"
		ctx.givenAResourceExistsInContainer("data.ttl", "mixed")
		// And a binary file "image.jpg" exists in container "mixed"
		ctx.givenAResourceExistsInContainer("image.jpg", "mixed")
		// And a sub-container "subfolder" exists in container "mixed"
		ctx.givenAContainerExistsInside("subfolder", "mixed")

		// When I list the members of container "mixed"
		ctx.whenIListTheMembersOfContainer("mixed")

		// Then I should get 3 members
		ctx.thenIShouldGetMembers(3)
		// And the members should have correct type information
		ctx.thenTheMembersShouldHaveCorrectTypeInformation()
		// And RDF resources should be marked as "Resource"
		ctx.thenRDFResourcesShouldBeMarkedAs("Resource")
		// And binary files should be marked as "NonRDFSource"
		ctx.thenBinaryFilesShouldBeMarkedAs("NonRDFSource")
		// And containers should be marked as "Container"
		ctx.thenContainersShouldBeMarkedAs("Container")
	})

	t.Run("AutomaticMembershipUpdatesOnResourceCreation", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I create a resource "new-doc.ttl" in container "documents"
		ctx.whenICreateAResourceInContainer("new-doc.ttl", "documents")

		// Then the container "documents" should automatically contain "new-doc.ttl"
		ctx.thenTheContainerShouldAutomaticallyContain("documents", "new-doc.ttl")
		// And the membership index should be updated
		ctx.thenTheMembershipIndexShouldBeUpdated()
		// And the container modification timestamp should be updated
		ctx.thenTheContainerModificationTimestampShouldBeUpdated()
	})

	t.Run("AutomaticMembershipCleanupOnResourceDeletion", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And a resource "temp-doc.ttl" exists in container "documents"
		ctx.givenAResourceExistsInContainer("temp-doc.ttl", "documents")

		// When I delete resource "temp-doc.ttl"
		ctx.whenIDeleteResource("temp-doc.ttl")

		// Then the container "documents" should not contain "temp-doc.ttl"
		ctx.thenTheContainerShouldNotContainResource("documents", "temp-doc.ttl")
		// And the membership index should be cleaned up
		ctx.thenTheMembershipIndexShouldBeCleanedUp()
		// And the container modification timestamp should be updated
		ctx.thenTheContainerModificationTimestampShouldBeUpdated()
	})

	t.Run("LDPMembershipTriplesGeneration", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And resources "doc1.ttl", "doc2.ttl" exist in container "documents"
		ctx.givenResourcesExistInContainer("doc1.ttl, doc2.ttl", "documents")

		// When I retrieve container "documents" as Turtle
		ctx.whenIRetrieveContainerAsFormat("documents", "Turtle")

		// Then the response should contain LDP membership triples
		ctx.thenTheResponseShouldContainLDPMembershipTriples()
		// And the triples should use "ldp:contains" predicate
		ctx.thenTheTriplesShouldUsePredicate("ldp:contains")
		// And each member should be properly referenced
		ctx.thenEachMemberShouldBeProperlyReferenced()
	})
}

// TestContainerHTTPAPICompliance tests BDD scenarios for HTTP API compliance
func TestContainerHTTPAPICompliance(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("GETContainerWithMembers", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And resources "doc1.ttl", "doc2.ttl" exist in container "documents"
		ctx.givenResourcesExistInContainer("doc1.ttl, doc2.ttl", "documents")

		// When I send GET request to "/containers/documents"
		ctx.whenISendRequestToWithAccept("GET", "/containers/documents", "")

		// Then the response status should be 200
		ctx.thenTheResponseStatusShouldBe(200)
		// And the response should contain container metadata
		ctx.thenTheResponseShouldContainContainerMetadata()
		// And the response should list all members
		ctx.thenTheResponseShouldListAllMembers()
		// And the Content-Type should be "text/turtle"
		ctx.thenTheContentTypeShouldBe("text/turtle")
	})

	t.Run("GETContainerWithContentNegotiation", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And resources "doc1.ttl", "doc2.ttl" exist in container "documents"
		ctx.givenResourcesExistInContainer("doc1.ttl, doc2.ttl", "documents")

		// When I send GET request to "/containers/documents" with Accept "application/ld+json"
		ctx.whenISendRequestToWithAccept("GET", "/containers/documents", "application/ld+json")

		// Then the response status should be 200
		ctx.thenTheResponseStatusShouldBe(200)
		// And the Content-Type should be "application/ld+json"
		ctx.thenTheContentTypeShouldBe("application/ld+json")
		// And the response should be valid JSON-LD
		ctx.thenTheResponseShouldBeValidJSONLD()
	})

	t.Run("POSTToCreateResourceInContainer", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I send POST request to "/containers/documents" with RDF content
		ctx.whenISendRequestToWithRDFContent("POST", "/containers/documents")

		// Then the response status should be 201
		ctx.thenTheResponseStatusShouldBe(201)
		// And the Location header should contain the new resource URI
		ctx.thenTheLocationHeaderShouldContainTheNewResourceURI()
		// And the resource should be added to the container
		ctx.thenTheResourceShouldBeAddedToTheContainer()
		// And the container should be updated
		ctx.thenTheContainerShouldBeUpdated()
	})

	t.Run("PUTToUpdateContainerMetadata", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I send PUT request to "/containers/documents" with updated metadata
		ctx.whenISendRequestToWithUpdatedMetadata("PUT", "/containers/documents")

		// Then the response status should be 200
		ctx.thenTheResponseStatusShouldBe(200)
		// And the container metadata should be updated
		ctx.thenTheContainerMetadataShouldBeUpdated()
		// And the modification timestamp should be updated
		ctx.thenTheModificationTimestampShouldBeUpdated()
	})

	t.Run("DELETEEmptyContainer", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And an empty container "temp" exists
		ctx.givenAnEmptyContainerExists("temp")

		// When I send DELETE request to "/containers/temp"
		ctx.whenISendRequestToWithAccept("DELETE", "/containers/temp", "")

		// Then the response status should be 204
		ctx.thenTheResponseStatusShouldBe(204)
		// And the container should be deleted
		ctx.thenTheContainerShouldBeDeleted()
		// And the container should emit "container_deleted" event
		ctx.thenTheContainerShouldEmitEvent("container_deleted")
	})

	t.Run("DELETENonEmptyContainerFails", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And a resource "doc1.ttl" exists in container "documents"
		ctx.givenAResourceExistsInContainer("doc1.ttl", "documents")

		// When I send DELETE request to "/containers/documents"
		ctx.whenISendRequestToWithAccept("DELETE", "/containers/documents", "")

		// Then the response status should be 409
		ctx.thenTheResponseStatusShouldBe(409)
		// And the response should contain "container not empty" error
		ctx.thenTheResponseShouldContainError("container not empty")
		// And the container should not be deleted
		ctx.thenTheContainerShouldNotBeDeleted()
	})

	t.Run("HEADRequestForContainerMetadata", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I send HEAD request to "/containers/documents"
		ctx.whenISendRequestToWithAccept("HEAD", "/containers/documents", "")

		// Then the response status should be 200
		ctx.thenTheResponseStatusShouldBe(200)
		// And the response should have no body
		ctx.thenTheResponseShouldHaveNoBody()
		// And the headers should contain container metadata
		ctx.thenTheHeadersShouldContainContainerMetadata()
		// And the Content-Length header should be present
		ctx.thenTheContentLengthHeaderShouldBePresent()
	})

	t.Run("OPTIONSRequestForSupportedMethods", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I send OPTIONS request to "/containers/documents"
		ctx.whenISendRequestToWithAccept("OPTIONS", "/containers/documents", "")

		// Then the response status should be 200
		ctx.thenTheResponseStatusShouldBe(200)
		// And the Allow header should contain "GET, POST, PUT, DELETE, HEAD, OPTIONS"
		ctx.thenTheAllowHeaderShouldContain("GET, POST, PUT, DELETE, HEAD, OPTIONS")
		// And the response should include LDP headers
		ctx.thenTheResponseShouldIncludeLDPHeaders()
	})

	t.Run("ContainerNotFound", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()

		// When I send GET request to "/containers/nonexistent"
		ctx.whenISendRequestToWithAccept("GET", "/containers/nonexistent", "")

		// Then the response status should be 404
		ctx.thenTheResponseStatusShouldBe(404)
		// And the response should contain "container not found" error
		ctx.thenTheResponseShouldContainError("container not found")
	})

	t.Run("InvalidContainerOperations", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")

		// When I send PATCH request to "/containers/documents"
		ctx.whenISendRequestToWithAccept("PATCH", "/containers/documents", "")

		// Then the response status should be 405
		ctx.thenTheResponseStatusShouldBe(405)
		// And the Allow header should list supported methods
		ctx.thenTheAllowHeaderShouldListSupportedMethods()
	})
}

// TestContainerPerformanceAndLargeCollections tests BDD scenarios for performance
func TestContainerPerformanceAndLargeCollections(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("LargeContainerPagination", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "large-collection" exists
		ctx.givenAContainerExists("large-collection")
		// And the container has 1000 resources
		ctx.givenTheContainerHasResources("large-collection", 1000)

		// When I request the first page with 50 items
		ctx.whenIRequestTheFirstPageWithItems(50)

		// Then the response should contain 50 items
		ctx.thenTheResponseShouldContainItems(50)
		// And the response should include pagination links
		ctx.thenTheResponseShouldIncludePaginationLinks()
		// And the response time should be under 1 second
		ctx.thenTheResponseTimeShouldBeUnderSecond(1)
	})

	t.Run("ContainerMemberFiltering", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "mixed-content" exists
		ctx.givenAContainerExists("mixed-content")
		// And the container has 100 RDF resources
		ctx.givenTheContainerHasResources("mixed-content", 100)
		// And the container has 100 binary files
		ctx.givenTheContainerHasResources("mixed-content", 100)

		// When I filter for RDF resources only
		ctx.whenIFilterForRDFResourcesOnly()

		// Then the response should contain only RDF resources
		ctx.thenTheResponseShouldContainOnlyRDFResources()
		// And the response should be returned quickly
		ctx.thenTheResponseShouldBeReturnedQuickly()
	})

	t.Run("ContainerMemberSorting", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "documents" exists
		ctx.givenAContainerExists("documents")
		// And the container has resources with different creation dates
		ctx.givenTheContainerHasResources("documents", 10)

		// When I sort members by creation date descending
		ctx.whenISortMembersByCreationDateDescending()

		// Then the members should be ordered by creation date
		ctx.thenTheMembersShouldBeOrderedByCreationDate()
		// And the newest resources should appear first
		ctx.thenTheNewestResourcesShouldAppearFirst()
	})

	t.Run("StreamingLargeContainerListings", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "huge-collection" exists
		ctx.givenAContainerExists("huge-collection")
		// And the container has 10000 resources
		ctx.givenTheContainerHasResources("huge-collection", 10000)

		// When I request the container listing
		ctx.whenIRequestTheContainerListing()

		// Then the response should be streamed
		ctx.thenTheResponseShouldBeStreamed()
		// And memory usage should remain constant
		ctx.thenMemoryUsageShouldRemainConstant()
		// And the response should start immediately
		ctx.thenTheResponseShouldStartImmediately()
	})

	t.Run("ConcurrentContainerAccess", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "shared" exists
		ctx.givenAContainerExists("shared")

		// When 10 clients simultaneously access the container
		ctx.whenClientsSimultaneouslyAccessTheContainer(10)

		// Then all requests should succeed
		ctx.thenAllRequestsShouldSucceed()
		// And the response times should be reasonable
		ctx.thenTheResponseTimesShouldBeReasonable()
		// And no race conditions should occur
		ctx.thenNoRaceConditionsShouldOccur()
	})

	t.Run("ContainerSizeCaching", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "cached" exists
		ctx.givenAContainerExists("cached")
		// And the container has 500 resources
		ctx.givenTheContainerHasResources("cached", 500)

		// When I request container metadata multiple times
		ctx.whenIRequestContainerMetadataMultipleTimes()

		// Then the size information should be cached
		ctx.thenTheSizeInformationShouldBeCached()
		// And subsequent requests should be faster
		ctx.thenSubsequentRequestsShouldBeFaster()
		// And the cache should be invalidated on updates
		ctx.thenTheCacheShouldBeInvalidatedOnUpdates()
	})

	t.Run("DeepHierarchyNavigationPerformance", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container hierarchy 10 levels deep
		// (This would need to be set up in the given step)

		// When I navigate to the deepest container
		ctx.whenINavigateToTheDeepestContainer()

		// Then the path resolution should be efficient
		ctx.thenThePathResolutionShouldBeEfficient()
		// And the response time should be acceptable
		ctx.thenTheResponseTimeShouldBeAcceptable()
		// And breadcrumb generation should be fast
		ctx.thenBreadcrumbGenerationShouldBeFast()
	})

	t.Run("BulkOperationsPerformance", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "bulk-test" exists
		ctx.givenAContainerExists("bulk-test")

		// When I add 100 resources to the container in batch
		ctx.whenIAddResourcesToTheContainerInBatch(100)

		// Then the operation should complete efficiently
		ctx.thenTheOperationShouldCompleteEfficiently()
		// And the membership index should be updated correctly
		ctx.thenTheMembershipIndexShouldBeUpdatedCorrectly()
		// And memory usage should remain reasonable
		ctx.thenMemoryUsageShouldRemainReasonable()
	})
}

// TestContainerConcurrencyAndRaceConditions tests BDD scenarios for concurrency
func TestContainerConcurrencyAndRaceConditions(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("ConcurrentContainerCreation", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()

		// When 5 clients simultaneously try to create container "shared"
		ctx.whenClientsSimultaneouslyTryToCreateContainer(5, "shared")

		// Then only one creation should succeed
		ctx.thenOnlyOneCreationShouldSucceed()
		// And the other attempts should fail gracefully
		ctx.thenTheOtherAttemptsShouldFailGracefully()
		// And no partial state should remain
		ctx.thenNoPartialStateShouldRemain()
	})

	t.Run("ConcurrentResourceAddition", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "concurrent" exists
		ctx.givenAContainerExists("concurrent")

		// When 10 clients simultaneously add different resources to the container
		ctx.whenClientsSimultaneouslyAddDifferentResources(10, "concurrent")

		// Then all resources should be added successfully
		ctx.thenAllResourcesShouldBeAddedSuccessfully()
		// And the membership index should be consistent
		ctx.thenTheMembershipIndexShouldBeConsistent()
		// And no resources should be lost
		ctx.thenNoResourcesShouldBeLost()
	})

	t.Run("ConcurrentMembershipUpdates", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "updates" exists
		ctx.givenAContainerExists("updates")
		// And the container has 10 resources
		ctx.givenTheContainerHasResources("updates", 10)

		// When multiple clients simultaneously add and remove resources
		ctx.whenMultipleClientsSimultaneouslyAddAndRemoveResources()

		// Then the final state should be consistent
		ctx.thenTheFinalStateShouldBeConsistent()
		// And the membership index should be accurate
		ctx.thenTheMembershipIndexShouldBeAccurate()
		// And no orphaned memberships should exist
		ctx.thenNoOrphanedMembershipsShouldExist()
	})

	t.Run("ContainerDeletionRaceCondition", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "race-delete" exists
		ctx.givenAContainerExists("race-delete")
		// And a resource "doc1.ttl" exists in the container
		ctx.givenAResourceExistsInContainer("doc1.ttl", "race-delete")

		// When one client deletes the resource while another deletes the container
		ctx.whenOneClientDeletesResourceWhileAnotherDeletesContainer()

		// Then the operations should be handled safely
		ctx.thenTheOperationsShouldBeHandledSafely()
		// And the final state should be consistent
		ctx.thenTheFinalStateShouldBeConsistent()
		// And no dangling references should remain
		ctx.thenNoDanglingReferencesShouldRemain()
	})

	t.Run("ConcurrentHierarchyModifications", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container hierarchy "root/level1/level2" exists
		ctx.givenAContainerExists("root")
		ctx.givenAContainerExistsInside("level1", "root")
		ctx.givenAContainerExistsInside("level2", "level1")

		// When multiple clients simultaneously modify different levels
		ctx.whenMultipleClientsSimultaneouslyModifyDifferentLevels()

		// Then all modifications should be applied correctly
		ctx.thenAllModificationsShouldBeAppliedCorrectly()
		// And the hierarchy should remain consistent
		ctx.thenTheHierarchyShouldRemainConsistent()
		// And no circular references should be created
		ctx.thenNoCircularReferencesShouldBeCreated()
	})

	t.Run("MembershipIndexConsistencyUnderLoad", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "load-test" exists
		ctx.givenAContainerExists("load-test")

		// When 20 clients perform random membership operations for 30 seconds
		ctx.whenClientsPerformRandomMembershipOperationsForSeconds(20, 30)

		// Then the membership index should remain consistent
		ctx.thenTheMembershipIndexShouldRemainConsistent()
		// And all operations should be atomic
		ctx.thenAllOperationsShouldBeAtomic()
		// And the index should match the actual container state
		ctx.thenTheIndexShouldMatchTheActualContainerState()
	})

	t.Run("ContainerMetadataRaceConditions", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "metadata-race" exists
		ctx.givenAContainerExists("metadata-race")

		// When multiple clients simultaneously update container metadata
		ctx.whenMultipleClientsSimultaneouslyUpdateContainerMetadata()

		// Then the updates should be serialized correctly
		ctx.thenTheUpdatesShouldBeSerializedCorrectly()
		// And the final metadata should be consistent
		ctx.thenTheFinalMetadataShouldBeConsistent()
		// And no updates should be lost
		ctx.thenNoUpdatesShouldBeLost()
	})

	t.Run("EventProcessingUnderConcurrency", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And a container "events" exists
		ctx.givenAContainerExists("events")

		// When multiple concurrent operations generate events
		ctx.whenMultipleConcurrentOperationsGenerateEvents()

		// Then all events should be processed correctly
		ctx.thenAllEventsShouldBeProcessedCorrectly()
		// And events should be in the correct order
		ctx.thenEventsShouldBeInTheCorrectOrder()
		// And no events should be lost or duplicated
		ctx.thenNoEventsShouldBeLostOrDuplicated()
	})
}

// TestContainerEventProcessing tests BDD scenarios for event processing
func TestContainerEventProcessing(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("ContainerCreationEvents", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()

		// When I create a container "events-test"
		ctx.whenICreateAContainerWithID("events-test")

		// Then a "container_created" event should be emitted
		ctx.thenAnEventShouldBeEmitted("container_created")
		// And the event should contain container ID
		ctx.thenTheEventShouldContainContainerID()
		// And the event should contain creation timestamp
		ctx.thenTheEventShouldContainCreationTimestamp()
		// And the event should contain container metadata
		ctx.thenTheEventShouldContainContainerMetadata()
	})

	t.Run("ContainerUpdateEvents", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And a container "update-test" exists
		ctx.givenAContainerExists("update-test")

		// When I update the container metadata
		ctx.whenIUpdateTheContainerMetadata()

		// Then a "container_updated" event should be emitted
		ctx.thenAnEventShouldBeEmitted("container_updated")
		// And the event should contain the changes
		ctx.thenTheEventShouldContainTheChanges()
		// And the event should contain update timestamp
		ctx.thenTheEventShouldContainUpdateTimestamp()
	})

	t.Run("ContainerDeletionEvents", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And an empty container "delete-test" exists
		ctx.givenAnEmptyContainerExists("delete-test")

		// When I delete the container
		ctx.whenIDeleteTheContainer()

		// Then a "container_deleted" event should be emitted
		ctx.thenAnEventShouldBeEmitted("container_deleted")
		// And the event should contain container ID
		ctx.thenTheEventShouldContainContainerID()
		// And the event should contain deletion timestamp
		ctx.thenTheEventShouldContainDeletionTimestamp()
	})

	t.Run("MemberAdditionEvents", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And a container "member-test" exists
		ctx.givenAContainerExists("member-test")
		// And a resource "doc1.ttl" exists
		ctx.givenAResourceExists("doc1.ttl")

		// When I add the resource to the container
		ctx.whenIAddResourceToContainer("doc1.ttl", "member-test")

		// Then a "member_added" event should be emitted
		ctx.thenAnEventShouldBeEmitted("member_added")
		// And the event should contain container ID
		ctx.thenTheEventShouldContainContainerID()
		// And the event should contain member ID
		ctx.thenTheEventShouldContainMemberID()
		// And the event should contain member type
		ctx.thenTheEventShouldContainMemberType()
	})

	t.Run("MemberRemovalEvents", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And a container "member-test" exists
		ctx.givenAContainerExists("member-test")
		// And a resource "doc1.ttl" exists in the container
		ctx.givenAResourceExistsInContainer("doc1.ttl", "member-test")

		// When I remove the resource from the container
		ctx.whenIRemoveResourceFromContainer("doc1.ttl", "member-test")

		// Then a "member_removed" event should be emitted
		ctx.thenAnEventShouldBeEmitted("member_removed")
		// And the event should contain container ID
		ctx.thenTheEventShouldContainContainerID()
		// And the event should contain member ID
		ctx.thenTheEventShouldContainMemberID()
	})

	t.Run("EventOrderingAndConsistency", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And a container "ordering-test" exists
		ctx.givenAContainerExists("ordering-test")

		// When I perform multiple operations in sequence
		ctx.whenIPerformMultipleOperationsInSequence()

		// Then events should be emitted in the correct order
		ctx.thenEventsShouldBeEmittedInTheCorrectOrder()
		// And each event should have a unique sequence number
		ctx.thenEachEventShouldHaveAUniqueSequenceNumber()
		// And events should be persisted reliably
		ctx.thenEventsShouldBePersistedReliably()
	})

	t.Run("EventHandlerProcessing", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And event handlers are registered
		ctx.givenEventHandlersAreRegistered()
		// And a container "handler-test" exists
		ctx.givenAContainerExists("handler-test")

		// When container operations occur
		ctx.whenContainerOperationsOccur()

		// Then event handlers should be invoked
		ctx.thenEventHandlersShouldBeInvoked()
		// And handlers should process events correctly
		ctx.thenHandlersShouldProcessEventsCorrectly()
		// And handler failures should not affect operations
		ctx.thenHandlerFailuresShouldNotAffectOperations()
	})

	t.Run("EventReplayCapability", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And a container "replay-test" exists
		ctx.givenAContainerExists("replay-test")
		// And several operations have been performed
		ctx.givenTheContainerHasResources("replay-test", 5)

		// When I replay events from a specific timestamp
		ctx.whenIReplayEventsFromASpecificTimestamp()

		// Then the container state should be reconstructed correctly
		ctx.thenTheContainerStateShouldBeReconstructedCorrectly()
		// And all events should be processed in order
		ctx.thenAllEventsShouldBeProcessedInOrder()
	})

	t.Run("EventFilteringAndQuerying", func(t *testing.T) {
		// Given a clean LDP server is running
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()
		// And event processing is enabled
		ctx.givenEventProcessingIsEnabled()
		// And multiple containers exist with various operations
		ctx.givenMultipleContainersExistWithVariousOperations()

		// When I query for events by container ID
		ctx.whenIQueryForEventsByContainerID()

		// Then only relevant events should be returned
		ctx.thenOnlyRelevantEventsShouldBeReturned()
		// And events should be properly filtered
		ctx.thenEventsShouldBeProperlyFiltered()
		// And query performance should be acceptable
		ctx.thenQueryPerformanceShouldBeAcceptable()
	})
}

// TestContainerEndToEndIntegration runs comprehensive end-to-end tests
func TestContainerEndToEndIntegration(t *testing.T) {
	ctx := NewContainerBDDContext(t)
	defer ctx.Cleanup()

	t.Run("CompleteContainerLifecycle", func(t *testing.T) {
		// Test complete container lifecycle from creation to deletion
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()

		// Create container
		ctx.whenICreateAContainerWithIDAndTitle("lifecycle-test", "Lifecycle Test Container")
		ctx.thenTheContainerShouldBeCreatedSuccessfully()
		ctx.thenTheContainerShouldHaveTitle("Lifecycle Test Container")

		// Add resources
		ctx.givenAResourceExists("resource1.ttl")
		ctx.givenAResourceExists("resource2.ttl")
		ctx.whenIAddResourceToContainer("resource1.ttl", "lifecycle-test")
		ctx.whenIAddResourceToContainer("resource2.ttl", "lifecycle-test")
		ctx.thenTheContainerShouldContainResource("lifecycle-test", "resource1.ttl")
		ctx.thenTheContainerShouldContainResource("lifecycle-test", "resource2.ttl")

		// Test HTTP API
		ctx.whenISendRequestToWithAccept("GET", "/containers/lifecycle-test", "text/turtle")
		ctx.thenTheResponseStatusShouldBe(200)
		ctx.thenTheResponseShouldContainLDPMembershipTriples()

		// Remove resources
		ctx.whenIRemoveResourceFromContainer("resource1.ttl", "lifecycle-test")
		ctx.whenIRemoveResourceFromContainer("resource2.ttl", "lifecycle-test")
		ctx.thenTheContainerShouldNotContainResource("lifecycle-test", "resource1.ttl")
		ctx.thenTheContainerShouldNotContainResource("lifecycle-test", "resource2.ttl")

		// Delete container
		ctx.whenISendRequestToWithAccept("DELETE", "/containers/lifecycle-test", "")
		ctx.thenTheResponseStatusShouldBe(204)
		ctx.thenTheContainerShouldBeDeleted()
	})

	t.Run("ContainerHierarchyWithMixedContent", func(t *testing.T) {
		// Test complex hierarchy with mixed content types
		ctx.givenACleanLDPServerIsRunning()
		ctx.givenTheServerSupportsContainerOperations()

		// Create hierarchy
		ctx.whenICreateAContainerWithID("root")
		ctx.whenICreateAContainerInside("documents", "root")
		ctx.whenICreateAContainerInside("images", "root")
		ctx.whenICreateAContainerInside("private", "documents")

		// Add mixed content
		ctx.givenAResourceExistsInContainer("doc1.ttl", "documents")
		ctx.givenAResourceExistsInContainer("doc2.ttl", "documents")
		ctx.givenAResourceExistsInContainer("secret.ttl", "private")
		ctx.givenAResourceExistsInContainer("photo1.jpg", "images")
		ctx.givenAResourceExistsInContainer("photo2.png", "images")

		// Test navigation and content negotiation
		ctx.whenISendRequestToWithAccept("GET", "/containers/root", "application/ld+json")
		ctx.thenTheResponseStatusShouldBe(200)
		ctx.thenTheResponseShouldBeValidJSONLD()

		ctx.whenISendRequestToWithAccept("GET", "/containers/documents", "text/turtle")
		ctx.thenTheResponseStatusShouldBe(200)
		ctx.thenTheResponseShouldContainLDPMembershipTriples()

		// Test performance with larger dataset
		ctx.givenTheContainerHasResources("documents", 100)
		ctx.whenISendRequestToWithAccept("GET", "/containers/documents", "")
		ctx.thenTheResponseStatusShouldBe(200)
		ctx.thenTheResponseTimeShouldBeAcceptable()
	})
}

// Helper function to run all container integration tests
func TestAllContainerIntegrationScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive integration tests in short mode")
	}

	t.Run("ContainerCreationAndHierarchy", TestContainerCreationAndHierarchyManagement)
	t.Run("ContainerMembership", TestContainerMembershipOperations)
	t.Run("ContainerHTTPAPI", TestContainerHTTPAPICompliance)
	t.Run("ContainerPerformance", TestContainerPerformanceAndLargeCollections)
	t.Run("ContainerConcurrency", TestContainerConcurrencyAndRaceConditions)
	t.Run("ContainerEvents", TestContainerEventProcessing)
	t.Run("ContainerEndToEnd", TestContainerEndToEndIntegration)
}
