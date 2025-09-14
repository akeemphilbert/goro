Feature: Container Concurrency and Race Conditions
  As a system administrator
  I want containers to handle concurrent operations safely
  So that data integrity is maintained under load

  Background:
    Given a clean LDP server is running
    And the server supports container operations

  Scenario: Concurrent container creation
    When 5 clients simultaneously try to create container "shared"
    Then only one creation should succeed
    And the other attempts should fail gracefully
    And no partial state should remain

  Scenario: Concurrent resource addition
    Given a container "concurrent" exists
    When 10 clients simultaneously add different resources to the container
    Then all resources should be added successfully
    And the membership index should be consistent
    And no resources should be lost

  Scenario: Concurrent membership updates
    Given a container "updates" exists
    And the container has 10 resources
    When multiple clients simultaneously add and remove resources
    Then the final state should be consistent
    And the membership index should be accurate
    And no orphaned memberships should exist

  Scenario: Container deletion race condition
    Given a container "race-delete" exists
    And a resource "doc1.ttl" exists in the container
    When one client deletes the resource while another deletes the container
    Then the operations should be handled safely
    And the final state should be consistent
    And no dangling references should remain

  Scenario: Concurrent hierarchy modifications
    Given a container hierarchy "root/level1/level2" exists
    When multiple clients simultaneously modify different levels
    Then all modifications should be applied correctly
    And the hierarchy should remain consistent
    And no circular references should be created

  Scenario: Membership index consistency under load
    Given a container "load-test" exists
    When 20 clients perform random membership operations for 30 seconds
    Then the membership index should remain consistent
    And all operations should be atomic
    And the index should match the actual container state

  Scenario: Container metadata race conditions
    Given a container "metadata-race" exists
    When multiple clients simultaneously update container metadata
    Then the updates should be serialized correctly
    And the final metadata should be consistent
    And no updates should be lost

  Scenario: Event processing under concurrency
    Given a container "events" exists
    When multiple concurrent operations generate events
    Then all events should be processed correctly
    And events should be in the correct order
    And no events should be lost or duplicated