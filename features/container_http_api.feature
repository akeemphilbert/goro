Feature: Container HTTP API Compliance
  As a developer
  I want LDP-compliant container endpoints
  So that my applications can interact with containers using standard HTTP methods

  Background:
    Given a clean LDP server is running
    And the server supports container operations

  Scenario: GET container with members
    Given a container "documents" exists
    And resources "doc1.ttl", "doc2.ttl" exist in container "documents"
    When I send GET request to "/containers/documents"
    Then the response status should be 200
    And the response should contain container metadata
    And the response should list all members
    And the Content-Type should be "text/turtle"

  Scenario: GET container with content negotiation
    Given a container "documents" exists
    And resources "doc1.ttl", "doc2.ttl" exist in container "documents"
    When I send GET request to "/containers/documents" with Accept "application/ld+json"
    Then the response status should be 200
    And the Content-Type should be "application/ld+json"
    And the response should be valid JSON-LD

  Scenario: POST to create resource in container
    Given a container "documents" exists
    When I send POST request to "/containers/documents" with RDF content
    Then the response status should be 201
    And the Location header should contain the new resource URI
    And the resource should be added to the container
    And the container should be updated

  Scenario: PUT to update container metadata
    Given a container "documents" exists
    When I send PUT request to "/containers/documents" with updated metadata
    Then the response status should be 200
    And the container metadata should be updated
    And the modification timestamp should be updated

  Scenario: DELETE empty container
    Given an empty container "temp" exists
    When I send DELETE request to "/containers/temp"
    Then the response status should be 204
    And the container should be deleted
    And the container should emit "container_deleted" event

  Scenario: DELETE non-empty container fails
    Given a container "documents" exists
    And a resource "doc1.ttl" exists in container "documents"
    When I send DELETE request to "/containers/documents"
    Then the response status should be 409
    And the response should contain "container not empty" error
    And the container should not be deleted

  Scenario: HEAD request for container metadata
    Given a container "documents" exists
    When I send HEAD request to "/containers/documents"
    Then the response status should be 200
    And the response should have no body
    And the headers should contain container metadata
    And the Content-Length header should be present

  Scenario: OPTIONS request for supported methods
    Given a container "documents" exists
    When I send OPTIONS request to "/containers/documents"
    Then the response status should be 200
    And the Allow header should contain "GET, POST, PUT, DELETE, HEAD, OPTIONS"
    And the response should include LDP headers

  Scenario: Container not found
    When I send GET request to "/containers/nonexistent"
    Then the response status should be 404
    And the response should contain "container not found" error

  Scenario: Invalid container operations
    Given a container "documents" exists
    When I send PATCH request to "/containers/documents"
    Then the response status should be 405
    And the Allow header should list supported methods