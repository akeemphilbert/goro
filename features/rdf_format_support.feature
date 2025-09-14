Feature: RDF Format Support
  As a pod owner
  I want to store RDF data in multiple formats
  So that applications can work with their preferred serialization

  Background:
    Given the storage system is running
    And the pod storage is available

  Scenario: Store and retrieve RDF data in Turtle format
    Given I have valid RDF data in Turtle format
    When I store the resource with content type "text/turtle"
    Then the resource should be stored successfully
    And I should be able to retrieve it in Turtle format
    And the semantic meaning should be preserved

  Scenario: Store and retrieve RDF data in JSON-LD format
    Given I have valid RDF data in JSON-LD format
    When I store the resource with content type "application/ld+json"
    Then the resource should be stored successfully
    And I should be able to retrieve it in JSON-LD format
    And the semantic meaning should be preserved

  Scenario: Store and retrieve RDF data in RDF/XML format
    Given I have valid RDF data in RDF/XML format
    When I store the resource with content type "application/rdf+xml"
    Then the resource should be stored successfully
    And I should be able to retrieve it in RDF/XML format
    And the semantic meaning should be preserved

  Scenario: Content negotiation for RDF formats
    Given I have stored RDF data in Turtle format
    When I request the resource with Accept header "application/ld+json"
    Then I should receive the data in JSON-LD format
    And the semantic meaning should be preserved

  Scenario: Convert between all supported RDF formats
    Given I have stored RDF data in JSON-LD format
    When I request the resource with Accept header "text/turtle"
    Then I should receive the data in Turtle format
    When I request the resource with Accept header "application/rdf+xml"
    Then I should receive the data in RDF/XML format
    And the semantic meaning should be preserved across all conversions

  Scenario: Reject unsupported RDF format
    Given I have valid RDF data
    When I request the resource with Accept header "application/n-triples"
    Then I should receive a 406 Not Acceptable response
    And the error message should indicate unsupported format