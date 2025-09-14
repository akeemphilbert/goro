Feature: Container Event Processing
  As a system integrator
  I want container operations to emit events
  So that I can build audit trails and integrations

  Background:
    Given a clean LDP server is running
    And the server supports container operations
    And event processing is enabled

  Scenario: Container creation events
    When I create a container "events-test"
    Then a "container_created" event should be emitted
    And the event should contain container ID
    And the event should contain creation timestamp
    And the event should contain container metadata

  Scenario: Container update events
    Given a container "update-test" exists
    When I update the container metadata
    Then a "container_updated" event should be emitted
    And the event should contain the changes
    And the event should contain update timestamp

  Scenario: Container deletion events
    Given an empty container "delete-test" exists
    When I delete the container
    Then a "container_deleted" event should be emitted
    And the event should contain container ID
    And the event should contain deletion timestamp

  Scenario: Member addition events
    Given a container "member-test" exists
    And a resource "doc1.ttl" exists
    When I add the resource to the container
    Then a "member_added" event should be emitted
    And the event should contain container ID
    And the event should contain member ID
    And the event should contain member type

  Scenario: Member removal events
    Given a container "member-test" exists
    And a resource "doc1.ttl" exists in the container
    When I remove the resource from the container
    Then a "member_removed" event should be emitted
    And the event should contain container ID
    And the event should contain member ID

  Scenario: Event ordering and consistency
    Given a container "ordering-test" exists
    When I perform multiple operations in sequence
    Then events should be emitted in the correct order
    And each event should have a unique sequence number
    And events should be persisted reliably

  Scenario: Event handler processing
    Given event handlers are registered
    And a container "handler-test" exists
    When container operations occur
    Then event handlers should be invoked
    And handlers should process events correctly
    And handler failures should not affect operations

  Scenario: Event replay capability
    Given a container "replay-test" exists
    And several operations have been performed
    When I replay events from a specific timestamp
    Then the container state should be reconstructed correctly
    And all events should be processed in order

  Scenario: Event filtering and querying
    Given multiple containers exist with various operations
    When I query for events by container ID
    Then only relevant events should be returned
    And events should be properly filtered
    And query performance should be acceptable