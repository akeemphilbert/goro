Feature: Error Handling
  As a developer
  I want comprehensive error handling
  So that I can properly handle failures and provide meaningful feedback

  Background:
    Given the storage system is running

  Scenario: Handle resource not found
    Given no resource exists with ID "nonexistent-resource"
    When I try to retrieve the resource
    Then I should receive a 404 Not Found response
    And the error message should indicate resource not found

  Scenario: Handle unsupported format request
    Given I have stored RDF data in Turtle format
    When I request the resource with Accept header "application/unsupported"
    Then I should receive a 406 Not Acceptable response
    And the error message should list supported formats

  Scenario: Handle insufficient storage space
    Given the storage system is at capacity
    When I try to store a new resource
    Then I should receive a 507 Insufficient Storage response
    And the error message should indicate storage limitation

  Scenario: Handle data corruption detection
    Given I have a stored resource
    And the stored data becomes corrupted
    When I try to retrieve the resource
    Then I should receive a 500 Internal Server Error response
    And the error should indicate data corruption
    And the corrupted data should not be served

  Scenario: Handle format conversion failure
    Given I have stored invalid RDF data
    When I request format conversion to another RDF format
    Then I should receive a 422 Unprocessable Entity response
    And the error message should indicate conversion failure

  Scenario: Handle concurrent access conflicts
    Given I have a stored resource
    When multiple clients try to modify the resource simultaneously
    Then only one modification should succeed
    And other clients should receive appropriate conflict responses
    And data consistency should be maintained

  Scenario: Handle invalid resource data
    Given I have malformed RDF data
    When I try to store the resource
    Then I should receive a 400 Bad Request response
    And the error message should indicate validation failure
    And the invalid data should not be stored

  Scenario: Handle storage operation timeout
    Given the storage system is under heavy load
    When a storage operation takes too long
    Then I should receive a 408 Request Timeout response
    And the operation should be safely cancelled
    And no partial data should be stored