Feature: Container Membership Operations
  As a developer
  I want LDP-compliant containers
  So that my Solid applications work with standard protocols

  Background:
    Given a clean LDP server is running
    And the server supports container operations

  Scenario: Add resource to container
    Given a container "documents" exists
    And a resource "doc1.ttl" exists
    When I add resource "doc1.ttl" to container "documents"
    Then the container "documents" should contain resource "doc1.ttl"
    And the membership should be recorded in the index
    And the container should emit "member_added" event

  Scenario: Remove resource from container
    Given a container "documents" exists
    And a resource "doc1.ttl" exists in container "documents"
    When I remove resource "doc1.ttl" from container "documents"
    Then the container "documents" should not contain resource "doc1.ttl"
    And the membership should be removed from the index
    And the container should emit "member_removed" event

  Scenario: List container members
    Given a container "documents" exists
    And resources "doc1.ttl", "doc2.ttl", "doc3.ttl" exist in container "documents"
    When I list the members of container "documents"
    Then I should get 3 members
    And the members should include "doc1.ttl", "doc2.ttl", "doc3.ttl"
    And each member should have type information

  Scenario: Container with mixed content types
    Given a container "mixed" exists
    And a RDF resource "data.ttl" exists in container "mixed"
    And a binary file "image.jpg" exists in container "mixed"
    And a sub-container "subfolder" exists in container "mixed"
    When I list the members of container "mixed"
    Then I should get 3 members
    And the members should have correct type information
    And RDF resources should be marked as "Resource"
    And binary files should be marked as "NonRDFSource"
    And containers should be marked as "Container"

  Scenario: Automatic membership updates on resource creation
    Given a container "documents" exists
    When I create a resource "new-doc.ttl" in container "documents"
    Then the container "documents" should automatically contain "new-doc.ttl"
    And the membership index should be updated
    And the container modification timestamp should be updated

  Scenario: Automatic membership cleanup on resource deletion
    Given a container "documents" exists
    And a resource "temp-doc.ttl" exists in container "documents"
    When I delete resource "temp-doc.ttl"
    Then the container "documents" should not contain "temp-doc.ttl"
    And the membership index should be cleaned up
    And the container modification timestamp should be updated

  Scenario: LDP membership triples generation
    Given a container "documents" exists
    And resources "doc1.ttl", "doc2.ttl" exist in container "documents"
    When I retrieve container "documents" as Turtle
    Then the response should contain LDP membership triples
    And the triples should use "ldp:contains" predicate
    And each member should be properly referenced