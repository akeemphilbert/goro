Feature: Data Integrity
  As a pod owner
  I want data integrity guarantees
  So that user data is never corrupted or lost

  Background:
    Given the storage system is running
    And the pod storage is available

  Scenario: Verify successful write operations
    Given I have valid RDF data
    When I store the resource
    Then the write operation should be verified as successful
    And the data should be immediately retrievable
    And the checksum should be calculated and stored

  Scenario: Detect and prevent serving corrupted data
    Given I have stored a resource successfully
    And the stored data becomes corrupted on disk
    When I try to retrieve the resource
    Then the system should detect the corruption
    And refuse to serve the corrupted data
    And return an appropriate error response

  Scenario: Maintain consistency during storage failures
    Given I am storing a large resource
    When the storage operation fails midway
    Then no partial data should be left behind
    And the system should remain in a consistent state
    And subsequent operations should work normally

  Scenario: Recover from storage errors gracefully
    Given the storage system encounters an error
    When the error condition is resolved
    Then the system should restore to a consistent state
    And all valid data should remain accessible
    And new operations should work normally

  Scenario: Validate data integrity on retrieval
    Given I have stored multiple resources
    When I retrieve each resource
    Then the checksum should be verified
    And any integrity violations should be detected
    And corrupted resources should be flagged

  Scenario: Atomic operations for data consistency
    Given I am updating an existing resource
    When the update operation is performed
    Then either the entire update succeeds
    Or the entire update fails with no changes
    And no intermediate state should be visible
    And concurrent reads should see consistent data

  Scenario: Event consistency for audit trail
    Given I perform multiple storage operations
    When each operation completes
    Then corresponding events should be recorded
    And the event log should be consistent with stored data
    And no events should be lost or duplicated